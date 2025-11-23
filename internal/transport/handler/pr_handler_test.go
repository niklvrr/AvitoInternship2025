package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"github.com/niklvrr/AvitoInternship2025/internal/usecase/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockPrService мок сервиса для тестов
type MockPrService struct {
	mock.Mock
}

func (m *MockPrService) Create(ctx context.Context, req *request.CreateRequest) (*response.CreateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CreateResponse), args.Error(1)
}

func (m *MockPrService) Merge(ctx context.Context, req *request.MergeRequest) (*response.MergeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MergeResponse), args.Error(1)
}

func (m *MockPrService) Reassign(ctx context.Context, req *request.ReassignRequest) (*response.ReassignResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ReassignResponse), args.Error(1)
}

func TestPrHandler_CreatePr_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPrService)
	handler := NewPrHandler(mockService, logger)

	reqBody := request.CreateRequest{
		PrId:     "pr1",
		PrName:   "Test PR",
		AuthorId: "author1",
	}

	expectedResp := &response.CreateResponse{
		PrId:              "pr1",
		PrName:            "Test PR",
		AuthorId:          "author1",
		Status:            "OPEN",
		AssignedReviewers: []string{"reviewer1", "reviewer2"},
		CreatedAt:         time.Now().Format(time.RFC3339),
		MergedAt:          nil,
	}

	mockService.On("Create", mock.Anything, mock.MatchedBy(func(r *request.CreateRequest) bool {
		return r.PrId == "pr1" && r.PrName == "Test PR" && r.AuthorId == "author1"
	})).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePr(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "pr")
	mockService.AssertExpectations(t)
}

func TestPrHandler_CreatePr_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPrService)
	handler := NewPrHandler(mockService, logger)

	reqBody := request.CreateRequest{
		PrId:     "",
		PrName:   "Test PR",
		AuthorId: "author1",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockService.On("Create", mock.Anything, &reqBody).Return(nil, service.ErrPrNotFound)

	handler.CreatePr(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestPrHandler_CreatePr_PrExists(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPrService)
	handler := NewPrHandler(mockService, logger)

	reqBody := request.CreateRequest{
		PrId:     "pr1",
		PrName:   "Test PR",
		AuthorId: "author1",
	}

	mockService.On("Create", mock.Anything, mock.Anything).Return(nil, service.WrapError(service.ErrPrExists, nil))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePr(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "error")
	mockService.AssertExpectations(t)
}

func TestPrHandler_MergePr_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPrService)
	handler := NewPrHandler(mockService, logger)

	reqBody := request.MergeRequest{
		PrId: "pr1",
	}

	mergedAt := time.Now().Format(time.RFC3339)
	expectedResp := &response.MergeResponse{
		PrId:              "pr1",
		PrName:            "Test PR",
		AuthorId:          "author1",
		Status:            "MERGED",
		AssignedReviewers: []string{"reviewer1"},
		CreatedAt:         time.Now().Format(time.RFC3339),
		MergedAt:          &mergedAt,
	}

	mockService.On("Merge", mock.Anything, mock.MatchedBy(func(r *request.MergeRequest) bool {
		return r.PrId == "pr1"
	})).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.MergePr(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "pr")
	mockService.AssertExpectations(t)
}

func TestPrHandler_MergePr_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPrService)
	handler := NewPrHandler(mockService, logger)

	reqBody := request.MergeRequest{
		PrId: "",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockService.On("Merge", mock.Anything, &reqBody).Return(nil, service.ErrPrNotFound)

	handler.MergePr(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestPrHandler_ReassignPr_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPrService)
	handler := NewPrHandler(mockService, logger)

	reqBody := request.ReassignRequest{
		PrId:      "pr1",
		OldUserId: "old_reviewer",
	}

	expectedResp := &response.ReassignResponse{
		PrId:              "pr1",
		PrName:            "Test PR",
		AuthorId:          "author1",
		Status:            "OPEN",
		AssignedReviewers: []string{"new_reviewer"},
		ReplacedBy:        "new_reviewer",
		CreatedAt:         time.Now().Format(time.RFC3339),
		MergedAt:          nil,
	}

	mockService.On("Reassign", mock.Anything, mock.MatchedBy(func(r *request.ReassignRequest) bool {
		return r.PrId == "pr1" && r.OldUserId == "old_reviewer"
	})).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ReassignPr(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "pr")
	assert.Contains(t, result, "replaced_by")
	mockService.AssertExpectations(t)
}

func TestPrHandler_ReassignPr_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPrService)
	handler := NewPrHandler(mockService, logger)

	reqBody := request.ReassignRequest{
		PrId:      "",
		OldUserId: "old_reviewer",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockService.On("Reassign", mock.Anything, &reqBody).Return(nil, service.ErrPrNotFound)

	handler.ReassignPr(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestPrHandler_ReassignPr_PrMerged(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockPrService)
	handler := NewPrHandler(mockService, logger)

	reqBody := request.ReassignRequest{
		PrId:      "pr1",
		OldUserId: "old_reviewer",
	}

	mockService.On("Reassign", mock.Anything, mock.Anything).Return(nil, service.WrapError(service.ErrPrMerged, nil))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ReassignPr(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "error")
	mockService.AssertExpectations(t)
}
