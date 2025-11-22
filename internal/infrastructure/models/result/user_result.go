package result

import (
	"github.com/google/uuid"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
)

type GetReviewResult struct {
	UserId uuid.UUID
	Prs    []*domain.Pr
}
