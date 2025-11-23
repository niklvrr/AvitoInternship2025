package service

import (
	"context"
	"errors"
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

// MockPrRepository мок репозитория для тестов
type MockPrRepository struct {
	mock.Mock
}

func (m *MockPrRepository) Create(ctx context.Context, dto *dto.CreatPrDTO, prReviewers []string) (*result.PrResult, error) {
	args := m.Called(ctx, dto, prReviewers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*result.PrResult), args.Error(1)
}

func (m *MockPrRepository) Merge(ctx context.Context, dto *dto.MergePrDTO) (*result.PrResult, error) {
	args := m.Called(ctx, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*result.PrResult), args.Error(1)
}

func (m *MockPrRepository) Reassign(ctx context.Context, dto *dto.ReassignPrDTO) (*result.ReassignResult, error) {
	args := m.Called(ctx, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*result.ReassignResult), args.Error(1)
}

func (m *MockPrRepository) SelectPotentialReviewers(ctx context.Context, userId string) ([]*domain.User, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func TestPrService_Create_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.CreateRequest{
		PrId:     "pr1",
		PrName:   "Test PR",
		AuthorId: "author1",
	}

	potentialReviewers := []*domain.User{
		{Id: "reviewer1", Name: "Reviewer 1", IsActive: true},
		{Id: "reviewer2", Name: "Reviewer 2", IsActive: true},
		{Id: "author1", Name: "Author", IsActive: true}, // должен быть исключен
	}

	expectedPrResult := &result.PrResult{
		Id:                "pr1",
		Name:              "Test PR",
		AuthorId:          "author1",
		Status:            "OPEN",
		AssignedReviewers: []string{"reviewer1", "reviewer2"},
		CreatedAt:         time.Now(),
		MergedAt:          nil,
	}

	mockRepo.On("SelectPotentialReviewers", mock.Anything, "author1").Return(potentialReviewers, nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(d *dto.CreatPrDTO) bool {
		return d.PrId == "pr1" && d.PrName == "Test PR" && d.AuthorId == "author1"
	}), mock.AnythingOfType("[]string")).Return(expectedPrResult, nil)

	resp, err := service.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pr1", resp.PrId)
	assert.Equal(t, "Test PR", resp.PrName)
	assert.Equal(t, "author1", resp.AuthorId)
	assert.Equal(t, "OPEN", resp.Status)
	assert.Len(t, resp.AssignedReviewers, 2)
	mockRepo.AssertExpectations(t)
}

func TestPrService_Create_AuthorNotFound(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.CreateRequest{
		PrId:     "pr1",
		PrName:   "Test PR",
		AuthorId: "author1",
	}

	mockRepo.On("SelectPotentialReviewers", mock.Anything, "author1").Return(nil, repository.ErrNotFound)

	resp, err := service.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestPrService_Create_InvalidInput_EmptyPrId(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.CreateRequest{
		PrId:     "",
		PrName:   "Test PR",
		AuthorId: "author1",
	}

	// AuthorId валидируется первым, поэтому нужно мокировать SelectPotentialReviewers
	potentialReviewers := []*domain.User{
		{Id: "reviewer1", Name: "Reviewer 1", IsActive: true},
	}
	mockRepo.On("SelectPotentialReviewers", mock.Anything, "author1").Return(potentialReviewers, nil)

	resp, err := service.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestPrService_Create_InvalidInput_EmptyAuthorId(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.CreateRequest{
		PrId:     "pr1",
		PrName:   "Test PR",
		AuthorId: "",
	}

	resp, err := service.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertNotCalled(t, "SelectPotentialReviewers")
}

func TestPrService_Create_NoReviewersAvailable(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.CreateRequest{
		PrId:     "pr1",
		PrName:   "Test PR",
		AuthorId: "author1",
	}

	// Команда существует, но нет активных ревьюеров (все неактивны или только автор)
	potentialReviewers := []*domain.User{
		{Id: "author1", Name: "Author", IsActive: true}, // Автор исключается
		{Id: "user1", Name: "User 1", IsActive: false},  // Неактивный
	}
	mockRepo.On("SelectPotentialReviewers", mock.Anything, "author1").Return(potentialReviewers, nil)

	expectedPrResult := &result.PrResult{
		Id:                "pr1",
		Name:              "Test PR",
		AuthorId:          "author1",
		Status:            "OPEN",
		AssignedReviewers: []string{}, // Пустой массив
		CreatedAt:         time.Now(),
		MergedAt:          nil,
	}
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(d *dto.CreatPrDTO) bool {
		return d.PrId == "pr1" && d.PrName == "Test PR" && d.AuthorId == "author1"
	}), mock.MatchedBy(func(reviewers []string) bool {
		return len(reviewers) == 0 // Пустой массив ревьюеров
	})).Return(expectedPrResult, nil)

	resp, err := service.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pr1", resp.PrId)
	assert.Equal(t, "Test PR", resp.PrName)
	assert.Equal(t, "OPEN", resp.Status)
	assert.Len(t, resp.AssignedReviewers, 0) // Пустой массив ревьюеров
	mockRepo.AssertExpectations(t)
}

