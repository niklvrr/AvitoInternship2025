package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUserService мок сервиса для тестов
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SetIsActive(ctx context.Context, req *request.SetIsActiveRequest) (*response.SetIsActiveResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.SetIsActiveResponse), args.Error(1)
}

func (m *MockUserService) GetReview(ctx context.Context, req *request.GetReviewRequest) (*response.GetReviewResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.GetReviewResponse), args.Error(1)
}

func TestUserHandler_SetIsActive_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService, logger)

	reqBody := request.SetIsActiveRequest{
		UserId:   "user1",
		IsActive: true,
	}

	expectedResp := &response.SetIsActiveResponse{
		UserId:   "user1",
		Username: "Test User",
		TeamName: "team1",
		IsActive: true,
	}

	mockService.On("SetIsActive", mock.Anything, mock.MatchedBy(func(r *request.SetIsActiveRequest) bool {
		return r.UserId == "user1" && r.IsActive == true
	})).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetIsActive(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "user")
	mockService.AssertExpectations(t)
}

func TestUserHandler_SetIsActive_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService, logger)

	reqBody := request.SetIsActiveRequest{
		UserId:   "",
		IsActive: true,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockService.On("SetIsActive", mock.Anything, &reqBody).Return(nil, service.ErrUserNotFound)

	handler.SetIsActive(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_SetIsActive_UserNotFound(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService, logger)

	reqBody := request.SetIsActiveRequest{
		UserId:   "user1",
		IsActive: true,
	}

	mockService.On("SetIsActive", mock.Anything, mock.Anything).Return(nil, service.WrapError(service.ErrUserNotFound, nil))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetIsActive(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "error")
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetReview_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService, logger)

	expectedPrs := []*domain.Pr{
		{
			Id:        "pr1",
			Name:      "PR 1",
			AuthorId:  "author1",
			Status:    "OPEN",
			CreatedAt: time.Now(),
		},
	}

	expectedResp := &response.GetReviewResponse{
		UserId: "user1",
		Prs:    expectedPrs,
	}

	mockService.On("GetReview", mock.Anything, mock.MatchedBy(func(r *request.GetReviewRequest) bool {
		return r.UserId == "user1"
	})).Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=user1", nil)
	w := httptest.NewRecorder()

	handler.GetReview(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "user1", result["user_id"])
	assert.Contains(t, result, "pull_requests")
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetReview_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
	w := httptest.NewRecorder()

	reqBody := request.GetReviewRequest{UserId: ""}
	mockService.On("GetReview", mock.Anything, &reqBody).Return(nil, service.ErrUserNotFound)

	handler.GetReview(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestUserHandler_GetReview_UserNotFound(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService, logger)

	mockService.On("GetReview", mock.Anything, mock.Anything).Return(nil, service.WrapError(service.ErrUserNotFound, nil))

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=user1", nil)
	w := httptest.NewRecorder()

	handler.GetReview(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "error")
	mockService.AssertExpectations(t)
}
