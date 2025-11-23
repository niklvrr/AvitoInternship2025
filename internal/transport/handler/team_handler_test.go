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

// MockTeamService мок сервиса для тестов
type MockTeamService struct {
	mock.Mock
}

func (m *MockTeamService) Add(ctx context.Context, req *request.AddTeamRequest) (*response.AddTeamResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AddTeamResponse), args.Error(1)
}

func (m *MockTeamService) Get(ctx context.Context, req *request.GetTeamRequest) (*response.GetTeamResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.GetTeamResponse), args.Error(1)
}

func TestTeamHandler_AddTeam_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService, logger)

	members := []*domain.User{
		{
			Id:        "user1",
			Name:      "User 1",
			TeamName:  "team1",
			IsActive:  true,
			CreatedAt: time.Now(),
		},
	}

	reqBody := request.AddTeamRequest{
		TeamName: "team1",
		Members:  members,
	}

	expectedResp := &response.AddTeamResponse{
		TeamName: "team1",
		Members:  members,
	}

	mockService.On("Add", mock.Anything, mock.MatchedBy(func(r *request.AddTeamRequest) bool {
		return r.TeamName == "team1" && len(r.Members) == 1
	})).Return(expectedResp, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddTeam(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "team")
	mockService.AssertExpectations(t)
}

func TestTeamHandler_AddTeam_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService, logger)

	reqBody := request.AddTeamRequest{
		TeamName: "",
		Members:  []*domain.User{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockService.On("Add", mock.Anything, &reqBody).Return(nil, service.ErrTeamNotFound)

	handler.AddTeam(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestTeamHandler_AddTeam_TeamExists(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService, logger)

	reqBody := request.AddTeamRequest{
		TeamName: "team1",
		Members:  []*domain.User{},
	}

	mockService.On("Add", mock.Anything, mock.Anything).Return(nil, service.WrapError(service.ErrTeamExists, nil))

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.AddTeam(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "error")
	mockService.AssertExpectations(t)
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService, logger)

	members := []*domain.User{
		{
			Id:        "user1",
			Name:      "User 1",
			TeamName:  "team1",
			IsActive:  true,
			CreatedAt: time.Now(),
		},
	}

	expectedResp := &response.GetTeamResponse{
		TeamName: "team1",
		Members:  members,
	}

	mockService.On("Get", mock.Anything, mock.MatchedBy(func(r *request.GetTeamRequest) bool {
		return r.TeamName == "team1"
	})).Return(expectedResp, nil)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=team1", nil)
	w := httptest.NewRecorder()

	handler.GetTeam(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "team1", result["team_name"])
	assert.Contains(t, result, "members")
	mockService.AssertExpectations(t)
}

func TestTeamHandler_GetTeam_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService, logger)

	req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
	w := httptest.NewRecorder()

	reqBody := request.GetTeamRequest{TeamName: ""}
	mockService.On("Get", mock.Anything, &reqBody).Return(nil, service.ErrTeamNotFound)

	handler.GetTeam(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestTeamHandler_GetTeam_TeamNotFound(t *testing.T) {
	logger := zap.NewNop()
	mockService := new(MockTeamService)
	handler := NewTeamHandler(mockService, logger)

	mockService.On("Get", mock.Anything, mock.Anything).Return(nil, service.WrapError(service.ErrTeamNotFound, nil))

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.GetTeam(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Contains(t, result, "error")
	mockService.AssertExpectations(t)
}
