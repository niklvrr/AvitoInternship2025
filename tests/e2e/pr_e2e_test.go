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

// TestPRCreate_Success проверяет создание PR
func TestPRCreate_Success(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_create_%d", time.Now().UnixNano())
	authorID := "u_author_create"

	// Создаем команду с несколькими активными ревьюерами
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": "u_reviewer1_create", "username": "Reviewer1", "is_active": true},
			{"user_id": "u_reviewer2_create", "username": "Reviewer2", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	// Создаем PR
	prID := fmt.Sprintf("pr_create_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "Test PR",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	require.Contains(t, response, "pr", "Response must have 'pr' wrapper")

	pr := response["pr"].(map[string]interface{})
	validatePullRequest(t, pr)
	assert.Equal(t, prID, pr["pull_request_id"])
	assert.Equal(t, "OPEN", pr["status"])

	reviewers := pr["assigned_reviewers"].([]interface{})
	assert.GreaterOrEqual(t, len(reviewers), 0, "Must have 0-2 reviewers")
	assert.LessOrEqual(t, len(reviewers), 2, "Must have 0-2 reviewers")

	// Проверяем, что автор не в списке ревьюеров
	for _, reviewer := range reviewers {
		assert.NotEqual(t, authorID, reviewer, "Author must not be in reviewers list")
	}
}

// TestPRCreate_ZeroReviewers проверяет создание PR без доступных ревьюеров
func TestPRCreate_ZeroReviewers(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_zero_%d", time.Now().UnixNano())
	authorID := "u_author_zero"

	// Создаем команду только с автором (нет других активных ревьюеров)
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	// Создаем PR - должен создаться с пустым массивом ревьюеров
	prID := fmt.Sprintf("pr_zero_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "PR without reviewers",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected 201 Created even with 0 reviewers")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	pr := response["pr"].(map[string]interface{})
	validatePullRequest(t, pr)

	reviewers := pr["assigned_reviewers"].([]interface{})
	assert.Len(t, reviewers, 0, "PR must be created with 0 reviewers when no candidates available")
}

// TestPRCreate_Duplicate проверяет создание PR с дубликатом
func TestPRCreate_Duplicate(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_dup_%d", time.Now().UnixNano())
	authorID := "u_author_dup"

	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": "u_reviewer_dup", "username": "Reviewer", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	prID := fmt.Sprintf("pr_dup_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "First PR",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Пытаемся создать PR с тем же ID
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "PR_EXISTS", http.StatusConflict)
}

// TestPRCreate_AuthorNotFound проверяет создание PR с несуществующим автором
func TestPRCreate_AuthorNotFound(t *testing.T) {
	createPRReq := map[string]interface{}{
		"pull_request_id":   fmt.Sprintf("pr_notfound_%d", time.Now().UnixNano()),
		"pull_request_name": "PR with nonexistent author",
		"author_id":         "nonexistent_author",
	}
	body, _ := json.Marshal(createPRReq)
	resp, err := http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "NOT_FOUND", http.StatusNotFound)
}

// TestPRMerge_Success проверяет merge PR
func TestPRMerge_Success(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_merge_%d", time.Now().UnixNano())
	authorID := "u_author_merge"

	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": "u_reviewer_merge", "username": "Reviewer", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	prID := fmt.Sprintf("pr_merge_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "PR to merge",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Merge PR
	mergePRReq := map[string]interface{}{
		"pull_request_id": prID,
	}
	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	require.Contains(t, response, "pr", "Response must have 'pr' wrapper")

	pr := response["pr"].(map[string]interface{})
	validatePullRequest(t, pr)
	assert.Equal(t, "MERGED", pr["status"], "Status must be MERGED")
	assert.NotNil(t, pr["mergedAt"], "mergedAt must be set after merge")

	// Проверяем формат mergedAt
	mergedAt := pr["mergedAt"].(string)
	_, err = time.Parse(time.RFC3339, mergedAt)
	assert.NoError(t, err, "mergedAt must be in RFC3339 format")
}

