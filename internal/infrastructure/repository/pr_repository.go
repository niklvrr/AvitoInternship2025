package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"go.uber.org/zap"
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
SELECT
    u.id,
    u.name,
    u.team_name,
    u.is_active,
    u.created_at
FROM team_members tm
JOIN users u ON u.id = tm.user_id
WHERE tm.team_id = $1;`

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
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewPrRepository(db *pgxpool.Pool, log *zap.Logger) *PrRepository {
	return &PrRepository{
		db:  db,
		log: log,
	}
}

func (r *PrRepository) Create(ctx context.Context, d *dto.CreatPrDTO, prReviewers []string) (*result.PrResult, error) {
	r.log.Info("create PR started",
		zap.String("pr_id", d.PrId),
		zap.String("author_id", d.AuthorId),
		zap.Int("reviewers_requested", len(prReviewers)),
	)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer tx.Rollback(ctx)

	prRes := &result.PrResult{}
	var mergedAt sql.NullTime

	// Создаем PR и читаем его состояние
	err = tx.QueryRow(ctx, insertPrQuery, d.PrId, d.PrName, d.AuthorId).Scan(
		&prRes.Id,
		&prRes.Name,
		&prRes.AuthorId,
		&prRes.Status,
		&prRes.CreatedAt,
		&mergedAt,
	)
	if err != nil {
		r.log.Error("failed to insert PR",
			zap.String("pr_id", d.PrId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}
	if mergedAt.Valid {
		prRes.MergedAt = &mergedAt.Time
	}

	assignedReviewers := make([]string, 0, len(prReviewers))
	// Записываем назначенных ревьюеров
	for _, prMemberId := range prReviewers {
		if prMemberId == "" {
			continue
		}
		if _, err := tx.Exec(ctx, insertPrReviewerQuery, prMemberId, prRes.Id); err != nil {
			r.log.Error("failed to insert PR reviewer",
				zap.String("pr_id", d.PrId),
				zap.String("reviewer_id", prMemberId),
				zap.Error(err),
			)
			return nil, handleDBError(err)
		}
		assignedReviewers = append(assignedReviewers, prMemberId)
	}

	if err := tx.Commit(ctx); err != nil {
		r.log.Error("failed to commit PR creation", zap.String("pr_id", d.PrId), zap.Error(err))
		return nil, handleDBError(err)
	}

	prRes.AssignedReviewers = assignedReviewers

	r.log.Info("PR created",
		zap.String("pr_id", prRes.Id),
		zap.Int("assigned_reviewers", len(prRes.AssignedReviewers)),
	)
	// Ответ
	return prRes, nil
}

func (r *PrRepository) Merge(ctx context.Context, d *dto.MergePrDTO) (*result.PrResult, error) {
	r.log.Info("merge PR started", zap.String("pr_id", d.PrId))

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer tx.Rollback(ctx)

	// Получаем текущее состояние pr
	prRes, err := readPr(ctx, tx, d.PrId)
	if err != nil {
		r.log.Error("failed to load PR before merge", zap.String("pr_id", d.PrId), zap.Error(err))
		return nil, handleDBError(err)
	}

	// Меняем статус, если PR еще не merged
	if prRes.Status != "MERGED" {
		cmdTag, err := tx.Exec(ctx, mergePrQuery, d.PrId)
		if err != nil {
			r.log.Error("failed to update PR status to MERGED",
				zap.String("pr_id", d.PrId),
				zap.Error(err),
			)
			return nil, handleDBError(err)
		}

		if cmdTag.RowsAffected() == 0 {
			// PR уже в состоянии MERGED, перечитываем состояние
			prRes, err = readPr(ctx, tx, d.PrId)
			if err != nil {
				r.log.Error("failed to reload merged PR state",
					zap.String("pr_id", d.PrId),
					zap.Error(err),
				)
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
		r.log.Error("failed to read PR reviewers after merge",
			zap.String("pr_id", d.PrId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}
	prRes.AssignedReviewers = prReviewers

	if err := tx.Commit(ctx); err != nil {
		r.log.Error("failed to commit merge transaction",
			zap.String("pr_id", d.PrId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}

	r.log.Info("PR merged",
		zap.String("pr_id", prRes.Id),
		zap.String("status", prRes.Status),
	)
	// Ответ
	return prRes, nil
}

func (r *PrRepository) Reassign(ctx context.Context, d *dto.ReassignPrDTO) (*result.ReassignResult, error) {
	r.log.Info("reassign reviewer started",
		zap.String("pr_id", d.PrId),
		zap.String("old_reviewer_id", d.OldReviewerId),
		zap.String("new_reviewer_id", d.ReplacedBy),
	)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer tx.Rollback(ctx)

	// Убедимся, что PR существует
	prRes, err := readPr(ctx, tx, d.PrId)
	if err != nil {
		r.log.Error("failed to load PR before reassign",
			zap.String("pr_id", d.PrId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}

	// Не даем переназначать ревьюеров после MERGED
	if prRes.Status == "MERGED" {
		return nil, errInvalidInput
	}

	// Удалить старого ревьюера из таблицы pr_reviewers
	cmdTag, err := tx.Exec(ctx, deletePrReviewerQuery, d.PrId, d.OldReviewerId)
	if err != nil {
		r.log.Error("failed to remove old reviewer",
			zap.String("pr_id", d.PrId),
			zap.String("old_reviewer_id", d.OldReviewerId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}
	if cmdTag.RowsAffected() == 0 {
		r.log.Warn("old reviewer not found on PR",
			zap.String("pr_id", d.PrId),
			zap.String("old_reviewer_id", d.OldReviewerId),
		)
		return nil, errNotFound
	}

	// Добавить нового ревьюера
	if _, err := tx.Exec(ctx, insertPrReviewerQuery, d.ReplacedBy, d.PrId); err != nil {
		return nil, handleDBError(err)
	}

	// Чтение всех ревьюеров для этого pr
	prReviewers, err := readReviewers(ctx, tx, d.PrId)
	if err != nil {
		r.log.Error("failed to read reviewers after reassign",
			zap.String("pr_id", d.PrId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}
	prRes.AssignedReviewers = prReviewers

	if err := tx.Commit(ctx); err != nil {
		r.log.Error("failed to commit reassign transaction",
			zap.String("pr_id", d.PrId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}

	r.log.Info("reviewer reassigned",
		zap.String("pr_id", prRes.Id),
		zap.Strings("assigned_reviewers", prRes.AssignedReviewers),
		zap.String("replaced_by", d.ReplacedBy),
	)
	// Ответ
	return &result.ReassignResult{
		Pr:         prRes,
		ReplacedBy: d.ReplacedBy,
	}, nil
}

// вспомогательная функция для поиска возможных ревьюеров, вызывается в сервисном слое для выбора ревьеров для pr
func (r *PrRepository) SelectPotentialReviewers(ctx context.Context, userId string) ([]*domain.User, error) {
	r.log.Debug("select potential reviewers", zap.String("user_id", userId))

	// Чтение команды пользователя
	var teamId string
	err := r.db.QueryRow(ctx, selectTeamQuery, userId).Scan(&teamId)
	if err != nil {
		r.log.Error("failed to load team for user",
			zap.String("user_id", userId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}

	// Чтение всех участников команды
	rows, err := r.db.Query(ctx, selectTeamMembersQuery, teamId)
	if err != nil {
		r.log.Error("failed to load team members",
			zap.String("team_id", teamId),
			zap.Error(err),
		)
		return nil, handleDBError(err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		member := &domain.User{}
		err = rows.Scan(
			&member.Id,
			&member.Name,
			&member.TeamName,
			&member.IsActive,
			&member.CreatedAt,
		)
		if err != nil {
			return nil, handleDBError(err)
		}
		users = append(users, member)
	}

	r.log.Debug("potential reviewers loaded",
		zap.String("team_id", teamId),
		zap.Int("members", len(users)),
	)
	// Ответ
	return users, nil
}

type queryExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// вспомогательная функция для чтения всех ревьюеров для pr
func readReviewers(ctx context.Context, exec queryExecutor, prId string) ([]string, error) {
	rows, err := exec.Query(ctx, selectPrReviewerQuery, prId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prReviewers []string
	for rows.Next() {
		var prReviewerId string
		if err = rows.Scan(&prReviewerId); err != nil {
			return nil, err
		}
		prReviewers = append(prReviewers, prReviewerId)
	}
	return prReviewers, nil
}

// вспомогательная функция для чтения данных для pr
func readPr(ctx context.Context, exec queryExecutor, prId string) (*result.PrResult, error) {
	prRes := &result.PrResult{}
	var mergedAt sql.NullTime
	err := exec.QueryRow(ctx, selectPrQuery, prId).Scan(
		&prRes.Id,
		&prRes.Name,
		&prRes.AuthorId,
		&prRes.Status,
		&prRes.CreatedAt,
		&mergedAt,
	)
	if err != nil {
		return nil, err
	}
	if mergedAt.Valid {
		prRes.MergedAt = &mergedAt.Time
	}

	return prRes, nil
}
