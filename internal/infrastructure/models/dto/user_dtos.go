package dto

type SetIsActiveDTO struct {
	IsActive bool `json:"is_active"`
}

type GetReviewDTO struct {
	UserId string `json:"user_id"`
}
