package dto

type SetIsActiveDTO struct {
	UserId   string `json:"userId"`
	IsActive bool   `json:"is_active"`
}

type GetReviewDTO struct {
	UserId string `json:"user_id"`
}
