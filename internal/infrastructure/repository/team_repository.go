package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
)

const (
	teamExistsQuery = `
SELECT id FROM teams
WHERE name = $1;`

	insertUserQuery = `
INSERT INTO users (id, name, is_active)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
	SET name = EXCLUDED.name,
	    is_active = EXCLUDED.is_active
RETURNING id, name, is_active, created_at;`

	insertTeamQuery = `
INSERT INTO teams (id, name) 
VALUES ($1, $2) 
RETURNING id, name, created_at;`

	insertTeamMemberQuery = `
INSERT INTO team_members (team_id, user_id)
VALUES ($1, $2)
ON CONFLICT (team_id, user_id) DO UPDATE
	SET joined_at = EXCLUDED.joined_at
RETURNING team_id, user_id, joined_at;`

	getTeamQuery = `
SELECT
    t.name        AS team_name,
    u.id          AS user_id,
    u.name        AS username,
    u.is_active   AS user_is_active,
    u.created_at  AS user_created_at,
FROM teams t
LEFT JOIN team_members tm ON tm.team_id = t.id
LEFT JOIN users u ON u.id = tm.user_id
WHERE t.name = $1
ORDER BY u.created_at ASC;`
)

type TeamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{
		db: db,
	}
}

func (r *TeamRepository) Add(ctx context.Context, d *dto.AddTeamDTO) (*result.AddTeamResult, error) {

}

func (r *TeamRepository) Get(ctx context.Context, d *dto.GetTeamDTO) (*result.GetTeamResult, error) {

}
