package request

import "github.com/niklvrr/AvitoInternship2025/internal/domain"

type AddTeamRequest struct {
	TeamName string         `json:"team_name"`
	Members  []*domain.User `json:"members"`
}

type GetTeamRequest struct {
	TeamName string `json:"team_name"`
}
