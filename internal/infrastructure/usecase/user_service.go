package usecase

import (
	"context"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/request"
	"github.com/niklvrr/AvitoInternship2025/internal/transport/dto/response"
	"go.uber.org/zap"
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

}

func (s *UserService) GetReview(ctx context.Context, req *request.GetReviewRequest) (*response.GetReviewResponse, error) {

}
