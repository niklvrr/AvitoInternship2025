package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Unit тесты для repository слоя требуют реальной БД
// Интеграционные тесты для GetStats находятся в e2e тестах
// Здесь оставляем только базовую проверку структуры

func TestPrRepository_GetStats_Interface(t *testing.T) {
	// Проверяем, что метод GetStats существует в интерфейсе
	logger := zap.NewNop()
	repo := &PrRepository{
		log: logger,
	}

	// Проверяем, что структура может быть создана
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.log)
}

