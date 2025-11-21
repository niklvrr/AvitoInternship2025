package config

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

var (
	dbUserEmptyError = errors.New("DB User is Empty")
	dbNameEmptyError = errors.New("DB Name is Empty")
	envLoadError     = errors.New(".env load Error")
)

type AppConfig struct {
	Env  string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	Password string
	User     string
	URL      string
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, envLoadError
	}

	c := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "dev"),
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", "localhost"),
			Port:     getEnv("DATABASE_PORT", "5432"),
			Name:     getEnv("DATABASE_NAME", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", "postgres"),
			User:     getEnv("DATABASE_USER", "postgres"),
		},
	}
	err := makeDbUrl(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func makeDbUrl(cfg *Config) error {
	if cfg.Database.URL == "" {
		if cfg.Database.User == "" {
			return dbUserEmptyError
		}
		if cfg.Database.Name == "" {
			return dbNameEmptyError
		}
		cfg.Database.URL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
	}
	return nil
}
