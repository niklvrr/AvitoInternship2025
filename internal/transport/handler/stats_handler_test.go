package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockStatsService мок сервиса для тестов
type MockStatsService struct {
	mock.Mock
}

func (m *MockStatsService) GetStats(ctx context.Context) (*response.StatsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.StatsResponse), args.Error(1)
}

func TestStatsHandler_GetStats_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockStatsService)
	handler := NewStatsHandler(mockService, logger)

	expectedResp := &response.StatsResponse{
		Users: []response.UserStat{
			{UserId: "u1", Username: "Alice", Assignments: 5},
			{UserId: "u2", Username: "Bob", Assignments: 3},
		},
		PRs: []response.PrStat{
			{PrId: "pr1", PrName: "PR 1", ReviewersCount: 2},
			{PrId: "pr2", PrName: "PR 2", ReviewersCount: 1},
		},
	}

	mockService.On("GetStats", mock.Anything).Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result response.StatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result.Users, 2)
	assert.Len(t, result.PRs, 2)
	assert.Equal(t, "u1", result.Users[0].UserId)
	assert.Equal(t, "Alice", result.Users[0].Username)
	assert.Equal(t, 5, result.Users[0].Assignments)
	assert.Equal(t, "pr1", result.PRs[0].PrId)
	assert.Equal(t, "PR 1", result.PRs[0].PrName)
	assert.Equal(t, 2, result.PRs[0].ReviewersCount)
	mockService.AssertExpectations(t)
}

func TestStatsHandler_GetStats_EmptyStats(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockStatsService)
	handler := NewStatsHandler(mockService, logger)

	expectedResp := &response.StatsResponse{
		Users: []response.UserStat{},
		PRs:   []response.PrStat{},
	}

	mockService.On("GetStats", mock.Anything).Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result response.StatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Len(t, result.Users, 0)
	assert.Len(t, result.PRs, 0)
	mockService.AssertExpectations(t)
}

func TestStatsHandler_GetStats_ServiceError(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockStatsService)
	handler := NewStatsHandler(mockService, logger)

	mockService.On("GetStats", mock.Anything).Return(nil, service.ErrPrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "error")
	mockService.AssertExpectations(t)
}

