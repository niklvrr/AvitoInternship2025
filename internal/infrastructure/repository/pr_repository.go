package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"time"
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
INSERT INTO team_members(user_id, pr_id)
VALUES ($1, $2);`

	mergePrReviewerQuery = `
UPDATE prs
SET status = 'MERGED'
WHERE id = $1;`

	deletePrReviewerQuery = `
DELETE FROM pr_members
WHERE user_id = $1;`

	selectPrReviewerQuery = `
SELECT user_id FROM pr_reviewers
WHERE pr_id = $1;`

	selectPrQuery = `
SELECT * FROM prs
WHERE id = $1;`
)

type PrRepository struct {
	db *pgxpool.Pool
}

func NewPrRepository(db *pgxpool.Pool) *PrRepository {
	return &PrRepository{db: db}
}

func (r *PrRepository) Create(ctx context.Context, d *dto.CreatPrDTO, prReviewers []*uuid.UUID) (*result.PrResult, error) {
	var (
		prStatus  string
		createdAt time.Time
	)

	// Создание pr
	err := r.db.QueryRow(ctx, insertPrReviewerQuery, d.PrId, d.PrName, d.AuthorId).Scan(
		&d.PrId,
		&d.PrName,
		&d.AuthorId,
		&prStatus,
		&createdAt)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Создание pr reviewers
	for _, prMemberId := range prReviewers {
		err := r.db.QueryRow(ctx, insertPrReviewerQuery, prMemberId, d.PrId).Scan(&d.PrId)
		if err != nil {
			return nil, handleDBError(err)
		}
	}

	// Ответ
	return &result.PrResult{
		Id:                d.PrId,
		Name:              d.PrName,
		AuthorId:          d.AuthorId,
		Status:            prStatus,
		CreatedAt:         createdAt,
		AssignedReviewers: prReviewers,
	}, nil
}

func (r *PrRepository) Merge(ctx context.Context, d *dto.MergePrDTO) (*result.PrResult, error) {
	// Изменение статуса pr на 'MERGE'
	cmdTag, err := r.db.Exec(ctx, mergePrReviewerQuery, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}

	if cmdTag.RowsAffected() == 0 {
		return nil, errNotFound
	}

	// Чтение данных pr для ответа
	prRes, err := r.selectPr(ctx, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Чтение всех ревьюеров этого pr
	prReviewers, err := r.selectReviewers(ctx, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}
	prRes.AssignedReviewers = prReviewers

	// Ответ
	return prRes, nil
}

func (r *PrRepository) Reassign(ctx context.Context, d *dto.ReassignPrDTO) (*result.ReassignResult, error) {
	// Удалить старого ревьюера из таблицы pr_reviewers
	cmdTag, err := r.db.Exec(ctx, deletePrReviewerQuery, d.OldReviewerId)
	if err != nil {
		return nil, handleDBError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		return nil, errNotFound
	}

	// Добавить в таблицу pr_reviewers нового ревьюера
	err = r.db.QueryRow(ctx, insertPrReviewerQuery, d.NewReviewerId, d.PrId).Scan()
	if err != nil {
		return nil, handleDBError(err)
	}

	// Чтение данных pr для ответа
	prRes, err := r.selectPr(ctx, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Чтение всех ревьюеров для этого pr
	prReviewers, err := r.selectReviewers(ctx, d.PrId)
	if err != nil {
		return nil, handleDBError(err)
	}
	prRes.AssignedReviewers = prReviewers

	// Ответ
	return &result.ReassignResult{
		Pr:         prRes,
		ReplacedBy: d.ReplacedBy,
	}, nil
}

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

func (r *PrRepository) selectReviewers(ctx context.Context, prId uuid.UUID) ([]*uuid.UUID, error) {
	rows, err := r.db.Query(ctx, selectPrReviewerQuery, prId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prReviewers []*uuid.UUID
	for rows.Next() {
		var prReviewerId uuid.UUID
		err = rows.Scan(&prReviewerId)
		if err != nil {
			return nil, err
		}
		prReviewers = append(prReviewers, &prReviewerId)
	}
	return prReviewers, nil
}

func (r *PrRepository) selectPr(ctx context.Context, prId uuid.UUID) (*result.PrResult, error) {
	prRes := &result.PrResult{}
	err := r.db.QueryRow(ctx, selectPrQuery, prId).Scan(
		&prRes.Id,
		&prRes.Name,
		&prRes.AuthorId,
		&prRes.Status,
		&prRes.CreatedAt,
		&prRes.AssignedReviewers,
	)
	if err != nil {
		return nil, err
	}

	return prRes, nil
}