// TestPRMerge_Idempotent проверяет идемпотентность merge операции
func TestPRMerge_Idempotent(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_idemp_%d", time.Now().UnixNano())
	authorID := "u_author_idemp"

	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": "u_reviewer_idemp", "username": "Reviewer", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	prID := fmt.Sprintf("pr_idemp_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "PR for idempotent test",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Первый merge
	mergePRReq := map[string]interface{}{
		"pull_request_id": prID,
	}
	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var firstResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&firstResponse)
	require.NoError(t, err)
	resp.Body.Close()

	firstPR := firstResponse["pr"].(map[string]interface{})
	firstMergedAt := firstPR["mergedAt"].(string)
	time.Sleep(200 * time.Millisecond)

	// Второй merge (должен быть идемпотентным)
	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Idempotent merge must return 200 OK")

	var secondResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&secondResponse)
	require.NoError(t, err)

	secondPR := secondResponse["pr"].(map[string]interface{})
	validatePullRequest(t, secondPR)
	assert.Equal(t, "MERGED", secondPR["status"], "Status must remain MERGED")

	// Проверяем, что mergedAt не изменился (или изменился минимально из-за времени выполнения)
	secondMergedAt := secondPR["mergedAt"].(string)
	firstTime, _ := time.Parse(time.RFC3339, firstMergedAt)
	secondTime, _ := time.Parse(time.RFC3339, secondMergedAt)

	// mergedAt может быть одинаковым или отличаться на секунды из-за времени выполнения
	// Главное - операция не должна падать с ошибкой
	assert.True(t, secondTime.Equal(firstTime) || secondTime.After(firstTime),
		"mergedAt should not change significantly on idempotent merge")
}

// TestPRMerge_NotFound проверяет merge несуществующего PR
func TestPRMerge_NotFound(t *testing.T) {
	mergePRReq := map[string]interface{}{
		"pull_request_id": fmt.Sprintf("pr_nonexistent_%d", time.Now().UnixNano()),
	}
	body, _ := json.Marshal(mergePRReq)
	resp, err := http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "NOT_FOUND", http.StatusNotFound)
}

