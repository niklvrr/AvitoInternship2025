package response

type CreateResponse struct {
	PrId              string   `json:"pull_request_id"`
	PrName            string   `json:"pull_request_name"`
	AuthorId          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type MergeResponse struct {
	PrId              string   `json:"pull_request_id"`
	PrName            string   `json:"pull_request_name"`
	AuthorId          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	MergedAt          string   `json:"merged_at"`
}

type ReassignResponse struct {
	PrId              string   `json:"pull_request_id"`
	PrName            string   `json:"pull_request_name"`
	AuthorId          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	ReplacedBy        string   `json:"replaced_by"`
}
