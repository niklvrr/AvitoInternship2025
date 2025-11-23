package domain

import "time"

type User struct {
	Id        string    `json:"user_id"`
	Name      string    `json:"username"`
	TeamName  string    `json:"team_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"-"`
}

type Team struct {
	Id        string
	Name      string
	CreatedAt time.Time
}

type TeamMember struct {
	TeamId   string
	UserId   string
	JoinedAt time.Time
}

type Pr struct {
	Id        string
	Name      string
	AuthorId  string
	Status    string
	CreatedAt time.Time
	MergedAt  *time.Time
}

type PrReviewer struct {
	UserId     string
	PrId       string
	AssignedAt time.Time
}
