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
	PrId      string `json:"pull_request_id"`
	OldUserId string `json:"old_user_id"`
}
