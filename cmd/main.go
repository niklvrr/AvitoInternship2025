package main

import (
	"context"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/repository"
	"github.com/niklvrr/AvitoInternship2025/internal/transport"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/handler"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase/service"
	"time"

	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/db"

	"github.com/niklvrr/AvitoInternship2025/internal/config"
	"github.com/niklvrr/AvitoInternship2025/pkg/logger"

	"go.uber.org/zap"

	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger, err := logger.NewLogger(cfg.App.Env)
	if err != nil {
		log.Fatal(err)
	}
	logger.Debug("Logger init success")

	db, err := db.NewDatabase(ctx, cfg.Database.URL, logger)
	if err != nil {
		logger.Fatal("Database init error", zap.Error(err))
	}
	defer db.Close()
	logger.Debug("Database init success")

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db, logger)
	teamRepo := repository.NewTeamRepository(db, logger)
	prRepo := repository.NewPrRepository(db, logger)

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

	httpServer := transport.NewServer(cfg.App.Port, router, logger)

	// Запуск сервера в горутине
	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("Server started", zap.String("port", cfg.App.Port))

	// Ожидание сигнала завершения
	<-ctx.Done()
	logger.Info("Shutdown signal received")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	} else {
		logger.Info("Server stopped")
	}
}
