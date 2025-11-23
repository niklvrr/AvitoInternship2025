package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPR_Create_Success(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-pr-create",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-author", "username": "PRAuthor", "is_active": true},
			{"user_id": "e2e-u-pr-reviewer1", "username": "PRReviewer1", "is_active": true},
			{"user_id": "e2e-u-pr-reviewer2", "username": "PRReviewer2", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	reqBody := map[string]interface{}{
		"pull_request_id":   "e2e-pr-create-1",
		"pull_request_name": "Create PR Test",
		"author_id":         "e2e-u-pr-author",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", reqBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "pr")
	pr, ok := result["pr"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "e2e-pr-create-1", pr["pull_request_id"])
	assert.Equal(t, "Create PR Test", pr["pull_request_name"])
	assert.Equal(t, "e2e-u-pr-author", pr["author_id"])
	assert.Equal(t, "OPEN", pr["status"])

	assignedReviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok)
	assert.GreaterOrEqual(t, len(assignedReviewers), 0)
	assert.LessOrEqual(t, len(assignedReviewers), 2)

	for _, reviewer := range assignedReviewers {
		reviewerID, ok := reviewer.(string)
		require.True(t, ok)
		assert.NotEqual(t, "e2e-u-pr-author", reviewerID)
		assert.Contains(t, []string{"e2e-u-pr-reviewer1", "e2e-u-pr-reviewer2"}, reviewerID)
	}

	assert.Contains(t, pr, "createdAt")
	createdAt, ok := pr["createdAt"].(string)
	if ok {
		assert.NotEmpty(t, createdAt)
	}
}

func TestPR_Create_ZeroReviewers(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-pr-zero",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-zero-author", "username": "ZeroAuthor", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	reqBody := map[string]interface{}{
		"pull_request_id":   "e2e-pr-zero",
		"pull_request_name": "Zero Reviewers PR",
		"author_id":         "e2e-u-pr-zero-author",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", reqBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	pr, ok := result["pr"].(map[string]interface{})
	require.True(t, ok)

	assignedReviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, assignedReviewers, 0)
}

func TestPR_Create_Duplicate(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-pr-duplicate",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-dup-author", "username": "DupAuthor", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	reqBody := map[string]interface{}{
		"pull_request_id":   "e2e-pr-duplicate",
		"pull_request_name": "Duplicate PR",
		"author_id":         "e2e-u-pr-dup-author",
	}

	resp1 := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", reqBody)
	resp1.Body.Close()
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	time.Sleep(100 * time.Millisecond)

	resp2 := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", reqBody)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)

	errorResp := parseErrorResponse(t, resp2)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "PR_EXISTS", errorObj["code"])
}

