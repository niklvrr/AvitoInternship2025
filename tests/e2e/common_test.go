package e2e

import (
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

// ==================== ВАЛИДАЦИЯ СТРУКТУР ====================

// validateTeamMember проверяет структуру TeamMember
func validateTeamMember(t *testing.T, member map[string]interface{}) {
	t.Helper()
	require.Contains(t, member, "user_id", "TeamMember must have user_id")
	require.Contains(t, member, "username", "TeamMember must have username")
	require.Contains(t, member, "is_active", "TeamMember must have is_active")

	assert.IsType(t, "", member["user_id"], "user_id must be string")
	assert.IsType(t, "", member["username"], "username must be string")
	assert.IsType(t, false, member["is_active"], "is_active must be boolean")
}

// validateTeam проверяет структуру Team
func validateTeam(t *testing.T, team map[string]interface{}) {
	t.Helper()
	require.Contains(t, team, "team_name", "Team must have team_name")
	require.Contains(t, team, "members", "Team must have members")

	assert.IsType(t, "", team["team_name"], "team_name must be string")
	assert.IsType(t, []interface{}{}, team["members"], "members must be array")

	members := team["members"].([]interface{})
	for _, memberRaw := range members {
		member := memberRaw.(map[string]interface{})
		validateTeamMember(t, member)
	}
}

// validateUser проверяет структуру User
func validateUser(t *testing.T, user map[string]interface{}) {
	t.Helper()
	require.Contains(t, user, "user_id", "User must have user_id")
	require.Contains(t, user, "username", "User must have username")
	require.Contains(t, user, "team_name", "User must have team_name")
	require.Contains(t, user, "is_active", "User must have is_active")

	assert.IsType(t, "", user["user_id"], "user_id must be string")
	assert.IsType(t, "", user["username"], "username must be string")
	assert.IsType(t, "", user["team_name"], "team_name must be string")
	assert.IsType(t, false, user["is_active"], "is_active must be boolean")
}

// validatePullRequest проверяет структуру PullRequest
func validatePullRequest(t *testing.T, pr map[string]interface{}) {
	t.Helper()
	require.Contains(t, pr, "pull_request_id", "PullRequest must have pull_request_id")
	require.Contains(t, pr, "pull_request_name", "PullRequest must have pull_request_name")
	require.Contains(t, pr, "author_id", "PullRequest must have author_id")
	require.Contains(t, pr, "status", "PullRequest must have status")
	require.Contains(t, pr, "assigned_reviewers", "PullRequest must have assigned_reviewers")

	assert.IsType(t, "", pr["pull_request_id"], "pull_request_id must be string")
	assert.IsType(t, "", pr["pull_request_name"], "pull_request_name must be string")
	assert.IsType(t, "", pr["author_id"], "author_id must be string")
	assert.IsType(t, "", pr["status"], "status must be string")
	assert.IsType(t, []interface{}{}, pr["assigned_reviewers"], "assigned_reviewers must be array")

	// Проверяем enum для status
	status := pr["status"].(string)
	assert.Contains(t, []string{"OPEN", "MERGED"}, status, "status must be OPEN or MERGED")

	// Проверяем assigned_reviewers (0..2)
	reviewers := pr["assigned_reviewers"].([]interface{})
	assert.GreaterOrEqual(t, len(reviewers), 0, "assigned_reviewers count must be >= 0")
	assert.LessOrEqual(t, len(reviewers), 2, "assigned_reviewers count must be <= 2")

	// Проверяем, что все ревьюеры - строки
	for _, reviewer := range reviewers {
		assert.IsType(t, "", reviewer, "assigned_reviewers items must be strings")
	}

	// Проверяем опциональные поля createdAt и mergedAt
	if pr["createdAt"] != nil {
		createdAt, ok := pr["createdAt"].(string)
		require.True(t, ok, "createdAt must be string if present")
		_, err := time.Parse(time.RFC3339, createdAt)
		assert.NoError(t, err, "createdAt must be in RFC3339 format (date-time)")
	}

	if pr["mergedAt"] != nil {
		mergedAt, ok := pr["mergedAt"].(string)
		require.True(t, ok, "mergedAt must be string if present")
		_, err := time.Parse(time.RFC3339, mergedAt)
		assert.NoError(t, err, "mergedAt must be in RFC3339 format (date-time)")
	}
}

// validatePullRequestShort проверяет структуру PullRequestShort
func validatePullRequestShort(t *testing.T, pr map[string]interface{}) {
	t.Helper()
	require.Contains(t, pr, "pull_request_id", "PullRequestShort must have pull_request_id")
	require.Contains(t, pr, "pull_request_name", "PullRequestShort must have pull_request_name")
	require.Contains(t, pr, "author_id", "PullRequestShort must have author_id")
	require.Contains(t, pr, "status", "PullRequestShort must have status")

	assert.IsType(t, "", pr["pull_request_id"], "pull_request_id must be string")
	assert.IsType(t, "", pr["pull_request_name"], "pull_request_name must be string")
	assert.IsType(t, "", pr["author_id"], "author_id must be string")
	assert.IsType(t, "", pr["status"], "status must be string")

	status := pr["status"].(string)
	assert.Contains(t, []string{"OPEN", "MERGED"}, status, "status must be OPEN or MERGED")
}

// validateErrorResponse проверяет структуру ErrorResponse
func validateErrorResponse(t *testing.T, resp *http.Response, expectedCode string, expectedStatus int) {
	t.Helper()
	assert.Equal(t, expectedStatus, resp.StatusCode, "HTTP status code mismatch")

	var errorResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err, "Response must be valid JSON")

	require.Contains(t, errorResp, "error", "ErrorResponse must have error field")

	errorObj := errorResp["error"].(map[string]interface{})
	require.Contains(t, errorObj, "code", "Error must have code")
	require.Contains(t, errorObj, "message", "Error must have message")

	assert.Equal(t, expectedCode, errorObj["code"], "Error code mismatch")
	assert.IsType(t, "", errorObj["code"], "code must be string")
	assert.IsType(t, "", errorObj["message"], "message must be string")

	// Проверяем, что код ошибки из допустимого enum
	validCodes := []string{"TEAM_EXISTS", "PR_EXISTS", "PR_MERGED", "NOT_ASSIGNED", "NO_CANDIDATE", "NOT_FOUND"}
	assert.Contains(t, validCodes, errorObj["code"], "Error code must be from enum")
}

// TestHealthCheck проверяет health check эндпоинт
func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(testServer.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check must return 200")

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Response must be valid JSON")
	assert.Equal(t, "ok", result["status"], "Status must be 'ok'")
}
