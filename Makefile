.PHONY: help build run test test-coverage test-e2e lint fmt clean docker-up docker-down migrate-up migrate-down

# Переменные
BINARY_NAME := server
MAIN_PATH := ./cmd/main.go
GO := go
DB_URL := postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable
MIGRATIONS_PATH := migrations

# Цвета для вывода
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

help: ## Показать справку по командам
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

build: ## Собрать бинарный файл
	@echo "$(GREEN)Сборка проекта...$(NC)"
	$(GO) build -o $(BINARY_NAME) $(MAIN_PATH)

run: build ## Собрать и запустить приложение
	@echo "$(GREEN)Запуск приложения...$(NC)"
	./$(BINARY_NAME)

test: ## Запустить unit тесты
	@echo "$(GREEN)Запуск unit тестов...$(NC)"
	$(GO) test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(NC)"
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии сохранен в coverage.html$(NC)"

test-e2e: ## Запустить E2E тесты
	@echo "$(GREEN)Запуск E2E тестов...$(NC)"
	$(GO) test -v -timeout 10m ./tests/e2e/...

lint: ## Запустить линтер
	@echo "$(GREEN)Запуск линтера...$(NC)"
	@if ! command -v golangci-lint > /dev/null; then \
		echo "golangci-lint не установлен. Установите: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	golangci-lint run

fmt: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	$(GO) fmt ./...

clean: ## Удалить скомпилированные файлы
	@echo "$(YELLOW)Очистка...$(NC)"
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Очистка завершена$(NC)"

docker-up: ## Запустить контейнеры
	@echo "$(GREEN)Запуск контейнеров...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Контейнеры запущены$(NC)"

docker-down: ## Остановить контейнеры
	@echo "$(YELLOW)Остановка контейнеров...$(NC)"
	docker-compose down

migrate-up: ## Применить миграции
	@echo "$(GREEN)Применение миграций...$(NC)"
	@if ! command -v migrate > /dev/null; then \
		echo "$(YELLOW)migrate не установлен. Установите: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest$(NC)"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down: ## Откатить миграции
	@echo "$(YELLOW)Откат миграций...$(NC)"
	@if ! command -v migrate > /dev/null; then \
		echo "$(YELLOW)migrate не установлен. Установите: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest$(NC)"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down

.DEFAULT_GOAL := help
