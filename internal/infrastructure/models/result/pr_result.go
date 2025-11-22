package result

import "time"

type ReassignResult struct {
	Pr         *PrResult
	ReplacedBy string
}

type PrResult struct {
	Id                string
	Name              string
	AuthorId          string
	Status            string
	CreatedAt         time.Time
	MergedAt          *time.Time
	AssignedReviewers []string
}
