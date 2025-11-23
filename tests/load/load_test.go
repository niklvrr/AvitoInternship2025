//go:build load
// +build load

package load

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	baseURL        = "http://localhost:8080"
	targetRPS      = 5
	duration       = 30 * time.Second
	maxLatencyP99  = 300 * time.Millisecond
	minSuccessRate = 0.999 // 99.9%
	// Допустимое отклонение RPS от целевого значения: ±10%
	rpsTolerance = 0.1
)

// Структура для хранения метрик нагрузочного тестирования
type metrics struct {
	totalRequests   int
	successRequests int
	errorRequests   int
	latencies       []time.Duration
}

// Тест нагрузочного тестирования создания Pull Request
func TestLoad_CreatePR(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск нагрузочного теста в коротком режиме")
	}

	// Проверка доступности сервера и подготовка тестовых данных
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	healthResp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Сервер не запущен по адресу %s. Пожалуйста, запустите сервер командой: make run\nОшибка: %v", baseURL, err)
	}
	healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("Проверка здоровья сервера не прошла со статусом %d", healthResp.StatusCode)
	}

	// Подготовка тестовых данных: создание команды и пользователей при необходимости
	setupTestData(t, client)

	loadClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	metrics := &metrics{
		latencies: make([]time.Duration, 0),
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Интервал между запросами для достижения целевого RPS
	interval := time.Second / time.Duration(targetRPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			goto done
		case <-ticker.C:
			reqStart := time.Now()

			reqBody := map[string]string{
				"pull_request_id":   fmt.Sprintf("pr-load-%d", time.Now().UnixNano()),
				"pull_request_name": "Load Test PR",
				"author_id":         "u1",
			}

			body, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", baseURL+"/pullRequest/create", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := loadClient.Do(req)
			latency := time.Since(reqStart)
			metrics.latencies = append(metrics.latencies, latency)
			metrics.totalRequests++

			if err != nil {
				metrics.errorRequests++
				if metrics.errorRequests <= 3 {
					t.Logf("Ошибка запроса: %v", err)
				}
				continue
			}

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				metrics.successRequests++
			} else {
				metrics.errorRequests++
				if metrics.errorRequests <= 3 {
					body, _ := io.ReadAll(resp.Body)
					t.Logf("Запрос не удался: status=%d, body=%s", resp.StatusCode, string(body))
					resp.Body.Close()
				} else {
					resp.Body.Close()
				}
				continue
			}
			resp.Body.Close()
		}
	}

done:
	elapsed := time.Since(start)
	printMetrics(t, "CreatePR", metrics, elapsed)
	validateMetrics(t, metrics, elapsed)
}

// Тест нагрузочного тестирования слияния Pull Request
func TestLoad_MergePR(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск нагрузочного теста в коротком режиме")
	}

	// Проверка доступности сервера и подготовка тестовых данных
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	healthResp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Сервер не запущен по адресу %s. Пожалуйста, запустите сервер командой: make run\nОшибка: %v", baseURL, err)
	}
	healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("Проверка здоровья сервера не прошла со статусом %d", healthResp.StatusCode)
	}

	// Подготовка тестовых данных: создание команды и пользователей при необходимости
	setupTestData(t, client)

	// Подготовка: создание PR перед началом теста
	prID := setupPR(t)

	loadClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	metrics := &metrics{
		latencies: make([]time.Duration, 0),
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Интервал между запросами для достижения целевого RPS
	interval := time.Second / time.Duration(targetRPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			goto done
		case <-ticker.C:
			reqStart := time.Now()

			reqBody := map[string]string{
				"pull_request_id": prID,
			}

			body, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", baseURL+"/pullRequest/merge", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := loadClient.Do(req)
			latency := time.Since(reqStart)
			metrics.latencies = append(metrics.latencies, latency)
			metrics.totalRequests++

			if err != nil {
				metrics.errorRequests++
				if metrics.errorRequests <= 3 {
					t.Logf("Ошибка запроса TestLoad_MergePR: %v", err)
				}
				continue
			}

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				metrics.successRequests++
				resp.Body.Close()
			} else {
				metrics.errorRequests++
				// Чтение и логирование тела ответа для не-2xx ответов
				if resp.Body != nil {
					bodyBytes, readErr := io.ReadAll(resp.Body)
					if readErr != nil {
						t.Errorf("TestLoad_MergePR: не удалось прочитать тело ответа с ошибкой: %v", readErr)
					} else {
						t.Logf("TestLoad_MergePR: ответ не-2xx (status %d): %s", resp.StatusCode, string(bodyBytes))
					}
					resp.Body.Close()
				}
			}
		}
	}

