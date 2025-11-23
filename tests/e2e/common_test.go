package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/db"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/repository"
	"github.com/niklvrr/AvitoInternship2025/internal/transport"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/handler"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase/service"
	"github.com/niklvrr/AvitoInternship2025/pkg/logger"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testServer *http.Server
var testDB *postgres.PostgresContainer
var baseURL = "http://localhost:8081"

func TestMain(m *testing.M) {
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start postgres container: %v", err))
	}
	testDB = postgresContainer

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %v", err))
	}

	log, err := logger.NewLogger("dev")
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	// Изменяем рабочую директорию на корень проекта для правильного пути к миграциям
	wd, err := os.Getwd()
	if err == nil {
		// Если мы в tests/e2e, переходим в корень проекта
		if filepath.Base(wd) == "e2e" && filepath.Base(filepath.Dir(wd)) == "tests" {
			projectRoot := filepath.Join(wd, "..", "..")
			if err := os.Chdir(projectRoot); err == nil {
				defer os.Chdir(wd)
			}
		}
	}

	database, err := db.NewDatabase(ctx, connStr, log)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}
	defer database.Close()

	userRepo := repository.NewUserRepository(database, log)
	teamRepo := repository.NewTeamRepository(database, log)
	prRepo := repository.NewPrRepository(database, log)

	userService := service.NewUserService(userRepo, log)
	teamService := service.NewTeamService(teamRepo, log)
	prService := service.NewPrService(prRepo, log)

	userHandler := handler.NewUserHandler(userService, log)
	teamHandler := handler.NewTeamHandler(teamService, log)
	prHandler := handler.NewPrHandler(prService, log)
	statsHandler := handler.NewStatsHandler(prService, log)
	healthHandler := handler.NewHealthHandler(log)

	router := transport.NewRouter(
		userHandler,
		teamHandler,
		prHandler,
		statsHandler,
		healthHandler,
		log,
	)

	testServer = &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	go func() {
		if err := testServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("failed to start test server: %v", err))
		}
	}()

	time.Sleep(500 * time.Millisecond)

	code := m.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	testServer.Shutdown(ctx)
	testDB.Terminate(ctx)

	os.Exit(code)
}

func makeRequest(t *testing.T, method, url string, body interface{}) *http.Response {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func parseJSONResponse(t *testing.T, resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err)
}

func parseErrorResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	defer resp.Body.Close()
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}
