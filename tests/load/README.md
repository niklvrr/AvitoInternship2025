# Нагрузочное тестирование

Этот каталог содержит скрипты и конфигурации для нагрузочного тестирования API с использованием [Vegeta](https://github.com/tsenart/vegeta).

## Требования

### Установка Vegeta CLI

```bash
# macOS
brew install vegeta

# Linux
wget https://github.com/tsenart/vegeta/releases/download/v12.13.0/vegeta-v12.13.0-linux-amd64.tar.gz
tar -xzf vegeta-v12.13.0-linux-amd64.tar.gz
sudo mv vegeta /usr/local/bin/

# Или через Go
go install github.com/tsenart/vegeta/v12@latest
```

## Использование

### Способ 1: Через Makefile (рекомендуется)

Все команды используют Go скрипт для надежной работы с POST запросами:

```bash
# Запустить все нагрузочные тесты
make load-test

# Запустить тест конкретного эндпоинта
make load-test-health
make load-test-team
make load-test-user
make load-test-pr

# Альтернативные команды (то же самое)
make load-test-go-health
make load-test-go-team
make load-test-go-user
make load-test-go-pr
```

### Способ 2: Через Go скрипт

```bash
# Запустить все сценарии
go run tests/load/load.go all

# Запустить конкретный сценарий
go run tests/load/load.go health
go run tests/load/load.go team
go run tests/load/load.go user
go run tests/load/load.go pr
```

### Способ 3: Через Vegeta CLI напрямую

```bash
# Тест health endpoint (5 RPS в течение 2 минут)
echo "GET http://localhost:8080/health" | vegeta attack -rate=5 -duration=2m | vegeta report

# Тест с сохранением результатов
echo "GET http://localhost:8080/health" | vegeta attack -rate=5 -duration=2m > results.bin
vegeta report results.bin
vegeta plot results.bin > plot.html
```

## Параметры тестирования

По умолчанию используются следующие параметры (соответствуют требованиям из TASK.md):

- **RPS**: 5 запросов в секунду
- **Длительность**: 2 минуты
- **SLI время ответа**: P95 < 300ms
- **SLI успешности**: > 99.9%

## Файлы

- `*.targets` - файлы с описанием HTTP запросов для Vegeta CLI
- `load.go` - Go скрипт для комплексного тестирования с динамической генерацией данных
- `README.md` - эта документация

## Интерпретация результатов

### Ключевые метрики

1. **Success Rate** - процент успешных запросов (должен быть > 99.9%)
2. **P95 Latency** - 95-й перцентиль времени ответа (должен быть < 300ms)
3. **P99 Latency** - 99-й перцентиль времени ответа
4. **Requests/sec** - фактическая пропускная способность

### Пример успешного результата

```
Requests Total:     600
Success Rate:       100.00%
Duration:           2m0s

Latency:
  Mean:             45ms
  P50:              42ms
  P95:              280ms  ✓ (target: < 300ms)
  P99:              295ms
  Max:              298ms

SLI Compliance:
  P95 Latency:      280.00 ms (target: < 300ms) - PASS
  Success Rate:     100.00% (target: > 99.9%) - PASS
```

## Перед запуском тестов

1. **Обязательно**: Убедитесь, что приложение запущено и доступно на `http://localhost:8080`
   ```bash
   # Запустить приложение
   make docker-up
   # или
   make run
   ```

2. Для тестов PR необходимо предварительно создать команду и пользователей (Go скрипт делает это автоматически)

3. Тесты используют динамическую генерацию данных, поэтому конфликты маловероятны

## Примечания

- Тесты используют случайные идентификаторы для избежания конфликтов
- Для тестов PR необходимо предварительно создать команду и пользователей
- Health endpoint можно тестировать независимо от состояния БД

