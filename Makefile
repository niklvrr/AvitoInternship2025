.PHONY: help build run test lint clean docker-build docker-up docker-down migrate-up migrate-down logs stop restart

# Переменные
APP_NAME := avito-internship
BINARY_NAME := server
MAIN_PATH := ./cmd/main.go
DOCKER_COMPOSE := docker-compose
GO := go
GOLANGCI_LINT := golangci-lint

# Цвета для вывода
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Показать справку по командам
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Сборка и запуск
build: ## Собрать бинарный файл
	@echo "$(GREEN)Сборка проекта...$(NC)"
	$(GO) build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Сборка завершена: $(BINARY_NAME)$(NC)"

run: build ## Собрать и запустить приложение локально
	@echo "$(GREEN)Запуск приложения...$(NC)"
	./$(BINARY_NAME)

# Docker команды
docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	$(DOCKER_COMPOSE) build

docker-up: ## Запустить контейнеры через docker-compose
	@echo "$(GREEN)Запуск контейнеров...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Контейнеры запущены$(NC)"
	@echo "$(YELLOW)Приложение доступно на http://localhost:8080$(NC)"

docker-down: ## Остановить и удалить контейнеры
	@echo "$(YELLOW)Остановка контейнеров...$(NC)"
	$(DOCKER_COMPOSE) down

docker-up-build: ## Собрать и запустить контейнеры
	@echo "$(GREEN)Сборка и запуск контейнеров...$(NC)"
	$(DOCKER_COMPOSE) up -d --build

docker-logs: ## Показать логи контейнеров
	$(DOCKER_COMPOSE) logs -f

docker-logs-app: ## Показать логи приложения
	$(DOCKER_COMPOSE) logs -f app

docker-logs-db: ## Показать логи базы данных
	$(DOCKER_COMPOSE) logs -f db

docker-stop: ## Остановить контейнеры без удаления
	$(DOCKER_COMPOSE) stop

docker-restart: ## Перезапустить контейнеры
	$(DOCKER_COMPOSE) restart

docker-ps: ## Показать статус контейнеров
	$(DOCKER_COMPOSE) ps

# Тестирование
test: ## Запустить тесты
	@echo "$(GREEN)Запуск тестов...$(NC)"
	$(GO) test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(NC)"
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет о покрытии сохранен в coverage.html$(NC)"

test-short: ## Запустить только быстрые тесты
	$(GO) test -short ./...

# Линтинг и форматирование
lint: ## Запустить линтер
	@echo "$(GREEN)Запуск линтера...$(NC)"
	@if ! command -v $(GOLANGCI_LINT) > /dev/null; then \
		echo "$(RED)golangci-lint не установлен. Установите: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run

lint-fix: ## Запустить линтер с автоисправлением
	@echo "$(GREEN)Запуск линтера с автоисправлением...$(NC)"
	@if ! command -v $(GOLANGCI_LINT) > /dev/null; then \
		echo "$(RED)golangci-lint не установлен. Установите: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run --fix

fmt: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	$(GO) fmt ./...

vet: ## Запустить go vet
	@echo "$(GREEN)Запуск go vet...$(NC)"
	$(GO) vet ./...

# Миграции
migrate-up: ## Применить миграции (требует запущенной БД)
	@echo "$(GREEN)Применение миграций...$(NC)"
	@if ! command -v migrate > /dev/null; then \
		echo "$(RED)migrate не установлен. Установите: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest$(NC)"; \
		exit 1; \
	fi
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable" up

migrate-down: ## Откатить миграции (требует запущенной БД)
	@echo "$(YELLOW)Откат миграций...$(NC)"
	@if ! command -v migrate > /dev/null; then \
		echo "$(RED)migrate не установлен. Установите: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest$(NC)"; \
		exit 1; \
	fi
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable" down

migrate-status: ## Показать статус миграций
	@if ! command -v migrate > /dev/null; then \
		echo "$(RED)migrate не установлен. Установите: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest$(NC)"; \
		exit 1; \
	fi
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable" version

# Зависимости
deps: ## Загрузить зависимости
	@echo "$(GREEN)Загрузка зависимостей...$(NC)"
	$(GO) mod download
	$(GO) mod tidy

deps-update: ## Обновить зависимости
	@echo "$(GREEN)Обновление зависимостей...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy

deps-vendor: ## Создать vendor директорию
	$(GO) mod vendor

# Очистка
clean: ## Удалить скомпилированные файлы
	@echo "$(YELLOW)Очистка...$(NC)"
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Очистка завершена$(NC)"

clean-docker: ## Остановить и удалить контейнеры, volumes и образы
	@echo "$(YELLOW)Очистка Docker...$(NC)"
	$(DOCKER_COMPOSE) down -v --rmi all
	@echo "$(GREEN)Docker очищен$(NC)"

clean-all: clean clean-docker ## Полная очистка (бинарники + Docker)

# Разработка
dev: docker-up ## Запустить окружение для разработки
	@echo "$(GREEN)Окружение для разработки запущено$(NC)"
	@echo "$(YELLOW)Приложение: http://localhost:8080$(NC)"
	@echo "$(YELLOW)База данных: localhost:5432$(NC)"

dev-stop: docker-down ## Остановить окружение разработки

# Проверки перед коммитом
check: fmt vet lint test ## Запустить все проверки (fmt, vet, lint, test)
	@echo "$(GREEN)Все проверки пройдены!$(NC)"

# Установка инструментов разработки
install-tools: ## Установить инструменты разработки
	@echo "$(GREEN)Установка инструментов разработки...$(NC)"
	@if ! command -v $(GOLANGCI_LINT) > /dev/null; then \
		echo "Установка golangci-lint..."; \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v migrate > /dev/null; then \
		echo "Установка migrate..."; \
		$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi
	@echo "$(GREEN)Инструменты установлены$(NC)"

# По умолчанию показываем справку
.DEFAULT_GOAL := help