func TestPR_Create_AuthorNotFound(t *testing.T) {
	reqBody := map[string]interface{}{
		"pull_request_id":   "e2e-pr-notfound",
		"pull_request_name": "Author Not Found PR",
		"author_id":         "e2e-nonexistent-author",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", reqBody)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errorObj["code"])
}

func TestPR_Merge_Success(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-pr-merge",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-merge-author", "username": "MergeAuthor", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	createPrReq := map[string]interface{}{
		"pull_request_id":   "e2e-pr-merge",
		"pull_request_name": "Merge PR",
		"author_id":         "e2e-u-pr-merge-author",
	}

	createPrResp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", createPrReq)
	createPrResp.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	mergeReq := map[string]interface{}{
		"pull_request_id": "e2e-pr-merge",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/merge", mergeReq)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	pr, ok := result["pr"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "e2e-pr-merge", pr["pull_request_id"])
	assert.Equal(t, "MERGED", pr["status"])

	assert.Contains(t, pr, "createdAt")
	createdAt, ok := pr["createdAt"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, createdAt)

	assert.Contains(t, pr, "mergedAt")
	mergedAt, ok := pr["mergedAt"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, mergedAt)
}

func TestPR_Merge_Idempotent(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-pr-idempotent",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-idempotent-author", "username": "IdempotentAuthor", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	createPrReq := map[string]interface{}{
		"pull_request_id":   "e2e-pr-idempotent",
		"pull_request_name": "Idempotent PR",
		"author_id":         "e2e-u-pr-idempotent-author",
	}

	createPrResp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", createPrReq)
	createPrResp.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	mergeReq := map[string]interface{}{
		"pull_request_id": "e2e-pr-idempotent",
	}

	resp1 := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/merge", mergeReq)
	var result1 map[string]interface{}
	parseJSONResponse(t, resp1, &result1)
	resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	pr1, ok := result1["pr"].(map[string]interface{})
	require.True(t, ok)
	firstMergedAt := pr1["mergedAt"]

	time.Sleep(100 * time.Millisecond)

	resp2 := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/merge", mergeReq)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var result2 map[string]interface{}
	err := json.NewDecoder(resp2.Body).Decode(&result2)
	require.NoError(t, err)

	pr2, ok := result2["pr"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "MERGED", pr2["status"])
	assert.Equal(t, firstMergedAt, pr2["mergedAt"])
}

func TestPR_Merge_NotFound(t *testing.T) {
	mergeReq := map[string]interface{}{
		"pull_request_id": "e2e-nonexistent-pr",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/merge", mergeReq)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errorObj["code"])
}

func TestPR_Reassign_Success(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-pr-reassign",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-reassign-author", "username": "ReassignAuthor", "is_active": true},
			{"user_id": "e2e-u-pr-reassign-old", "username": "OldReviewer", "is_active": true},
			{"user_id": "e2e-u-pr-reassign-new", "username": "NewReviewer", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	createPrReq := map[string]interface{}{
		"pull_request_id":   "e2e-pr-reassign",
		"pull_request_name": "Reassign PR",
		"author_id":         "e2e-u-pr-reassign-author",
	}

	createPrResp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", createPrReq)
	createPrResp.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp.StatusCode)

	time.Sleep(200 * time.Millisecond)

	reassignReq := map[string]interface{}{
		"pull_request_id": "e2e-pr-reassign",
		"old_user_id":     "e2e-u-pr-reassign-old",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/reassign", reassignReq)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Contains(t, result, "pr")
	assert.Contains(t, result, "replaced_by")

	pr, ok := result["pr"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "e2e-pr-reassign", pr["pull_request_id"])
	assert.Equal(t, "OPEN", pr["status"])

	assert.Contains(t, pr, "createdAt")
	createdAt, ok := pr["createdAt"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, createdAt)

	replacedBy, ok := result["replaced_by"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, replacedBy)
	assert.NotEqual(t, "e2e-u-pr-reassign-old", replacedBy)

	assignedReviewers, ok := pr["assigned_reviewers"].([]interface{})
	require.True(t, ok)
	foundReplacedBy := false
	for _, reviewer := range assignedReviewers {
		if reviewer == replacedBy {
			foundReplacedBy = true
			break
		}
	}
	assert.True(t, foundReplacedBy, "replaced_by should be in assigned_reviewers")
}

func TestPR_Reassign_MergedPR(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-pr-reassign-merged",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-merged-author", "username": "MergedAuthor", "is_active": true},
			{"user_id": "e2e-u-pr-merged-reviewer", "username": "MergedReviewer", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	createPrReq := map[string]interface{}{
		"pull_request_id":   "e2e-pr-reassign-merged",
		"pull_request_name": "Merged Reassign PR",
		"author_id":         "e2e-u-pr-merged-author",
	}

	createPrResp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", createPrReq)
	createPrResp.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	mergeReq := map[string]interface{}{
		"pull_request_id": "e2e-pr-reassign-merged",
	}

	mergeResp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/merge", mergeReq)
	mergeResp.Body.Close()
	require.Equal(t, http.StatusOK, mergeResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	reassignReq := map[string]interface{}{
		"pull_request_id": "e2e-pr-reassign-merged",
		"old_user_id":     "e2e-u-pr-merged-reviewer",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/reassign", reassignReq)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "PR_MERGED", errorObj["code"])
	assert.Contains(t, errorObj["message"], "cannot reassign on merged PR")
}

func TestPR_Reassign_NotAssigned(t *testing.T) {
	teamReq1 := map[string]interface{}{
		"team_name": "e2e-team-pr-notassigned-1",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-notassigned-author", "username": "NotAssignedAuthor", "is_active": true},
			{"user_id": "e2e-u-pr-notassigned-reviewer", "username": "NotAssignedReviewer", "is_active": true},
		},
	}

	createTeamResp1 := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq1)
	createTeamResp1.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp1.StatusCode)

	teamReq2 := map[string]interface{}{
		"team_name": "e2e-team-pr-notassigned-2",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-notassigned-other", "username": "OtherUser", "is_active": true},
		},
	}

	createTeamResp2 := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq2)
	createTeamResp2.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp2.StatusCode)

	time.Sleep(100 * time.Millisecond)

	createPrReq := map[string]interface{}{
		"pull_request_id":   "e2e-pr-notassigned",
		"pull_request_name": "Not Assigned PR",
		"author_id":         "e2e-u-pr-notassigned-author",
	}

	createPrResp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", createPrReq)
	createPrResp.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp.StatusCode)

	time.Sleep(200 * time.Millisecond)

	reassignReq := map[string]interface{}{
		"pull_request_id": "e2e-pr-notassigned",
		"old_user_id":     "e2e-u-pr-notassigned-other",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/reassign", reassignReq)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_ASSIGNED", errorObj["code"])
	assert.Contains(t, errorObj["message"], "reviewer is not assigned to this PR")
}

