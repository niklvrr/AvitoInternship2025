package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/dto"
	"github.com/niklvrr/AvitoInternship2025/internal/infrastructure/models/result"
	"time"
)

const (
	teamExistsQuery = `
SELECT id FROM teams
WHERE name = $1
RETURNING team_id;`

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
	var (
		teamId     uuid.UUID
		teamName   string
		created_at time.Time
	)

	// Проверяем, существует ли уже команда с таким навзанием
	err := r.db.QueryRow(ctx, insertTeamQuery, d.TeamName).Scan(&teamId)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Добавляем пользователей или обновляем данные колонок, если такие уже существуют
	for _, member := range d.Members {
		err := r.db.QueryRow(ctx, insertUserQuery, member.Id).Scan(
			&member.Id,
			&member.Name,
			&member.IsActive,
			&member.CreatedAt,
		)
		if err != nil {
			return nil, handleDBError(err)
		}
	}

	// Создаем команду
	err = r.db.QueryRow(ctx, insertTeamQuery, d.TeamName).Scan(
		&teamId,
		&teamName,
		&created_at,
	)
	if err != nil {
		return nil, handleDBError(err)
	}

	// Добавляем в таблицу членов команды
	for _, member := range d.Members {
		err := r.db.QueryRow(ctx, insertTeamMemberQuery, teamId, member.Id).Scan(
			&teamName,
			&member.Id,
			&member.Name,
			&member.IsActive,
			&member.CreatedAt,
		)
		if err != nil {
			return nil, handleDBError(err)
		}
	}

	// Ответ
	return &result.AddTeamResult{
		TeamName: teamName,
		Members:  d.Members,
	}, nil
}

func (r *TeamRepository) Get(ctx context.Context, d *dto.GetTeamDTO) (*result.GetTeamResult, error) {
	// Чтение команды и всех участников
	rows, err := r.db.Query(ctx, getTeamQuery, d.TeamName)
	if err != nil {
		return nil, handleDBError(err)
	}
	defer rows.Close()

	var members []*domain.User
	for rows.Next() {
		member := &domain.User{}
		err := rows.Scan(
			&d.TeamName,
			&member.Id,
			&member.Name,
			&member.IsActive,
			&member.CreatedAt,
		)
		if err != nil {
			return nil, handleDBError(err)
		}
		members = append(members, member)
	}

	// Ответ
	return &result.GetTeamResult{
		TeamName: d.TeamName,
		Members:  members,
	}, nil
}
