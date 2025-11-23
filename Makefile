.PHONY: help build run test test-e2e test-all lint clean docker-build docker-up docker-down migrate-up migrate-down logs stop restart load-test load-test-check load-test-health load-test-team load-test-user load-test-pr load-test-go load-test-go-health load-test-go-team load-test-go-user load-test-go-pr load-test-save load-test-plot

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
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

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

test-e2e: ## Запустить E2E тесты
	@echo "$(GREEN)Запуск E2E тестов...$(NC)"
	$(GO) test -v -timeout 10m ./tests/e2e/...

test-all: test test-e2e ## Запустить все тесты (unit + E2E)
	@echo "$(GREEN)Все тесты завершены!$(NC)"

# Нагрузочное тестирование
VEGETA := vegeta
LOAD_TEST_RATE := 5
LOAD_TEST_DURATION := 2m
LOAD_TEST_URL := http://localhost:8080

load-test-check: ## Проверить наличие vegeta
	@if ! command -v $(VEGETA) > /dev/null; then \
		echo "$(RED)vegeta не установлен. Установите: go install github.com/tsenart/vegeta/v12@latest$(NC)"; \
		echo "$(YELLOW)Или используйте: make load-test-go$(NC)"; \
		exit 1; \
	fi

load-test: ## Запустить все нагрузочные тесты через Go скрипт
	@echo "$(GREEN)Запуск нагрузочных тестов...$(NC)"
	@echo "$(YELLOW)Убедитесь, что приложение запущено на $(LOAD_TEST_URL)$(NC)"
	@cd tests/load && go run load.go all

load-test-health: ## Тест health endpoint (использует Go скрипт)
	@echo "$(GREEN)Тестирование health endpoint...$(NC)"
	@cd tests/load && go run load.go health

load-test-team: ## Тест team endpoints (использует Go скрипт)
	@echo "$(GREEN)Тестирование team endpoints...$(NC)"
	@cd tests/load && go run load.go team

load-test-user: ## Тест user endpoints (использует Go скрипт)
	@echo "$(GREEN)Тестирование user endpoints...$(NC)"
	@cd tests/load && go run load.go user

load-test-pr: ## Тест PR endpoints (использует Go скрипт)
	@echo "$(GREEN)Тестирование PR endpoints...$(NC)"
	@echo "$(YELLOW)Примечание: для тестов PR необходимо предварительно создать команду и пользователей$(NC)"
	@cd tests/load && go run load.go pr

load-test-go: ## Запустить нагрузочные тесты через Go скрипт
	@echo "$(GREEN)Запуск нагрузочных тестов через Go...$(NC)"
	@echo "$(YELLOW)Убедитесь, что приложение запущено на $(LOAD_TEST_URL)$(NC)"
	@cd tests/load && go run load.go all

load-test-go-health: ## Тест health через Go скрипт
	@cd tests/load && go run load.go health

load-test-go-team: ## Тест team через Go скрипт
	@cd tests/load && go run load.go team

load-test-go-user: ## Тест user через Go скрипт
	@cd tests/load && go run load.go user

load-test-go-pr: ## Тест PR через Go скрипт
	@cd tests/load && go run load.go pr

load-test-save: load-test-check ## Запустить тесты и сохранить результаты в файл
	@echo "$(GREEN)Запуск тестов с сохранением результатов...$(NC)"
	@echo "GET $(LOAD_TEST_URL)/health" | $(VEGETA) attack -rate=$(LOAD_TEST_RATE) -duration=$(LOAD_TEST_DURATION) > tests/load/results.bin
	@$(VEGETA) report tests/load/results.bin
	@echo "$(GREEN)Результаты сохранены в tests/load/results.bin$(NC)"
	@echo "$(YELLOW)Для визуализации: vegeta plot tests/load/results.bin > tests/load/plot.html$(NC)"

load-test-plot: load-test-check ## Создать график из сохраненных результатов
	@if [ ! -f tests/load/results.bin ]; then \
		echo "$(RED)Файл results.bin не найден. Сначала запустите: make load-test-save$(NC)"; \
		exit 1; \
	fi
	@$(VEGETA) plot tests/load/results.bin > tests/load/plot.html
	@echo "$(GREEN)График сохранен в tests/load/plot.html$(NC)"

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