done:
	elapsed := time.Since(start)
	printMetrics(t, "MergePR", metrics, elapsed)
	validateMetrics(t, metrics, elapsed)
}

// Тест нагрузочного тестирования получения статистики
func TestLoad_GetStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск нагрузочного теста в коротком режиме")
	}

	// Проверка доступности сервера
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	healthResp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Сервер не запущен по адресу %s. Пожалуйста, запустите сервер командой: make run\nОшибка: %v", baseURL, err)
	}
	healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("Проверка здоровья сервера не прошла со статусом %d", healthResp.StatusCode)
	}

	// Проверка существования эндпоинта статистики
	statsResp, err := client.Get(baseURL + "/stats")
	if err != nil {
		t.Fatalf("Не удалось достичь эндпоинта статистики: %v", err)
	}
	statsResp.Body.Close()
	if statsResp.StatusCode == http.StatusNotFound {
		t.Fatalf("Эндпоинт статистики /stats не найден (404). Проверьте, что эндпоинт зарегистрирован.")
	}

	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	metrics := &metrics{
		latencies: make([]time.Duration, 0),
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	// Интервал между запросами для достижения целевого RPS
	interval := time.Second / time.Duration(targetRPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			goto done
		case <-ticker.C:
			reqStart := time.Now()

			req, _ := http.NewRequest("GET", baseURL+"/stats", nil)

			resp, err := client.Do(req)
			latency := time.Since(reqStart)
			metrics.latencies = append(metrics.latencies, latency)
			metrics.totalRequests++

			if err != nil {
				metrics.errorRequests++
				if metrics.totalRequests <= 3 {
					t.Logf("Ошибка запроса: %v", err)
				}
				continue
			}

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				metrics.successRequests++
			} else {
				metrics.errorRequests++
				if metrics.errorRequests <= 3 {
					body, _ := io.ReadAll(resp.Body)
					t.Logf("Запрос не удался: status=%d, body=%s", resp.StatusCode, string(body))
					resp.Body.Close()
				} else {
					resp.Body.Close()
				}
				continue
			}
			resp.Body.Close()
		}
	}

done:
	elapsed := time.Since(start)
	printMetrics(t, "GetStatistics", metrics, elapsed)
	validateMetrics(t, metrics, elapsed)
}

// Подготовка тестовых данных: создание команды "backend" с пользователями при необходимости
func setupTestData(t *testing.T, client *http.Client) {
	teamBody := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{
				"user_id":   "u1",
				"username":  "Alice",
				"is_active": true,
			},
			{
				"user_id":   "u2",
				"username":  "Bob",
				"is_active": true,
			},
			{
				"user_id":   "u3",
				"username":  "Charlie",
				"is_active": true,
			},
		},
	}

	body, _ := json.Marshal(teamBody)
	req, _ := http.NewRequest("POST", baseURL+"/team/add", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Предупреждение: не удалось подготовить тестовые данные: %v", err)
		return
	}
	resp.Body.Close()
	// Игнорируем ошибку, если команда уже существует (409) или успешно создана (201) - это нормально
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		t.Logf("Предупреждение: неожиданный статус при создании команды: %d", resp.StatusCode)
	}
}

