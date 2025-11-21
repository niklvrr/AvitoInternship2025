package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/dto"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) SetIsActive(ctx context.Context, d *dto.SetIsActiveDTO) (*domain.User, error) {

}

func (r *UserRepository) GetReview(ctx context.Context, d *dto.GetReviewDTO) (*domain.User, error) {

}
