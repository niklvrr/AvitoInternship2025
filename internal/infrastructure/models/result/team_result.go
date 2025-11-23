package result

import "github.com/niklvrr/AvitoInternship2025/internal/domain"

type AddTeamResult struct {
	TeamName string
	Members  []*domain.User
}

type GetTeamResult struct {
	TeamName string
	Members  []*domain.User
}
