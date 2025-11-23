package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	baseURL      = "http://localhost:8080"
	targetRPS    = 5
	testDuration = 2 * time.Minute
)

var rng *rand.Rand

type TeamRequest struct {
	TeamName string   `json:"team_name"`
	Members  []Member `json:"members"`
}

type Member struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type CreatePrRequest struct {
	PrID     string `json:"pull_request_id"`
	PrName   string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type MergePrRequest struct {
	PrID string `json:"pull_request_id"`
}

type ReassignPrRequest struct {
	PrID      string `json:"pull_request_id"`
	OldUserID string `json:"old_user_id"`
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run load_test.go <scenario>")
		fmt.Println("Scenarios: health, team, user, pr, all")
		os.Exit(1)
	}

	scenario := os.Args[1]
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	var metrics vegeta.Metrics
	var err error

	switch scenario {
	case "health":
		metrics, err = testHealth()
	case "team":
		metrics, err = testTeam()
	case "user":
		metrics, err = testUser()
	case "pr":
		metrics, err = testPR()
	case "all":
		metrics, err = testAll()
	default:
		fmt.Printf("Unknown scenario: %s\n", scenario)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printMetrics(metrics)
}

func testHealth() (vegeta.Metrics, error) {
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    baseURL + "/health",
	})

	return runAttack(targeter, "Health Check")
}

