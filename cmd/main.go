package cmd

import (
	"context"
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

	// TODO config init
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// TODO logger init
	logger, err := logger.NewLogger(cfg.App.Env)
	if err != nil {
		log.Fatal(err)
	}

	// TODO db init
	db, err := db.NewDatabase(ctx, cfg.Database.URL, logger)
	if err != nil {
		logger.Fatal("Database init error", zap.Error(err))
	}

	// TODO layers init

	// TODO server init
}