// Подготовка PR перед началом нагрузочного теста
func setupPR(t *testing.T) string {
	client := &http.Client{Timeout: 5 * time.Second}

	prID := fmt.Sprintf("pr-load-setup-%d", time.Now().UnixNano())
	reqBody := map[string]string{
		"pull_request_id":   prID,
		"pull_request_name": "Setup PR",
		"author_id":         "u1",
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/pullRequest/create", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверка статуса ответа перед возвратом ID PR
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)
		if readErr != nil {
			bodyStr = fmt.Sprintf("не удалось прочитать тело ответа: %v", readErr)
		}
		t.Fatalf("setupPR не удался: ожидался статус 2xx, получен %d. Тело ответа: %s", resp.StatusCode, bodyStr)
	}

	return prID
}

// Вывод метрик нагрузочного тестирования
func printMetrics(t *testing.T, testName string, m *metrics, elapsed time.Duration) {
	if len(m.latencies) == 0 {
		return
	}

	// Вычисление перцентилей
	sorted := make([]time.Duration, len(m.latencies))
	copy(sorted, m.latencies)
	sortDurations(sorted)

	p50 := sorted[len(sorted)*50/100]
	p95 := sorted[len(sorted)*95/100]
	p99 := sorted[len(sorted)*99/100]
	p999 := sorted[len(sorted)*999/1000]

	avgLatency := time.Duration(0)
	for _, lat := range m.latencies {
		avgLatency += lat
	}
	avgLatency /= time.Duration(len(m.latencies))

	successRate := float64(m.successRequests) / float64(m.totalRequests)
	actualRPS := float64(m.totalRequests) / elapsed.Seconds()

	t.Logf("\n=== Результаты нагрузочного теста: %s ===", testName)
	t.Logf("Длительность: %v", elapsed)
	t.Logf("Всего запросов: %d", m.totalRequests)
	t.Logf("Успешных запросов: %d", m.successRequests)
	t.Logf("Запросов с ошибками: %d", m.errorRequests)
	t.Logf("Процент успешности: %.4f%%", successRate*100)
	t.Logf("Фактический RPS: %.2f", actualRPS)
	t.Logf("Средняя задержка: %v", avgLatency)
	t.Logf("P50 задержка: %v", p50)
	t.Logf("P95 задержка: %v", p95)
	t.Logf("P99 задержка: %v", p99)
	t.Logf("P99.9 задержка: %v", p999)
}

// Валидация метрик нагрузочного тестирования согласно требованиям SLI
func validateMetrics(t *testing.T, m *metrics, elapsed time.Duration) {
	if len(m.latencies) == 0 {
		return
	}

	successRate := float64(m.successRequests) / float64(m.totalRequests)

	sorted := make([]time.Duration, len(m.latencies))
	copy(sorted, m.latencies)
	sortDurations(sorted)
	p99 := sorted[len(sorted)*99/100]

	// Вычисление фактического RPS
	actualRPS := float64(m.totalRequests) / elapsed.Seconds()
	minRPS := float64(targetRPS) * (1 - rpsTolerance)
	maxRPS := float64(targetRPS) * (1 + rpsTolerance)

	require.GreaterOrEqual(t, successRate, minSuccessRate,
		"Процент успешности %.4f%% ниже требуемого %.4f%%", successRate*100, minSuccessRate*100)

	require.LessOrEqual(t, p99, maxLatencyP99,
		"P99 задержка %v превышает максимальную %v", p99, maxLatencyP99)

	require.GreaterOrEqual(t, actualRPS, minRPS,
		"Фактический RPS %.2f ниже минимального %.2f (целевой: %.2f)", actualRPS, minRPS, float64(targetRPS))

	require.LessOrEqual(t, actualRPS, maxRPS,
		"Фактический RPS %.2f превышает максимальный %.2f (целевой: %.2f)", actualRPS, maxRPS, float64(targetRPS))
}

// Сортировка массива задержек по возрастанию
func sortDurations(durations []time.Duration) {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
}
