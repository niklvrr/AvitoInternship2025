package result

import (
	"github.com/google/uuid"
	"github.com/niklvrr/AvitoInternship2025/internal/domain"
)

type ReassignResult struct {
	Pr         *domain.Pr
	ReplacedBy uuid.UUID
}