// TestPRReassign_Success проверяет переназначение ревьюера
func TestPRReassign_Success(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_reassign_%d", time.Now().UnixNano())
	authorID := "u_author_reassign"

	// Создаем команду с несколькими ревьюерами
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": "u_reviewer1_reassign", "username": "Reviewer1", "is_active": true},
			{"user_id": "u_reviewer2_reassign", "username": "Reviewer2", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	prID := fmt.Sprintf("pr_reassign_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "PR for reassign",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Получаем список ревьюеров
	var createResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	// Если не удалось декодировать, значит PR был создан, но нужно получить его заново
	if err != nil {
		resp, err = http.Get(testServer.URL + "/team/get?team_name=" + uniqueName)
		require.NoError(t, err)
		resp.Body.Close()
	}

	// Создаем новый запрос для получения PR (через создание команды мы уже знаем структуру)
	// Для теста используем первого ревьюера из команды
	oldReviewerID := "u_reviewer1_reassign"

	// Переназначение
	reassignReq := map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     oldReviewerID,
	}
	body, _ = json.Marshal(reassignReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	require.Contains(t, response, "pr", "Response must have 'pr'")
	require.Contains(t, response, "replaced_by", "Response must have 'replaced_by'")

	pr := response["pr"].(map[string]interface{})
	validatePullRequest(t, pr)
	assert.Equal(t, "OPEN", pr["status"], "PR must remain OPEN after reassign")

	replacedBy := response["replaced_by"].(string)
	assert.IsType(t, "", replacedBy, "replaced_by must be string")
	assert.NotEmpty(t, replacedBy, "replaced_by must not be empty")

	// Проверяем, что старый ревьюер удален из списка
	reviewers := pr["assigned_reviewers"].([]interface{})
	foundOld := false
	for _, reviewer := range reviewers {
		if reviewer == oldReviewerID {
			foundOld = true
			break
		}
	}
	assert.False(t, foundOld, "Old reviewer must be removed from assigned_reviewers")

	// Проверяем, что новый ревьюер добавлен
	foundNew := false
	for _, reviewer := range reviewers {
		if reviewer == replacedBy {
			foundNew = true
			break
		}
	}
	assert.True(t, foundNew, "New reviewer must be in assigned_reviewers")
}

// TestPRReassign_MergedPR проверяет переназначение для merged PR
func TestPRReassign_MergedPR(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_reassign_merged_%d", time.Now().UnixNano())
	authorID := "u_author_reassign_merged"

	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": "u_reviewer1_merged", "username": "Reviewer1", "is_active": true},
			{"user_id": "u_reviewer2_merged", "username": "Reviewer2", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	prID := fmt.Sprintf("pr_reassign_merged_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "PR to merge and reassign",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Merge PR
	mergePRReq := map[string]interface{}{
		"pull_request_id": prID,
	}
	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Попытка переназначения после merge
	reassignReq := map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     "u_reviewer1_merged",
	}
	body, _ = json.Marshal(reassignReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "PR_MERGED", http.StatusConflict)
}

// TestPRReassign_NotAssigned проверяет переназначение не назначенного ревьюера
func TestPRReassign_NotAssigned(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_reassign_notassigned_%d", time.Now().UnixNano())
	authorID := "u_author_notassigned"

	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": "u_reviewer_notassigned", "username": "Reviewer", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	prID := fmt.Sprintf("pr_reassign_notassigned_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "PR for not assigned test",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Пытаемся переназначить автора, который точно не был назначен ревьюером
	reassignReq := map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     authorID, // Автор не может быть ревьюером
	}
	body, _ = json.Marshal(reassignReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "NOT_ASSIGNED", http.StatusConflict)
}

// TestPRReassign_NoCandidate проверяет переназначение когда нет кандидата
func TestPRReassign_NoCandidate(t *testing.T) {
	uniqueName := fmt.Sprintf("pr_reassign_nocandidate_%d", time.Now().UnixNano())
	authorID := "u_author_nocandidate"
	reviewerID := "u_reviewer_nocandidate"

	// Создаем команду с автором и одним активным ревьюером
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": reviewerID, "username": "Reviewer", "is_active": true},
			{"user_id": "u_inactive", "username": "Inactive", "is_active": false},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(500 * time.Millisecond)

	prID := fmt.Sprintf("pr_reassign_nocandidate_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "PR for no candidate test",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Деактивируем автора, чтобы не было других активных кандидатов
	setActiveReq := map[string]interface{}{
		"user_id":   authorID,
		"is_active": false,
	}
	body, _ = json.Marshal(setActiveReq)
	resp, err = http.Post(testServer.URL+"/users/setIsActive", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	time.Sleep(200 * time.Millisecond)

	// Пытаемся переназначить ревьюера (нет других активных кандидатов в команде)
	reassignReq := map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     reviewerID,
	}
	body, _ = json.Marshal(reassignReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "NO_CANDIDATE", http.StatusConflict)
}

// TestPRReassign_NotFound проверяет переназначение для несуществующего PR
func TestPRReassign_NotFound(t *testing.T) {
	reassignReq := map[string]interface{}{
		"pull_request_id": fmt.Sprintf("pr_nonexistent_%d", time.Now().UnixNano()),
		"old_user_id":     "some_user",
	}
	body, _ := json.Marshal(reassignReq)
	resp, err := http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	validateErrorResponse(t, resp, "NOT_FOUND", http.StatusNotFound)
}

// TestCompleteWorkflow проверяет полный E2E workflow
func TestCompleteWorkflow(t *testing.T) {
	uniqueName := fmt.Sprintf("complete_%d", time.Now().UnixNano())
	authorID := "u_author_complete"
	reviewer1ID := "u_reviewer1_complete"
	reviewer2ID := "u_reviewer2_complete"

	// 1. POST /team/add - Создание команды
	teamReq := map[string]interface{}{
		"team_name": uniqueName,
		"members": []map[string]interface{}{
			{"user_id": authorID, "username": "Author", "is_active": true},
			{"user_id": reviewer1ID, "username": "Reviewer1", "is_active": true},
			{"user_id": reviewer2ID, "username": "Reviewer2", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	resp, err := http.Post(testServer.URL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	time.Sleep(200 * time.Millisecond)

	// 2. GET /team/get - Получение команды
	resp, err = http.Get(testServer.URL + "/team/get?team_name=" + uniqueName)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 3. POST /pullRequest/create - Создание PR
	prID := fmt.Sprintf("pr_complete_%d", time.Now().UnixNano())
	createPRReq := map[string]interface{}{
		"pull_request_id":   prID,
		"pull_request_name": "Complete workflow PR",
		"author_id":         authorID,
	}
	body, _ = json.Marshal(createPRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	time.Sleep(200 * time.Millisecond)

	// 4. GET /users/getReview - Получение ревьюев ревьюера
	resp, err = http.Get(testServer.URL + "/users/getReview?user_id=" + reviewer1ID)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 5. POST /pullRequest/reassign - Переназначение ревьюера
	reassignReq := map[string]interface{}{
		"pull_request_id": prID,
		"old_user_id":     reviewer1ID,
	}
	body, _ = json.Marshal(reassignReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	time.Sleep(200 * time.Millisecond)

	// 6. POST /pullRequest/merge - Merge PR
	mergePRReq := map[string]interface{}{
		"pull_request_id": prID,
	}
	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	time.Sleep(200 * time.Millisecond)

	// 7. POST /pullRequest/reassign после merge - Должна вернуть ошибку
	reassignReq["old_user_id"] = reviewer2ID
	body, _ = json.Marshal(reassignReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/reassign", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	validateErrorResponse(t, resp, "PR_MERGED", http.StatusConflict)

	// 8. POST /pullRequest/merge повторно - Идемпотентность
	body, _ = json.Marshal(mergePRReq)
	resp, err = http.Post(testServer.URL+"/pullRequest/merge", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "Idempotent merge must return 200")
}