func testTeam() (vegeta.Metrics, error) {
	teamID := rng.Intn(10000)
	teamName := fmt.Sprintf("load_team_%d", teamID)

	targeter := vegeta.NewStaticTargeter(
		vegeta.Target{
			Method: "POST",
			URL:    baseURL + "/team/add",
			Body:   createTeamBody(teamName),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
		vegeta.Target{
			Method: "GET",
			URL:    baseURL + "/team/get?team_name=" + teamName,
		},
	)

	return runAttack(targeter, "Team Operations")
}

func testUser() (vegeta.Metrics, error) {
	userID := fmt.Sprintf("u_load_%d", rng.Intn(10000))

	targeter := vegeta.NewStaticTargeter(
		vegeta.Target{
			Method: "POST",
			URL:    baseURL + "/users/setIsActive",
			Body:   createSetIsActiveBody(userID, false),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
		vegeta.Target{
			Method: "GET",
			URL:    baseURL + "/users/getReview?user_id=" + userID,
		},
	)

	return runAttack(targeter, "User Operations")
}

func testPR() (vegeta.Metrics, error) {
	prID := fmt.Sprintf("pr_load_%d", rng.Intn(10000))
	authorID := fmt.Sprintf("u_load_%d", rng.Intn(10000))

	targeter := vegeta.NewStaticTargeter(
		vegeta.Target{
			Method: "POST",
			URL:    baseURL + "/pullRequest/create",
			Body:   createPRBody(prID, "Load Test PR", authorID),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
		vegeta.Target{
			Method: "POST",
			URL:    baseURL + "/pullRequest/merge",
			Body:   createMergePRBody(prID),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
	)

	return runAttack(targeter, "PR Operations")
}

func testAll() (vegeta.Metrics, error) {
	teamID := rng.Intn(10000)
	teamName := fmt.Sprintf("load_team_%d", teamID)
	userID1 := fmt.Sprintf("u_load_%d_1", teamID)
	userID2 := fmt.Sprintf("u_load_%d_2", teamID)
	prID := fmt.Sprintf("pr_load_%d", teamID)

	targeter := vegeta.NewStaticTargeter(
		vegeta.Target{
			Method: "GET",
			URL:    baseURL + "/health",
		},
		vegeta.Target{
			Method: "POST",
			URL:    baseURL + "/team/add",
			Body:   createTeamBodyWithUsers(teamName, userID1, userID2),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
		vegeta.Target{
			Method: "GET",
			URL:    baseURL + "/team/get?team_name=" + teamName,
		},
		vegeta.Target{
			Method: "POST",
			URL:    baseURL + "/pullRequest/create",
			Body:   createPRBody(prID, "Load Test PR", userID1),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
		vegeta.Target{
			Method: "POST",
			URL:    baseURL + "/pullRequest/merge",
			Body:   createMergePRBody(prID),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
	)

	return runAttack(targeter, "All Endpoints")
}

func runAttack(targeter vegeta.Targeter, name string) (vegeta.Metrics, error) {
	rate := vegeta.Rate{Freq: targetRPS, Per: time.Second}
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, testDuration, name) {
		metrics.Add(res)
	}
	metrics.Close()

	return metrics, nil
}

func createTeamBody(teamName string) []byte {
	req := TeamRequest{
		TeamName: teamName,
		Members: []Member{
			{UserID: "u1_" + teamName, Username: "Alice", IsActive: true},
			{UserID: "u2_" + teamName, Username: "Bob", IsActive: true},
		},
	}
	body, _ := json.Marshal(req)
	return body
}

func createTeamBodyWithUsers(teamName, userID1, userID2 string) []byte {
	req := TeamRequest{
		TeamName: teamName,
		Members: []Member{
			{UserID: userID1, Username: "Alice", IsActive: true},
			{UserID: userID2, Username: "Bob", IsActive: true},
		},
	}
	body, _ := json.Marshal(req)
	return body
}

func createPRBody(prID, prName, authorID string) []byte {
	req := CreatePrRequest{
		PrID:     prID,
		PrName:   prName,
		AuthorID: authorID,
	}
	body, _ := json.Marshal(req)
	return body
}

func createMergePRBody(prID string) []byte {
	req := MergePrRequest{PrID: prID}
	body, _ := json.Marshal(req)
	return body
}

func createSetIsActiveBody(userID string, isActive bool) []byte {
	req := SetIsActiveRequest{
		UserID:   userID,
		IsActive: isActive,
	}
	body, _ := json.Marshal(req)
	return body
}

func printMetrics(metrics vegeta.Metrics) {
	fmt.Printf("\n=== Load Test Results ===\n\n")
	fmt.Printf("Requests Total:     %d\n", metrics.Requests)
	fmt.Printf("Success Rate:       %.2f%%\n", metrics.Success*100)
	fmt.Printf("Duration:           %v\n", metrics.Duration)

	if metrics.Requests > 0 {
		fmt.Printf("\nLatency:\n")
		fmt.Printf("  Mean:             %v\n", metrics.Latencies.Mean)
		fmt.Printf("  P50:              %v\n", metrics.Latencies.P50)
		fmt.Printf("  P95:              %v\n", metrics.Latencies.P95)
		fmt.Printf("  P99:              %v\n", metrics.Latencies.P99)
		fmt.Printf("  Max:              %v\n", metrics.Latencies.Max)

		fmt.Printf("\nThroughput:\n")
		fmt.Printf("  Requests/sec:     %.2f\n", metrics.Rate)

		fmt.Printf("\nStatus Codes:\n")
		for code, count := range metrics.StatusCodes {
			fmt.Printf("  %s: %d\n", code, count)
		}

		fmt.Printf("\nErrors:\n")
		if len(metrics.Errors) > 0 {
			for _, err := range metrics.Errors {
				fmt.Printf("  %s\n", err)
			}
		} else {
			fmt.Printf("  None\n")
		}

		fmt.Printf("\nSLI Compliance:\n")
		p95ms := metrics.Latencies.P95.Seconds() * 1000
		successRate := metrics.Success * 100
		fmt.Printf("  P95 Latency:      %.2f ms (target: < 300ms) - %s\n",
			p95ms,
			checkStatus(p95ms < 300, "PASS", "FAIL"))
		fmt.Printf("  Success Rate:     %.2f%% (target: > 99.9%%) - %s\n",
			successRate,
			checkStatus(successRate >= 99.9, "PASS", "FAIL"))
	}
	fmt.Printf("\n")
}

func checkStatus(condition bool, pass, fail string) string {
	if condition {
		return pass
	}
	return fail
}
