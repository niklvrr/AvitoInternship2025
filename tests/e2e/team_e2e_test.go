package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_Add_Success(t *testing.T) {
	reqBody := map[string]interface{}{
		"team_name": "e2e-team-1",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u1", "username": "Alice", "is_active": true},
			{"user_id": "e2e-u2", "username": "Bob", "is_active": true},
		},
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/team/add", reqBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "team")
	team, ok := result["team"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "e2e-team-1", team["team_name"])
	assert.Contains(t, team, "members")

	members, ok := team["members"].([]interface{})
	require.True(t, ok)
	assert.Len(t, members, 2)

	member1, ok := members[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "e2e-u1", member1["user_id"])
	assert.Equal(t, "Alice", member1["username"])
	assert.Equal(t, true, member1["is_active"])

	member2, ok := members[1].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "e2e-u2", member2["user_id"])
	assert.Equal(t, "Bob", member2["username"])
	assert.Equal(t, true, member2["is_active"])
}

func TestTeam_Add_DuplicateTeam(t *testing.T) {
	reqBody := map[string]interface{}{
		"team_name": "e2e-team-duplicate",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u3", "username": "Charlie", "is_active": true},
		},
	}

	resp1 := makeRequest(t, http.MethodPost, baseURL+"/team/add", reqBody)
	resp1.Body.Close()
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	time.Sleep(100 * time.Millisecond)

	resp2 := makeRequest(t, http.MethodPost, baseURL+"/team/add", reqBody)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	errorResp := parseErrorResponse(t, resp2)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "TEAM_EXISTS", errorObj["code"])
	assert.Contains(t, errorObj["message"], "team_name already exists")
}

func TestTeam_Add_EmptyMembers(t *testing.T) {
	reqBody := map[string]interface{}{
		"team_name": "e2e-team-empty",
		"members":   []interface{}{},
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/team/add", reqBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "team")
	team, ok := result["team"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "e2e-team-empty", team["team_name"])

	members, ok := team["members"].([]interface{})
	require.True(t, ok)
	assert.Len(t, members, 0)
}

func TestTeam_Get_Success(t *testing.T) {
	reqBody := map[string]interface{}{
		"team_name": "e2e-team-get",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u4", "username": "David", "is_active": true},
			{"user_id": "e2e-u5", "username": "Eve", "is_active": false},
		},
	}

	createResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", reqBody)
	createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	resp := makeRequest(t, http.MethodGet, baseURL+"/team/get?team_name=e2e-team-get", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "e2e-team-get", result["team_name"])
	assert.Contains(t, result, "members")

	members, ok := result["members"].([]interface{})
	require.True(t, ok)
	assert.Len(t, members, 2)

	memberMap := make(map[string]map[string]interface{})
	for _, m := range members {
		member, ok := m.(map[string]interface{})
		require.True(t, ok)
		userID, ok := member["user_id"].(string)
		require.True(t, ok)
		memberMap[userID] = member
	}

	assert.Equal(t, "David", memberMap["e2e-u4"]["username"])
	assert.Equal(t, true, memberMap["e2e-u4"]["is_active"])
	assert.Equal(t, "Eve", memberMap["e2e-u5"]["username"])
	assert.Equal(t, false, memberMap["e2e-u5"]["is_active"])
}

func TestTeam_Get_NotFound(t *testing.T) {
	resp := makeRequest(t, http.MethodGet, baseURL+"/team/get?team_name=e2e-nonexistent-team", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errorObj["code"])
}

func TestTeam_Get_MissingTeamName(t *testing.T) {
	resp := makeRequest(t, http.MethodGet, baseURL+"/team/get", nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