func TestPrService_Merge_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.MergeRequest{
		PrId: "pr1",
	}

	mergedAt := time.Now()
	expectedPrResult := &result.PrResult{
		Id:                "pr1",
		Name:              "Test PR",
		AuthorId:          "author1",
		Status:            "MERGED",
		AssignedReviewers: []string{"reviewer1"},
		CreatedAt:         time.Now(),
		MergedAt:          &mergedAt,
	}

	mockRepo.On("Merge", mock.Anything, mock.MatchedBy(func(d *dto.MergePrDTO) bool {
		return d.PrId == "pr1"
	})).Return(expectedPrResult, nil)

	resp, err := service.Merge(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pr1", resp.PrId)
	assert.Equal(t, "MERGED", resp.Status)
	assert.NotNil(t, resp.MergedAt)
	mockRepo.AssertExpectations(t)
}

func TestPrService_Merge_PrNotFound(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.MergeRequest{
		PrId: "nonexistent",
	}

	mockRepo.On("Merge", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)

	resp, err := service.Merge(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestPrService_Reassign_Success(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.ReassignRequest{
		PrId:      "pr1",
		OldUserId: "old_reviewer",
	}

	potentialReviewers := []*domain.User{
		{Id: "new_reviewer", Name: "New Reviewer", IsActive: true},
		{Id: "old_reviewer", Name: "Old Reviewer", IsActive: true}, // должен быть исключен
	}

	expectedReassignResult := &result.ReassignResult{
		Pr: &result.PrResult{
			Id:                "pr1",
			Name:              "Test PR",
			AuthorId:          "author1",
			Status:            "OPEN",
			AssignedReviewers: []string{"new_reviewer"},
			CreatedAt:         time.Now(),
			MergedAt:          nil,
		},
		ReplacedBy: "new_reviewer",
	}

	mockRepo.On("SelectPotentialReviewers", mock.Anything, "old_reviewer").Return(potentialReviewers, nil)
	mockRepo.On("Reassign", mock.Anything, mock.MatchedBy(func(d *dto.ReassignPrDTO) bool {
		return d.PrId == "pr1" && d.OldReviewerId == "old_reviewer" && d.ReplacedBy == "new_reviewer"
	})).Return(expectedReassignResult, nil)

	resp, err := service.Reassign(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pr1", resp.PrId)
	assert.Equal(t, "new_reviewer", resp.ReplacedBy)
	mockRepo.AssertExpectations(t)
}

func TestPrService_Reassign_PrMerged(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.ReassignRequest{
		PrId:      "pr1",
		OldUserId: "old_reviewer",
	}

	potentialReviewers := []*domain.User{
		{Id: "new_reviewer", Name: "New Reviewer", IsActive: true},
	}

	mockRepo.On("SelectPotentialReviewers", mock.Anything, "old_reviewer").Return(potentialReviewers, nil)
	mockRepo.On("Reassign", mock.Anything, mock.Anything).Return(nil, repository.ErrPrMergedStatus)

	resp, err := service.Reassign(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var domainErr *DomainError
	assert.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "PR_MERGED", domainErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestPrService_Reassign_NoCandidate(t *testing.T) {
	logger := zap.NewNop()
	mockRepo := new(MockPrRepository)
	service := NewPrService(mockRepo, logger)

	req := &request.ReassignRequest{
		PrId:      "pr1",
		OldUserId: "old_reviewer",
	}

	// Только неактивные ревьюеры или только старый ревьюер
	potentialReviewers := []*domain.User{
		{Id: "old_reviewer", Name: "Old Reviewer", IsActive: true},
	}

	mockRepo.On("SelectPotentialReviewers", mock.Anything, "old_reviewer").Return(potentialReviewers, nil)

	resp, err := service.Reassign(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrNoCandidate, err)
	mockRepo.AssertExpectations(t)
}

func TestFindReviewers(t *testing.T) {
	tests := []struct {
		name          string
		potential     []*domain.User
		excludedId    string
		reviewerCount int
		expectedCount int
		expectedError error
	}{
		{
			name: "successful selection",
			potential: []*domain.User{
				{Id: "user1", IsActive: true},
				{Id: "user2", IsActive: true},
				{Id: "excluded", IsActive: true},
			},
			excludedId:    "excluded",
			reviewerCount: 2,
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "no active reviewers",
			potential: []*domain.User{
				{Id: "user1", IsActive: false},
				{Id: "user2", IsActive: false},
			},
			excludedId:    "excluded",
			reviewerCount: 2,
			expectedCount: 0,
			expectedError: errors.New("no active reviewer available"),
		},
		{
			name: "less reviewers than requested",
			potential: []*domain.User{
				{Id: "user1", IsActive: true},
			},
			excludedId:    "excluded",
			reviewerCount: 2,
			expectedCount: 1,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := findReviewers(tt.potential, tt.excludedId, tt.reviewerCount)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)
			// Проверяем, что excludedId не в результате
			for _, id := range result {
				assert.NotEqual(t, tt.excludedId, id)
			}
		})
	}
}
