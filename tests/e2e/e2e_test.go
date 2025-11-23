package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/repository"
	"github.com/niklvrr/AvitoInternship2025/internal/transport"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/handler"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var (
	testServer *httptest.Server
	testDB     *postgres.PostgresContainer
	dbURL      string
)

// runMigrations применяет миграции к тестовой БД
func runMigrations(dbURL string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Если мы в tests/e2e, переходим на два уровня выше
	var migrationsPath string
	if filepath.Base(wd) == "e2e" {
		projectRoot := filepath.Join(wd, "..", "..")
		migrationsPath = filepath.Join(projectRoot, "migrations")
	} else {
		migrationsPath = filepath.Join(wd, "migrations")
	}

	// Применяем миграции напрямую
	mg, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("migration init err: %w", err)
	}

	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration run err: %w", err)
	}

	return nil
}

// setupTestServer создает тестовый HTTP сервер
func setupTestServer(dbURL string) (*httptest.Server, error) {
	logger := zap.NewNop()

	// Применяем миграции
	if err := runMigrations(dbURL); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	database := pool

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(database, logger)
	teamRepo := repository.NewTeamRepository(database, logger)
	prRepo := repository.NewPrRepository(database, logger)

	// Инициализация сервисов
	userService := service.NewUserService(userRepo, logger)
	teamService := service.NewTeamService(teamRepo, logger)
	prService := service.NewPrService(prRepo, logger)

	// Инициализация хэндлеров
	userHandler := handler.NewUserHandler(userService, logger)
	teamHandler := handler.NewTeamHandler(teamService, logger)
	prHandler := handler.NewPrHandler(prService, logger)
	healthHandler := handler.NewHealthHandler(logger)

	// Инициализация роутера
	router := transport.NewRouter(
		userHandler,
		teamHandler,
		prHandler,
		healthHandler,
		logger,
	)

	return httptest.NewServer(router), nil
}

// TestMain настраивает тестовое окружение
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Создаем тестовую БД
	var err error
	testDB, err = postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start test container: %v", err))
	}

	dbURL, err = testDB.ConnectionString(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %v", err))
	}
	// Парсим URL и добавляем sslmode=disable
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse connection string: %v", err))
	}
	query := parsedURL.Query()
	query.Set("sslmode", "disable")
	parsedURL.RawQuery = query.Encode()
	dbURL = parsedURL.String()

	// Создаем тестовый сервер (миграции применятся автоматически в NewDatabase)
	testServer, err = setupTestServer(dbURL)
	if err != nil {
		panic(fmt.Sprintf("failed to setup test server: %v", err))
	}

	// Запускаем тесты
	code := m.Run()

	// Очистка
	if testServer != nil {
		testServer.Close()
	}
	if testDB != nil {
		if err := testDB.Terminate(ctx); err != nil {
			panic(fmt.Sprintf("failed to terminate container: %v", err))
		}
	}

	os.Exit(code)
}

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}

func TestE2E_TeamFlow(t *testing.T) {
	// 1. Создание команды
	teamReq := map[string]interface{}{
		"team_name": "backend_teamflow",
		"members": []map[string]interface{}{
			{
				"user_id":   "u1_teamflow",
				"username":  "Alice",
				"is_active": true,
			},
			{
				"user_id":   "u2_teamflow",
				"username":  "Bob",
				"is_active": true,
			},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var teamResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&teamResp)
	require.NoError(t, err)
	assert.Contains(t, teamResp, "team")

	// Небольшая задержка для завершения транзакции
	time.Sleep(100 * time.Millisecond)

	// 2. Получение команды
	resp, err = http.Get(testServer.URL + "/team/get?team_name=backend_teamflow")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var getTeamResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&getTeamResp)
	require.NoError(t, err)
	assert.Equal(t, "backend_teamflow", getTeamResp["team_name"])
	assert.Contains(t, getTeamResp, "members")
}

