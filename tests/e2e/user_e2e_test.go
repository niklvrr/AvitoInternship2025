package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserSetIsActive_Success проверяет установку флага активности пользователя
func TestUserSetIsActive_Success(t *testing.T) {
	uniqueName := fmt.Sprintf("user_setactive_%d", time.Now().UnixNano())
	userID := "u_setactive"

	// Создаем команду с пользователем
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": userID, "username": "TestUser", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	// Деактивируем пользователя
	setActiveReq := map[string]interface{}{
		"user_id":   userID,
		"is_active": false,
	}
	body, _ = json.Marshal(setActiveReq)
	resp, err = http.Post(testServer.URL+"/users/setIsActive", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	require.Contains(t, response, "user", "Response must have 'user' wrapper")

	user := response["user"].(map[string]interface{})
	validateUser(t, user)
	assert.Equal(t, userID, user["user_id"])
	assert.Equal(t, false, user["is_active"])
}

// TestUserSetIsActive_NotFound проверяет установку флага для несуществующего пользователя
func TestUserSetIsActive_NotFound(t *testing.T) {
	setActiveReq := map[string]interface{}{
		"user_id":   "nonexistent_user",
		"is_active": true,
	}
	body, _ := json.Marshal(setActiveReq)
	resp, err := http.Post(testServer.URL+"/users/setIsActive", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "NOT_FOUND", http.StatusNotFound)
}

// TestUserGetReview_Success проверяет получение ревьюев пользователя
func TestUserGetReview_Success(t *testing.T) {
	uniqueName := fmt.Sprintf("user_getreview_%d", time.Now().UnixNano())
	authorID := "u_author_getreview"
	reviewerID := "u_reviewer_getreview"

	// Создаем команду
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": reviewerID, "username": "Reviewer", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	// Создаем PR
	prID := fmt.Sprintf("pr_getreview_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "PR for get review",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Получаем ревьюи ревьюера
	resp, err = http.Get(testServer.URL + "/users/getReview?user_id=" + reviewerID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	require.Contains(t, response, "user_id", "Response must have user_id")
	require.Contains(t, response, "pull_requests", "Response must have pull_requests")

	assert.Equal(t, reviewerID, response["user_id"], "user_id must match")
	assert.IsType(t, []interface{}{}, response["pull_requests"], "pull_requests must be array")

	pullRequests := response["pull_requests"].([]interface{})
	for _, prRaw := range pullRequests {
		pr := prRaw.(map[string]interface{})
		validatePullRequestShort(t, pr)
	}
}
