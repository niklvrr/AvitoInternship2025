package dto

import "github.com/google/uuid"

type CreatPrDTO struct {
	PrId     uuid.UUID
	PrName   string
	AuthorId uuid.UUID
}

type MergePrDTO struct {
	PrId uuid.UUID
}

type ReassignPrDTO struct {
	PrId          uuid.UUID
	OldReviewerId uuid.UUID
	NewReviewerId uuid.UUID
	ReplacedBy    uuid.UUID
}