func TestE2E_UserFlow(t *testing.T) {
	// Сначала создаем команду с пользователем
	teamReq := map[string]interface{}{
		"team_name": "frontend_userflow",
		"members": []map[string]interface{}{
			{
				"user_id":   "u3_userflow",
				"username":  "Charlie",
				"is_active": true,
			},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Проверяем ответ создания команды
	var teamResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&teamResp)
	require.NoError(t, err)
	resp.Body.Close()

	// Убеждаемся, что команда создана
	assert.Contains(t, teamResp, "team")

	// Задержка для завершения транзакции и индексации
	time.Sleep(500 * time.Millisecond)

	// 1. Деактивация пользователя
	setActiveReq := map[string]interface{}{
		"user_id":   "u3_userflow",
		"is_active": false,
	}

	body, _ = json.Marshal(setActiveReq)
	resp, err = http.Post(testServer.URL+"/users/setIsActive", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK for setIsActive")

	var userResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	require.NoError(t, err)
	assert.Contains(t, userResp, "user")

	user, ok := userResp["user"].(map[string]interface{})
	require.True(t, ok, "user should be a map")
	assert.Equal(t, "u3_userflow", user["user_id"])
	assert.Equal(t, false, user["is_active"])

	// 2. Получение ревьюев пользователя (пока пустой список)
	resp, err = http.Get(testServer.URL + "/users/getReview?user_id=u3_userflow")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reviewResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&reviewResp)
	require.NoError(t, err)
	assert.Equal(t, "u3_userflow", reviewResp["user_id"])
	assert.Contains(t, reviewResp, "pull_requests")
}

func TestE2E_PRFlow(t *testing.T) {
	// Создаем команду с несколькими пользователями
	teamReq := map[string]interface{}{
		"team_name": "devops_prflow",
		"members": []map[string]interface{}{
			{
				"user_id":   "u4_prflow",
				"username":  "David",
				"is_active": true,
			},
			{
				"user_id":   "u5_prflow",
				"username":  "Eve",
				"is_active": true,
			},
			{
				"user_id":   "u6_prflow",
				"username":  "Frank",
				"is_active": true,
			},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Проверяем ответ создания команды
	var teamResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&teamResp)
	require.NoError(t, err)
	resp.Body.Close()

	// Убеждаемся, что команда создана
	assert.Contains(t, teamResp, "team")

	// Задержка для завершения транзакции и индексации
	time.Sleep(500 * time.Millisecond)

	// 1. Создание PR
	createPRReq := map[string]interface{}{
		"pull_request_id":   "pr-1001_prflow",
		"pull_request_name": "Add feature",
		"author_id":         "u4_prflow",
	}

	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var prResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&prResp)
	require.NoError(t, err)
	assert.Contains(t, prResp, "pr")

	pr := prResp["pr"].(map[string]interface{})
	assert.Equal(t, "pr-1001_prflow", pr["pull_request_id"])
	assert.Equal(t, "OPEN", pr["status"])
	assert.Contains(t, pr, "assigned_reviewers")
	reviewers := pr["assigned_reviewers"].([]interface{})
	assert.GreaterOrEqual(t, len(reviewers), 0)
	assert.LessOrEqual(t, len(reviewers), 2)

	// 2. Merge PR
	mergePRReq := map[string]interface{}{
		"pull_request_id": "pr-1001_prflow",
	}

	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var mergeResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&mergeResp)
	require.NoError(t, err)
	assert.Contains(t, mergeResp, "pr")

	mergedPR := mergeResp["pr"].(map[string]interface{})
	assert.Equal(t, "MERGED", mergedPR["status"])
	assert.NotNil(t, mergedPR["mergedAt"])

	// 3. Попытка переназначения после merge (должна вернуть ошибку)
	if len(reviewers) > 0 {
		reassignReq := map[string]interface{}{
			"pull_request_id": "pr-1001_prflow",
			"old_user_id":     reviewers[0].(string),
		}

		body, _ = json.Marshal(reassignReq)
		resp, err = http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var errorResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Contains(t, errorResp, "error")
	}

}

func TestE2E_PRReassignFlow(t *testing.T) {
	// Создаем команду
	teamReq := map[string]interface{}{
		"team_name": "qa_reassign",
		"members": []map[string]interface{}{
			{
				"user_id":   "u7_reassign",
				"username":  "Grace",
				"is_active": true,
			},
			{
				"user_id":   "u8_reassign",
				"username":  "Henry",
				"is_active": true,
			},
			{
				"user_id":   "u9_reassign",
				"username":  "Ivy",
				"is_active": true,
			},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Проверяем ответ создания команды
	var teamResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&teamResp)
	require.NoError(t, err)
	resp.Body.Close()

	// Убеждаемся, что команда создана
	assert.Contains(t, teamResp, "team")

	// Задержка для завершения транзакции и индексации
	time.Sleep(500 * time.Millisecond)

	// Создаем PR
	createPRReq := map[string]interface{}{
		"pull_request_id":   "pr-2001_reassign",
		"pull_request_name": "Fix bug",
		"author_id":         "u7_reassign",
	}

	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()

	var prResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&prResp)
	require.NoError(t, err)
	pr := prResp["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})

	if len(reviewers) > 0 {
		// Переназначение ревьюера
		reassignReq := map[string]interface{}{
			"pull_request_id": "pr-2001_reassign",
			"old_user_id":     reviewers[0].(string),
		}

		body, _ = json.Marshal(reassignReq)
		resp, err = http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var reassignResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&reassignResp)
		require.NoError(t, err)
		assert.Contains(t, reassignResp, "pr")
		assert.Contains(t, reassignResp, "replaced_by")
	}
}

