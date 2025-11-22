package domain

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id        uuid.UUID
	Name      string
	IsActive  bool
	CreatedAt time.Time
}

type Team struct {
	Id        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type TeamMember struct {
	TeamId   uuid.UUID
	UserId   uuid.UUID
	JoinedAt time.Time
}

type Pr struct {
	Id        uuid.UUID
	Name      string
	AuthorId  uuid.UUID
	Status    string
	CreatedAt time.Time
	MergedAt  time.Time
}

type PrReviewer struct {
	UserId     uuid.UUID
	PrId       uuid.UUID
	AssignedAt time.Time
}
