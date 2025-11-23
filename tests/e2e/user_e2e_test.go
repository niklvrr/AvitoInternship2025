package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_SetIsActive_Success(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-setactive",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-setactive", "username": "SetActiveUser", "is_active": true},
		},
	}

	createResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	reqBody := map[string]interface{}{
		"user_id":   "e2e-u-setactive",
		"is_active": false,
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/users/setIsActive", reqBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "user")
	user, ok := result["user"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "e2e-u-setactive", user["user_id"])
	assert.Equal(t, "SetActiveUser", user["username"])
	assert.Equal(t, "e2e-team-setactive", user["team_name"])
	assert.Equal(t, false, user["is_active"])
}

func TestUser_SetIsActive_NotFound(t *testing.T) {
	reqBody := map[string]interface{}{
		"user_id":   "e2e-nonexistent-user",
		"is_active": true,
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/users/setIsActive", reqBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errorObj["code"])
}

func TestUser_SetIsActive_Activate(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-activate",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-activate", "username": "ActivateUser", "is_active": false},
		},
	}

	createResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	reqBody := map[string]interface{}{
		"user_id":   "e2e-u-activate",
		"is_active": true,
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/users/setIsActive", reqBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	user, ok := result["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, user["is_active"])
}

func TestUser_GetReview_Success(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-review",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-author", "username": "Author", "is_active": true},
			{"user_id": "e2e-u-reviewer", "username": "Reviewer", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	prReq := map[string]interface{}{
		"pull_request_id":   "e2e-pr-review",
		"pull_request_name": "Review PR",
		"author_id":         "e2e-u-author",
	}

	createPrResp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", prReq)
	createPrResp.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	resp := makeRequest(t, http.MethodGet, baseURL+"/users/getReview?user_id=e2e-u-reviewer", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "e2e-u-reviewer", result["user_id"])
	assert.Contains(t, result, "pull_requests")

	pullRequests, ok := result["pull_requests"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(pullRequests), 1)

	found := false
	for _, pr := range pullRequests {
		prMap, ok := pr.(map[string]interface{})
		require.True(t, ok)
		if prMap["pull_request_id"] == "e2e-pr-review" {
			found = true
			assert.Equal(t, "Review PR", prMap["pull_request_name"])
			assert.Equal(t, "e2e-u-author", prMap["author_id"])
			assert.Contains(t, []interface{}{"OPEN", "MERGED"}, prMap["status"])

			assert.Contains(t, prMap, "pull_request_id")
			assert.Contains(t, prMap, "pull_request_name")
			assert.Contains(t, prMap, "author_id")
			assert.Contains(t, prMap, "status")

			assert.NotContains(t, prMap, "assigned_reviewers")
			assert.NotContains(t, prMap, "createdAt")
			assert.NotContains(t, prMap, "mergedAt")
			break
		}
	}
	assert.True(t, found, "PR should be found in reviewer's list")
}

func TestUser_GetReview_EmptyList(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-empty-review",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-no-reviews", "username": "NoReviews", "is_active": true},
		},
	}

	createResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	resp := makeRequest(t, http.MethodGet, baseURL+"/users/getReview?user_id=e2e-u-no-reviews", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "e2e-u-no-reviews", result["user_id"])
	pullRequests, ok := result["pull_requests"].([]interface{})
	require.True(t, ok)
	assert.Len(t, pullRequests, 0)
}

func TestUser_GetReview_NotFound(t *testing.T) {
	resp := makeRequest(t, http.MethodGet, baseURL+"/users/getReview?user_id=e2e-nonexistent-reviewer", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errorObj["code"])
}

func TestUser_GetReview_MissingUserId(t *testing.T) {
	resp := makeRequest(t, http.MethodGet, baseURL+"/users/getReview", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
