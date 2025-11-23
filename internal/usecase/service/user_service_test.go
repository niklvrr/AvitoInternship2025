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

// MockUserRepository мок репозитория для тестов
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) SetIsActive(ctx context.Context, d *dto.SetIsActiveDTO) (*domain.User, error) {
	args := m.Called(ctx, d)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetReview(ctx context.Context, d *dto.GetReviewDTO) (*result.GetReviewResult, error) {
	args := m.Called(ctx, d)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*result.GetReviewResult), args.Error(1)
}

func TestUserService_SetIsActive_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, logger)

	req := &request.SetIsActiveRequest{
		UserId:   "user1",
		IsActive: true,
	}

	expectedUser := &domain.User{
		Id:        "user1",
		Name:      "Test User",
		TeamName:  "team1",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	mockRepo.On("SetIsActive", mock.Anything, mock.MatchedBy(func(d *dto.SetIsActiveDTO) bool {
		return d.UserId == "user1" && d.IsActive == true
	})).Return(expectedUser, nil)

	resp, err := service.SetIsActive(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user1", resp.UserId)
	assert.Equal(t, "Test User", resp.Username)
	assert.Equal(t, "team1", resp.TeamName)
	assert.True(t, resp.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestUserService_SetIsActive_UserNotFound(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, logger)

	req := &request.SetIsActiveRequest{
		UserId:   "user1",
		IsActive: true,
	}

	mockRepo.On("SetIsActive", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)

	resp, err := service.SetIsActive(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_SetIsActive_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, logger)

	req := &request.SetIsActiveRequest{
		UserId:   "",
		IsActive: true,
	}

	resp, err := service.SetIsActive(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertNotCalled(t, "SetIsActive")
}

func TestUserService_GetReview_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, logger)

	req := &request.GetReviewRequest{
		UserId: "user1",
	}

	expectedPrs := []*domain.Pr{
		{
			Id:        "pr1",
			Name:      "PR 1",
			AuthorId:  "author1",
			Status:    "OPEN",
			CreatedAt: time.Now(),
		},
		{
			Id:        "pr2",
			Name:      "PR 2",
			AuthorId:  "author2",
			Status:    "MERGED",
			CreatedAt: time.Now(),
			MergedAt:  func() *time.Time { t := time.Now(); return &t }(),
		},
	}

	expectedResult := &result.GetReviewResult{
		UserId: "user1",
		Prs:    expectedPrs,
	}

	mockRepo.On("GetReview", mock.Anything, mock.MatchedBy(func(d *dto.GetReviewDTO) bool {
		return d.UserId == "user1"
	})).Return(expectedResult, nil)

	resp, err := service.GetReview(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "user1", resp.UserId)
	assert.Len(t, resp.Prs, 2)
	assert.Equal(t, "pr1", resp.Prs[0].Id)
	assert.Equal(t, "pr2", resp.Prs[1].Id)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetReview_UserNotFound(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, logger)

	req := &request.GetReviewRequest{
		UserId: "user1",
	}

	mockRepo.On("GetReview", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)

	resp, err := service.GetReview(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetReview_InvalidInput(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo, logger)

	req := &request.GetReviewRequest{
		UserId: "",
	}

	resp, err := service.GetReview(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertNotCalled(t, "GetReview")
}