func TestPR_Reassign_NoCandidate(t *testing.T) {
	teamReq := map[string]interface{}{
		"team_name": "e2e-team-pr-nocandidate",
		"members": []map[string]interface{}{
			{"user_id": "e2e-u-pr-nocandidate-author", "username": "NoCandidateAuthor", "is_active": true},
			{"user_id": "e2e-u-pr-nocandidate-reviewer", "username": "NoCandidateReviewer", "is_active": true},
		},
	}

	createTeamResp := makeRequest(t, http.MethodPost, baseURL+"/team/add", teamReq)
	createTeamResp.Body.Close()
	require.Equal(t, http.StatusCreated, createTeamResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	createPrReq := map[string]interface{}{
		"pull_request_id":   "e2e-pr-nocandidate",
		"pull_request_name": "No Candidate PR",
		"author_id":         "e2e-u-pr-nocandidate-author",
	}

	createPrResp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/create", createPrReq)
	createPrResp.Body.Close()
	require.Equal(t, http.StatusCreated, createPrResp.StatusCode)

	time.Sleep(200 * time.Millisecond)

	setInactiveReq := map[string]interface{}{
		"user_id":   "e2e-u-pr-nocandidate-reviewer",
		"is_active": false,
	}

	setInactiveResp := makeRequest(t, http.MethodPost, baseURL+"/users/setIsActive", setInactiveReq)
	setInactiveResp.Body.Close()
	require.Equal(t, http.StatusOK, setInactiveResp.StatusCode)

	time.Sleep(100 * time.Millisecond)

	reassignReq := map[string]interface{}{
		"pull_request_id": "e2e-pr-nocandidate",
		"old_user_id":     "e2e-u-pr-nocandidate-reviewer",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/reassign", reassignReq)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NO_CANDIDATE", errorObj["code"])
	assert.Contains(t, errorObj["message"], "no active replacement candidate in team")
}

func TestPR_Reassign_NotFound(t *testing.T) {
	reassignReq := map[string]interface{}{
		"pull_request_id": "e2e-nonexistent-pr",
		"old_user_id":     "e2e-nonexistent-user",
	}

	resp := makeRequest(t, http.MethodPost, baseURL+"/pullRequest/reassign", reassignReq)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	errorResp := parseErrorResponse(t, resp)
	assert.Contains(t, errorResp, "error")

	errorObj, ok := errorResp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errorObj["code"])
}
