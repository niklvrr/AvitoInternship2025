package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/repository"

	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"go.uber.org/zap"
)

var (
	setIsActiveError = errors.New("set user status error")
	getReviewError   = errors.New("get review error")
)

// Интерфейс репозитория
type UserRepository interface {
	SetIsActive(ctx context.Context, d *dto.SetIsActiveDTO) (*domain.User, error)
	GetReview(ctx context.Context, d *dto.GetReviewDTO) (*result.GetReviewResult, error)
}

type UserService struct {
	repo UserRepository
	log  *zap.Logger
}

func NewUserService(repo UserRepository, log *zap.Logger) *UserService {
	return &UserService{
		repo: repo,
		log:  log,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, req *request.SetIsActiveRequest) (*response.SetIsActiveResponse, error) {
	// Проверяем корректность идентификатора
	userId, err := normalizeID(req.UserId, "user_id")
	if err != nil {
		return nil, WrapError(ErrInvalidInput, err)
	}

	// Собираем dto
	dto := &dto.SetIsActiveDTO{
		UserId:   userId,
		IsActive: req.IsActive,
	}

	// Запрос в бд
	res, err := s.repo.SetIsActive(ctx, dto)
	if err != nil {

		// Маппим ошибки
		if errors.Is(err, repository.ErrNotFound) {
			return nil, WrapError(ErrUserNotFound, err)
		}
		if errors.Is(err, repository.ErrInvalidInput) {
			return nil, WrapError(ErrInvalidInput, err)
		}

		// Неизвестная ошибка
		return nil, fmt.Errorf(`%w: %w`, setIsActiveError, err)
	}

	// Ответ
	return &response.SetIsActiveResponse{
		UserId:   userId,
		Username: res.Name,
		TeamName: res.TeamName,
		IsActive: res.IsActive,
	}, nil
}

func (s *UserService) GetReview(ctx context.Context, req *request.GetReviewRequest) (*response.GetReviewResponse, error) {
	// Проверяем корректность идентификатора
	userId, err := normalizeID(req.UserId, "user_id")
	if err != nil {
		return nil, WrapError(ErrInvalidInput, err)
	}

	// Собираем dto
	dto := &dto.GetReviewDTO{
		UserId: userId,
	}

	// Запрос в бд
	res, err := s.repo.GetReview(ctx, dto)
	if err != nil {
		// Маппим ошибки
		if errors.Is(err, repository.ErrNotFound) {
			return nil, WrapError(ErrUserNotFound, err)
		}
		if errors.Is(err, repository.ErrInvalidInput) {
			return nil, WrapError(ErrInvalidInput, err)
		}

		// Неизвестная ошибка
		return nil, fmt.Errorf(`%w: %w`, getReviewError, err)
	}

	// Ответ
	return &response.GetReviewResponse{
		UserId: userId,
		Prs:    res.Prs,
	}, nil
}
