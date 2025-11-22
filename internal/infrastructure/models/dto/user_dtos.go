package dto

import "github.com/google/uuid"

type SetIsActiveDTO struct {
	UserId   uuid.UUID `json:"userId"`
	IsActive bool      `json:"is_active"`
}

type GetReviewDTO struct {
	UserId uuid.UUID `json:"user_id"`
}
