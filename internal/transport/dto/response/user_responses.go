package response

import "github.com/niklvrr/AvitoInternship2025/internal/domain"

type SetIsActiveResponse struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type GetReviewResponse struct {
	UserId string       `json:"user_id"`
	Prs    []*domain.Pr `json:"pull_requests"`
}
