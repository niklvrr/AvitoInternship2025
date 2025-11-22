package response

import "github.com/niklvrr/AvitoInternship2025/internal/domain"

type AddTeamResponse struct {
	TeamName string         `json:"team_name"`
	Members  []*domain.User `json:"members"`
}

type GetTeamResponse struct {
	TeamName string         `json:"team_name"`
	Members  []*domain.User `json:"members"`
}
