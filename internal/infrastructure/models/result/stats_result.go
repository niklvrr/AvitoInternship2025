package result

type UserStats struct {
	UserId      string
	Username    string
	Assignments int
}

type PrStats struct {
	PrId            string
	PrName          string
	ReviewersCount  int
}

type StatsResult struct {
	Users []UserStats
	PRs   []PrStats
}

