package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
)

const (
	insertPrQuery = `
INSERT INTO prs(id, name, author_id)
VALUES ($1, $2, $3)
RETURNING id;`

	selectTeamQuery = `
SELECT team_id FROM team_members
WHERE user_id = $1;`

	selectTeamMembersQuery = `
SELECT user_id FROM team_members
WHERE team_id = $1;`

	insertPrMemberQuery = `
INSERT INTO team_members(user_id, pr_id)
VALUES ($1, $2);`

	mergePrMemberQuery = `
UPDATE prs
SET status = 'MERGED'
WHERE id = $1;`

	deletePrMemberQuery = `
DELETE FROM pr_members
WHERE user_id = $1;`
)

type PrRepository struct {
	db *pgxpool.Pool
}

func NewPrRepository(db *pgxpool.Pool) *PrRepository {
	return &PrRepository{db: db}
}

func (r *PrRepository) Create(ctx context.Context, d *dto.CreatPrDTO, prMembers []*uuid.UUID) (*domain.Pr, error) {
	var prId uuid.UUID
	err := r.db.QueryRow(ctx, insertPrMemberQuery, d.PrId, d.PrName, d.AuthorId).Scan(&prId)
	if err != nil {
		return nil, handleDBError(err)
	}

	for _, prMemberId := range prMembers {
		err := r.db.QueryRow(ctx, insertPrMemberQuery, prMemberId, prId)
		if err != nil {
			return
		}
	}
}

func (r *PrRepository) Merge(ctx context.Context, d *dto.MergePrDTO) (*domain.Pr, error) {

}

func (r *PrRepository) Reassign(ctx context.Context, d *dto.ReassignPrDTO) (*result.ReassignResult, error) {

}

func (r *PrRepository) SelectPotentialReviewers(ctx context.Context, userId uuid.UUID) ([]*uuid.UUID, error) {
	var teamId uuid.UUID
	err := r.db.QueryRow(ctx, selectTeamQuery, userId).Scan(&teamId)
	if err != nil {
		return nil, handleDBError(err)
	}

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

	return users, nil
}
