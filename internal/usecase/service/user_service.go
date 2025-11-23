package service

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
	s.log.Info("setIsActive request accepted",
		zap.String("user_id", req.UserId),
		zap.Bool("is_active", req.IsActive),
	)

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
		s.log.Error("failed to set user active status",
			zap.String("user_id", userId),
			zap.Bool("is_active", req.IsActive),
			zap.Error(err),
		)

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

	s.log.Info("user active status updated",
		zap.String("user_id", userId),
		zap.String("username", res.Name),
		zap.Bool("is_active", res.IsActive),
	)

	// Ответ
	return &response.SetIsActiveResponse{
		UserId:   userId,
		Username: res.Name,
		TeamName: res.TeamName,
		IsActive: res.IsActive,
	}, nil
}

func (s *UserService) GetReview(ctx context.Context, req *request.GetReviewRequest) (*response.GetReviewResponse, error) {
	s.log.Info("getReview request accepted",
		zap.String("user_id", req.UserId),
	)

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
		s.log.Error("failed to get user reviews",
			zap.String("user_id", userId),
			zap.Error(err),
		)

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

	s.log.Info("user reviews retrieved",
		zap.String("user_id", userId),
		zap.Int("pull_requests_count", len(res.Prs)),
	)

	// Ответ
	return &response.GetReviewResponse{
		UserId: userId,
		Prs:    res.Prs,
	}, nil
}
