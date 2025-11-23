package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"go.uber.org/zap"
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
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewUserRepository(db *pgxpool.Pool, log *zap.Logger) *UserRepository {
	return &UserRepository{
		db:  db,
		log: log,
	}
}

func (r *UserRepository) SetIsActive(ctx context.Context, d *dto.SetIsActiveDTO) (*domain.User, error) {
	r.log.Info("set user activity",
		zap.String("user_id", d.UserId),
		zap.Bool("is_active", d.IsActive),
	)

	// Изменение поле is_active
	cmdTag, err := r.db.Exec(ctx, setIsActiveQuery, d.IsActive, d.UserId)
	if err != nil {
		r.log.Error("set user activity failed",
			zap.String("user_id", d.UserId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}

	if cmdTag.RowsAffected() == 0 {
		r.log.Warn("user not found while updating activity", zap.String("user_id", d.UserId))
		return nil, ErrNotFound
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
		r.log.Error("failed to read user after activity update",
			zap.String("user_id", d.UserId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}

	r.log.Info("user activity updated",
		zap.String("user_id", user.Id),
		zap.Bool("is_active", user.IsActive),
	)
	// Ответ
	return user, nil
}

func (r *UserRepository) GetReview(ctx context.Context, d *dto.GetReviewDTO) (*result.GetReviewResult, error) {
	r.log.Info("get user reviews", zap.String("user_id", d.UserId))

	// Читаем все PR, где пользователь назначен ревьюером
	rows, err := r.db.Query(ctx, getReviewQuery, d.UserId)
	if err != nil {
		r.log.Error("failed to load user reviews",
			zap.String("user_id", d.UserId),
			zap.Error(err),
		)
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

	r.log.Info("user reviews loaded",
		zap.String("user_id", d.UserId),
		zap.Int("prs", len(prs)),
	)
	// Ответ
	return &result.GetReviewResult{
		UserId: d.UserId,
		Prs:    prs,
	}, nil
}

func (r *UserRepository) CheckUserExists(ctx context.Context, userId string) (bool, error) {
	r.log.Debug("check user exists", zap.String("user_id", userId))

	var id string
	err := r.db.QueryRow(ctx, selectUserQuery, userId).Scan(&id, new(string), new(string), new(bool), new(sql.NullTime))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		r.log.Error("failed to check user existence",
			zap.String("user_id", userId),
			zap.Error(err),
		)
		return false, handleDBError(err)
	}

	return true, nil
}
