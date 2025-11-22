package request

type CreateRequest struct {
	PrId     string `json:"pull_request_id"`
	PrName   string `json:"pull_request_name"`
	AuthorId string `json:"author_id"`
}

type MergeRequest struct {
	PrId string `json:"pull_request_id"`
}

type ReassignRequest struct {
	PrId          string `json:"pull_request_id"`
	OldReviewerId string `json:"old_reviewer_id"`
}
