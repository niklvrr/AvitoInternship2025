package response

type UserStat struct {
	UserId      string `json:"user_id"`
	Username    string `json:"username"`
	Assignments int    `json:"assignments"`
}

type PrStat struct {
	PrId           string `json:"pr_id"`
	PrName         string `json:"pr_name"`
	ReviewersCount int    `json:"reviewers_count"`
}

type StatsResponse struct {
	Users []UserStat `json:"users"`
	PRs   []PrStat   `json:"prs"`
}

