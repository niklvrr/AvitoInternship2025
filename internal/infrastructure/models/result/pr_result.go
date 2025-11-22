package result

import (
	"github.com/google/uuid"
	"time"
)

type ReassignResult struct {
	Pr         *PrResult
	ReplacedBy uuid.UUID
}

type PrResult struct {
	Id                uuid.UUID
	Name              string
	AuthorId          uuid.UUID
	Status            string
	CreatedAt         time.Time
	MergedAt          time.Time
	AssignedReviewers []*uuid.UUID
}
