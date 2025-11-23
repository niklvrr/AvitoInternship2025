package service

import (
	"context"
	"testing"
	"time"

	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/repository"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockTeamRepository мок репозитория для тестов
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Add(ctx context.Context, dto *dto.AddTeamDTO) (*result.AddTeamResult, error) {
	args := m.Called(ctx, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*result.AddTeamResult), args.Error(1)
}

func (m *MockTeamRepository) Get(ctx context.Context, dto *dto.GetTeamDTO) (*result.GetTeamResult, error) {
	args := m.Called(ctx, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*result.GetTeamResult), args.Error(1)
}

func TestTeamService_Add_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockTeamRepository)
	service := NewTeamService(mockRepo, logger)

	members := []*domain.User{
		{
			Id:        "user1",
			Name:      "User 1",
			TeamName:  "team1",
			IsActive:  true,
			CreatedAt: time.Now(),
		},
		{
			Id:        "user2",
			Name:      "User 2",
			TeamName:  "team1",
			IsActive:  true,
			CreatedAt: time.Now(),
		},
	}

	req := &request.AddTeamRequest{
		TeamName: "team1",
		Members:  members,
	}

	expectedResult := &result.AddTeamResult{
		TeamName: "team1",
		Members:  members,
	}

	mockRepo.On("Add", mock.Anything, mock.MatchedBy(func(d *dto.AddTeamDTO) bool {
		return d.TeamName == "team1" && len(d.Members) == 2
	})).Return(expectedResult, nil)

	resp, err := service.Add(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "team1", resp.TeamName)
	assert.Len(t, resp.Members, 2)
	mockRepo.AssertExpectations(t)
}

func TestTeamService_Add_TeamExists(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockTeamRepository)
	service := NewTeamService(mockRepo, logger)

	req := &request.AddTeamRequest{
		TeamName: "team1",
		Members:  []*domain.User{},
	}

	mockRepo.On("Add", mock.Anything, mock.Anything).Return(nil, repository.ErrAlreadyExists)

	resp, err := service.Add(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "TEAM_EXISTS", domainErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestTeamService_Get_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockTeamRepository)
	service := NewTeamService(mockRepo, logger)

	req := &request.GetTeamRequest{
		TeamName: "team1",
	}

	members := []*domain.User{
		{
			Id:        "user1",
			Name:      "User 1",
			TeamName:  "team1",
			IsActive:  true,
			CreatedAt: time.Now(),
		},
	}

	expectedResult := &result.GetTeamResult{
		TeamName: "team1",
		Members:  members,
	}

	mockRepo.On("Get", mock.Anything, mock.MatchedBy(func(d *dto.GetTeamDTO) bool {
		return d.TeamName == "team1"
	})).Return(expectedResult, nil)

	resp, err := service.Get(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "team1", resp.TeamName)
	assert.Len(t, resp.Members, 1)
	mockRepo.AssertExpectations(t)
}

func TestTeamService_Get_TeamNotFound(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockTeamRepository)
	service := NewTeamService(mockRepo, logger)

	req := &request.GetTeamRequest{
		TeamName: "nonexistent",
	}

	mockRepo.On("Get", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)

	resp, err := service.Get(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertExpectations(t)
}
