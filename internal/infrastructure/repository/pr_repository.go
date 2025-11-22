package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
)

const (
	insertPrQuery = `
INSERT INTO prs(id, name, author_id)
VALUES ($1, $2, $3)
RETURNING id, name, author_id, status, created_at, merged_at;`

	selectTeamQuery = `
SELECT team_id FROM team_members
WHERE user_id = $1;`

	selectTeamMembersQuery = `
SELECT user_id FROM team_members
WHERE team_id = $1;`

	insertPrReviewerQuery = `
INSERT INTO pr_reviewers(user_id, pr_id)
VALUES ($1, $2)
ON CONFLICT (user_id, pr_id) DO NOTHING;`

	mergePrQuery = `
UPDATE prs
SET status = 'MERGED',
    merged_at = CURRENT_TIMESTAMP
WHERE id = $1 AND status <> 'MERGED';`

	deletePrReviewerQuery = `
DELETE FROM pr_reviewers
WHERE pr_id = $1 AND user_id = $2;`

	selectPrReviewerQuery = `
SELECT user_id FROM pr_reviewers
WHERE pr_id = $1;`

	selectPrQuery = `
SELECT id, name, author_id, status, created_at, merged_at FROM prs
WHERE id = $1;`
)

type PrRepository struct {
	db *pgxpool.Pool
}

func NewPrRepository(db *pgxpool.Pool) *PrRepository {
	return &PrRepository{db: db}
}

func (r *PrRepository) Create(ctx context.Context, d *dto.CreatPrDTO, prReviewers []*uuid.UUID) (*result.PrResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer tx.Rollback(ctx)

	prRes := &result.PrResult{}

	// Создание pr
	err = tx.QueryRow(ctx, insertPrQuery, d.PrId, d.PrName, d.AuthorId).Scan(
		&prRes.Id,
		&prRes.Name,
		&prRes.AuthorId,
		&prRes.Status,
		&prRes.CreatedAt,
		&prRes.MergedAt,
	)
	if err != nil {
		return nil, handleDBError(err)
	}

	assignedReviewers := make([]*uuid.UUID, 0, len(prReviewers))
	// Создание pr reviewers
	for _, prMemberId := range prReviewers {
		if prMemberId == nil {
			continue
		}
		if _, err := tx.Exec(ctx, insertPrReviewerQuery, prMemberId, prRes.Id); err != nil {
			return nil, handleDBError(err)
		}
		assignedReviewers = append(assignedReviewers, prMemberId)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, handleDBError(err)
	}

	prRes.AssignedReviewers = assignedReviewers

	// Ответ
	return prRes, nil
}

func (r *PrRepository) Merge(ctx context.Context, d *dto.MergePrDTO) (*result.PrResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer tx.Rollback(ctx)

	// Получаем текущее состояние pr
	prRes, err := readPr(ctx, tx, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Меняем статус, если PR еще не merged
	if prRes.Status != "MERGED" {
		cmdTag, err := tx.Exec(ctx, mergePrQuery, d.PrId)
		if err != nil {
			return nil, handleDBError(err)
		}

		if cmdTag.RowsAffected() == 0 {
			// PR уже в состоянии MERGED, перечитываем состояние
			prRes, err = readPr(ctx, tx, d.PrId)
			if err != nil {
				return nil, handleDBError(err)
			}
		} else {
			prRes, err = readPr(ctx, tx, d.PrId)
			if err != nil {
				return nil, handleDBError(err)
			}
		}
	}

	// Чтение всех ревьюеров этого pr
	prReviewers, err := readReviewers(ctx, tx, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}
	prRes.AssignedReviewers = prReviewers

	if err := tx.Commit(ctx); err != nil {
		return nil, handleDBError(err)
	}

	// Ответ
	return prRes, nil
}

func (r *PrRepository) Reassign(ctx context.Context, d *dto.ReassignPrDTO) (*result.ReassignResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer tx.Rollback(ctx)

	// Убедимся, что PR существует
	prRes, err := readPr(ctx, tx, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Удалить старого ревьюера из таблицы pr_reviewers
	cmdTag, err := tx.Exec(ctx, deletePrReviewerQuery, d.PrId, d.OldReviewerId)
	if err != nil {
		return nil, handleDBError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return nil, errNotFound
	}

	// Добавить нового ревьюера
	if _, err := tx.Exec(ctx, insertPrReviewerQuery, d.NewReviewerId, d.PrId); err != nil {
		return nil, handleDBError(err)
	}

	// Чтение всех ревьюеров для этого pr
	prReviewers, err := readReviewers(ctx, tx, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}
	prRes.AssignedReviewers = prReviewers

	if err := tx.Commit(ctx); err != nil {
		return nil, handleDBError(err)
	}

	// Ответ
	return &result.ReassignResult{
		Pr:         prRes,
		ReplacedBy: d.ReplacedBy,
	}, nil
}

// вспомогательная функция для поиска возможных ревьюеров, вызывается в сервисном слое для выбора ревьеров для pr
func (r *PrRepository) SelectPotentialReviewers(ctx context.Context, userId uuid.UUID) ([]*uuid.UUID, error) {
	// Чтение команды пользователя
	var teamId uuid.UUID
	err := r.db.QueryRow(ctx, selectTeamQuery, userId).Scan(&teamId)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Чтение всех участников команды
	rows, err := r.db.Query(ctx, selectTeamMembersQuery, teamId)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer rows.Close()

	var users []*uuid.UUID
	for rows.Next() {
		var userId uuid.UUID
		err = rows.Scan(&userId)
		if err != nil {
			return nil, handleDBError(err)
		}
		users = append(users, &userId)
	}

	// Ответ
	return users, nil
}

type queryExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// вспомогательная функция для чтения всех ревьюеров для pr
func readReviewers(ctx context.Context, exec queryExecutor, prId uuid.UUID) ([]*uuid.UUID, error) {
	rows, err := exec.Query(ctx, selectPrReviewerQuery, prId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prReviewers []*uuid.UUID
	for rows.Next() {
		var prReviewerId uuid.UUID
		if err = rows.Scan(&prReviewerId); err != nil {
			return nil, err
		}
		prReviewer := prReviewerId
		prReviewers = append(prReviewers, &prReviewer)
	}
	return prReviewers, nil
}

// вспомогательная функция для чтения данных для pr
func readPr(ctx context.Context, exec queryExecutor, prId uuid.UUID) (*result.PrResult, error) {
	prRes := &result.PrResult{}
	err := exec.QueryRow(ctx, selectPrQuery, prId).Scan(
		&prRes.Id,
		&prRes.Name,
		&prRes.AuthorId,
		&prRes.Status,
		&prRes.CreatedAt,
		&prRes.MergedAt,
	)
	if err != nil {
		return nil, err
	}

	return prRes, nil
}
