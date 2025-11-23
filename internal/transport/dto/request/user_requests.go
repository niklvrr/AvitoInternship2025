package request

type SetIsActiveRequest struct {
	UserId   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type GetReviewRequest struct {
	UserId string `json:"user_id"`
}
