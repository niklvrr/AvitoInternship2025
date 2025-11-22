package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
)

const (
	setIsActiveQuery = `
UPDATE users
SET is_active = $1
WHERE id = $2`

	getReviewQuery = `
SELECT
    p.id,
    p.name,
    p.author_id,
    p.team_id,
    p.status,
    p.created_at,
    p.merged_at,
FROM pr_reviewers prr
JOIN prs p ON prr.pr_id = p.id
WHERE prr.user_id = $1
ORDER BY p.created_at DESC;`
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

func (r *UserRepository) GetReview(ctx context.Context, d *dto.GetReviewDTO) (*result.GetReviewResult, error) {

}
