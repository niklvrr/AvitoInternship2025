package repository

import (
	"context"
	"database/sql"

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

	selectUserQuery = `
SELECT id, name, team_name, is_active, created_at
FROM users
WHERE id = $1`

	getReviewQuery = `
SELECT
    p.id,
    p.name,
    p.author_id,
    p.status,
    p.created_at,
    p.merged_at
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
	// Изменение поле is_active
	cmdTag, err := r.db.Exec(ctx, setIsActiveQuery, d.IsActive, d.UserId)
	if err != nil {
		return nil, handleDBError(err)
	}

	if cmdTag.RowsAffected() == 0 {
		return nil, errNotFound
	}

	// Читаем пользователя повторно, чтобы вернуть актуальные данные
	user := &domain.User{}
	err = r.db.QueryRow(ctx, selectUserQuery, d.UserId).Scan(
		&user.Id,
		&user.Name,
		&user.TeamName,
		&user.IsActive,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Ответ
	return user, nil
}

func (r *UserRepository) GetReview(ctx context.Context, d *dto.GetReviewDTO) (*result.GetReviewResult, error) {
	// Читаем все PR, где пользователь назначен ревьюером
	rows, err := r.db.Query(ctx, getReviewQuery, d.UserId)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer rows.Close()

	var prs []*domain.Pr
	for rows.Next() {
		pr := &domain.Pr{}
		var mergedAt sql.NullTime
		err = rows.Scan(
			&pr.Id,
			&pr.Name,
			&pr.AuthorId,
			&pr.Status,
			&pr.CreatedAt,
			&mergedAt,
		)
		if err != nil {
			return nil, handleDBError(err)
		}
		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}
		prs = append(prs, pr)
	}

	// Ответ
	return &result.GetReviewResult{
		UserId: d.UserId,
		Prs:    prs,
	}, nil
}
