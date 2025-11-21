package db

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.uber.org/zap"
)

var (
	errDBPathIsEmpty = errors.New("database path is empty")
	errDBInit        = errors.New("database init error")
)

func NewDatabase(ctx context.Context, dbUrl string, logger *zap.Logger) (*pgxpool.Pool, error) {
	if dbUrl == "" {
		return nil, errDBPathIsEmpty
	}

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, errDBInit
	}

	mustRunMigrations(dbUrl, logger)

	return pool, nil
}

func mustRunMigrations(dbUrl string, logger *zap.Logger) {
	if dbUrl == "" {
		logger.Fatal("dbUrl is empty")
	}

	mg, err := migrate.New(
		"file://migrations",
		dbUrl,
	)
	if err != nil {
		logger.Fatal("migration init err", zap.Error(err))
	}

	version, dirty, err := mg.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		logger.Fatal("migration version check err", zap.Error(err))
	}

	if dirty {
		logger.Warn("database is in dirty state, forcing version", zap.Uint("version", version))
		if err := mg.Force(int(version)); err != nil {
			logger.Fatal("failed to force migration version", zap.Error(err))
		}
		logger.Debug("dirty state cleared, retrying migration")
	}

	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Fatal("migration run err", zap.Error(err))
	}

	logger.Debug("migration run ok")
}
