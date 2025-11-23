package dto

type CreatPrDTO struct {
	PrId     string
	PrName   string
	AuthorId string
}

type MergePrDTO struct {
	PrId string
}

type ReassignPrDTO struct {
	PrId          string
	OldReviewerId string
	ReplacedBy    string
}
