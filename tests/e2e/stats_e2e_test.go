package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStats_GetStats_Success(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-stats",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-stats-1", "username": "StatsUser1", "is_active": true},
			{"user_id": "e2e-u-stats-2", "username": "StatsUser2", "is_active": true},
			{"user_id": "e2e-u-stats-3", "username": "StatsUser3", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	prReq1 := map[string]interface{}{
		"pull_request_id":   "e2e-pr-stats-1",
		"pull_request_name": "Stats PR 1",
		"author_id":         "e2e-u-stats-1",
	}

	createPrResp1 := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", prReq1)
	createPrResp1.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp1.StatusCode)

	time.Sleep(100 * time.Millisecond)

	prReq2 := map[string]interface{}{
		"pull_request_id":   "e2e-pr-stats-2",
		"pull_request_name": "Stats PR 2",
		"author_id":         "e2e-u-stats-2",
	}

	createPrResp2 := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", prReq2)
	createPrResp2.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp2.StatusCode)

	time.Sleep(200 * time.Millisecond)

	resp := makeRequest(t, http.MethodGet, baseURL+"/stats", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "users")
	assert.Contains(t, result, "prs")

	users, ok := result["users"].([]interface{})
	require.True(t, ok, "users should be an array")

	prs, ok := result["prs"].([]interface{})
	require.True(t, ok, "prs should be an array")

	assert.GreaterOrEqual(t, len(users), 3)
	assert.GreaterOrEqual(t, len(prs), 2)
}

func TestStats_GetStats_EmptyDatabase(t *testing.T) {
	resp := makeRequest(t, http.MethodGet, baseURL+"/stats", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "users")
	assert.Contains(t, result, "prs")

	users, ok := result["users"].([]interface{})
	require.True(t, ok)
	prs, ok := result["prs"].([]interface{})
	require.True(t, ok)

	assert.NotNil(t, users)
	assert.NotNil(t, prs)
}

