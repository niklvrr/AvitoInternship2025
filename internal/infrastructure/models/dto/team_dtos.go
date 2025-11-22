package dto

import "github.com/niklvrr/AvitoInternship2025/internal/domain"

type AddTeamDTO struct {
	TeamName string         `json:"team_name"`
	Members  []*domain.User `json:"members"`
}

type GetTeamDTO struct {
	TeamName string `json:"team_name"`
}
