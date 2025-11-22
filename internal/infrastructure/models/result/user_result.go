package result

import "github.com/niklvrr/AvitoInternship2025/internal/domain"

type GetReviewResult struct {
	UserId string
	Prs    []*domain.Pr
}
