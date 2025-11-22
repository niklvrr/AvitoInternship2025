package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"go.uber.org/zap"
)

var (
	incorrectIdError = errors.New("incorrect user id")
	setIsActiveError = errors.New("set user status error")
)

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
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, fmt.Errorf(`invalid user id: %w`, err)
	}

	dto := &dto.SetIsActiveDTO{
		UserId:   userId,
		IsActive: req.IsActive,
	}

	res, err := s.repo.SetIsActive(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf(`set user status error: %w`, err)
	}

	//return &response.SetIsActiveResponse{
	//	res.Id.String(),
	//	res.Name,
	//	res.IsActive,
	//}, nil
}

func (s *UserService) GetReview(ctx context.Context, req *request.GetReviewRequest) (*response.GetReviewResponse, error) {

}