func TestE2E_GetReviewFlow(t *testing.T) {
	// Создаем команду
	teamReq := map[string]interface{}{
		"team_name": "security_review",
		"members": []map[string]interface{}{
			{
				"user_id":   "u10_review",
				"username":  "Jack",
				"is_active": true,
			},
			{
				"user_id":   "u11_review",
				"username":  "Kate",
				"is_active": true,
			},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Проверяем ответ создания команды
	var teamResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&teamResp)
	require.NoError(t, err)
	resp.Body.Close()

	// Убеждаемся, что команда создана
	assert.Contains(t, teamResp, "team")

	// Задержка для завершения транзакции и индексации
	time.Sleep(500 * time.Millisecond)

	// Создаем PR от u10_review
	createPRReq := map[string]interface{}{
		"pull_request_id":   "pr-3001_review",
		"pull_request_name": "Security update",
		"author_id":         "u10_review",
	}

	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Небольшая задержка для завершения транзакции
	time.Sleep(100 * time.Millisecond)

	// Получаем ревьюи для u11_review (если он был назначен)
	resp, err = http.Get(testServer.URL + "/users/getReview?user_id=u11_review")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var reviewResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&reviewResp)
	require.NoError(t, err)
	assert.Equal(t, "u11_review", reviewResp["user_id"])
	assert.Contains(t, reviewResp, "pull_requests")
}

func TestE2E_ErrorCases(t *testing.T) {
	// Используем уникальное имя с timestamp для избежания конфликтов
	uniqueName := fmt.Sprintf("duplicate_error_%d", time.Now().UnixNano())

	// 1. Создание команды с существующим именем
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members":   []map[string]interface{}{},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	time.Sleep(200 * time.Millisecond)

	// Пытаемся создать еще раз
	body, _ = json.Marshal(teamReq)
	resp, err = http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// 2. Получение несуществующей команды
	nonexistentName := fmt.Sprintf("nonexistent_error_%d", time.Now().UnixNano())
	resp, err = http.Get(testServer.URL + "/team/get?team_name=" + nonexistentName)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// 3. Создание PR с несуществующим автором
	createPRReq := map[string]interface{}{
		"pull_request_id":   "pr-9999_error",
		"pull_request_name": "Test",
		"author_id":         "nonexistent_user_error",
	}

	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// 4. Merge несуществующего PR
	mergePRReq := map[string]interface{}{
		"pull_request_id": "pr-nonexistent_error",
	}

	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestE2E_IdempotentMerge(t *testing.T) {
	// Создаем команду и PR (нужно минимум 2 пользователя для ревьюеров)
	teamReq := map[string]interface{}{
		"team_name": "idempotent_merge_final",
		"members": []map[string]interface{}{
			{
				"user_id":   "u12_idempotent_final",
				"username":  "Liam",
				"is_active": true,
			},
			{
				"user_id":   "u13_idempotent_final",
				"username":  "Mia",
				"is_active": true,
			},
		},
	}

	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Проверяем ответ создания команды
	var teamResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&teamResp)
	require.NoError(t, err)
	resp.Body.Close()

	// Убеждаемся, что команда создана
	assert.Contains(t, teamResp, "team")

	// Задержка для завершения транзакции и индексации
	time.Sleep(500 * time.Millisecond)

	createPRReq := map[string]interface{}{
		"pull_request_id":   "pr-4001_idempotent_final",
		"pull_request_name": "Idempotent test",
		"author_id":         "u12_idempotent_final",
	}

	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Небольшая задержка для завершения транзакции
	time.Sleep(500 * time.Millisecond)

	// Первый merge
	mergePRReq := map[string]interface{}{
		"pull_request_id": "pr-4001_idempotent_final",
	}

	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Небольшая задержка для завершения транзакции
	time.Sleep(500 * time.Millisecond)

	// Второй merge (должен быть идемпотентным)
	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var mergeResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&mergeResp)
	require.NoError(t, err)
	mergedPR := mergeResp["pr"].(map[string]interface{})
	assert.Equal(t, "MERGED", mergedPR["status"])
}
