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

// TestTeamAdd_Success проверяет создание команды
func TestTeamAdd_Success(t *testing.T) {
	uniqueName := fmt.Sprintf("team_add_%d", time.Now().UnixNano())

	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": "u1_add", "username": "Alice", "is_active": true},
			{"user_id": "u2_add", "username": "Bob", "is_active": true},
		},
	}

	body, err := json.Marshal(teamReq)
	require.NoError(t, err)

	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	require.Contains(t, response, "team", "Response must have 'team' wrapper")

	team := response["team"].(map[string]interface{})
	validateTeam(t, team)
	assert.Equal(t, uniqueName, team["team_name"])

	time.Sleep(200 * time.Millisecond)
}

// TestTeamAdd_Duplicate проверяет создание команды с дубликатом
func TestTeamAdd_Duplicate(t *testing.T) {
	uniqueName := fmt.Sprintf("team_dup_%d", time.Now().UnixNano())

	// Создаем команду первый раз
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members":   []map[string]interface{}{},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	time.Sleep(200 * time.Millisecond)

	// Пытаемся создать еще раз
	body, _ = json.Marshal(teamReq)
	resp, err = http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "TEAM_EXISTS", http.StatusBadRequest)
}

// TestTeamGet_Success проверяет получение команды
func TestTeamGet_Success(t *testing.T) {
	uniqueName := fmt.Sprintf("team_get_%d", time.Now().UnixNano())

	// Создаем команду
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": "u1_get", "username": "Alice", "is_active": true},
			{"user_id": "u2_get", "username": "Bob", "is_active": false},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Получаем команду
	resp, err = http.Get(testServer.URL + "/team/get?team_name=" + uniqueName)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var team map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&team)
	require.NoError(t, err)

	validateTeam(t, team)
	assert.Equal(t, uniqueName, team["team_name"])

	members := team["members"].([]interface{})
	assert.Len(t, members, 2, "Team must have 2 members")
}

// TestTeamGet_NotFound проверяет получение несуществующей команды
func TestTeamGet_NotFound(t *testing.T) {
	nonexistentName := fmt.Sprintf("nonexistent_%d", time.Now().UnixNano())

	resp, err := http.Get(testServer.URL + "/team/get?team_name=" + nonexistentName)
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "NOT_FOUND", http.StatusNotFound)
}

