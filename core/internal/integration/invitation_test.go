package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateInvitation_OwnerSuccess(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	resp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})

	assert.Equal(t, http.StatusCreated, resp.Code)

	var result map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result["id"])
	assert.NotEmpty(t, result["token"])
	assert.Equal(t, "Test Group", result["group_name"])
	assert.Equal(t, "pending", result["status"])
	assert.Equal(t, "member", result["role"])
	assert.NotEmpty(t, result["expires_at"])
}

func TestCreateInvitation_NonOwnerForbidden(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	resp := ts.DoRequestAs(http.MethodPost, "/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	}, secondUserToken)

	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestCreateInvitation_NotGroupMember(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	resp := ts.DoRequestAs(http.MethodPost, "/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	}, secondUserToken)

	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestCreateInvitation_Unauthorized(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	resp := ts.DoRequest(http.MethodPost, "/api/v1/groups/00000000-0000-0000-0000-000000000000/invitations", map[string]interface{}{
		"role": "member",
	}, http.Header{})

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestCreateInvitation_InvalidGroupID(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	resp := ts.Post("/api/v1/groups/invalid-uuid/invitations", map[string]interface{}{
		"role": "member",
	})

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestCreateInvitation_GroupNotFound(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	resp := ts.Post("/api/v1/groups/00000000-0000-0000-0000-000000000000/invitations", map[string]interface{}{
		"role": "member",
	})

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestListInvitations_OwnerSuccess(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})

	resp := ts.Get("/api/v1/groups/" + groupID + "/invitations")

	assert.Equal(t, http.StatusOK, resp.Code)

	var result []map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(result), 1)
}

func TestListInvitations_NonOwnerForbidden(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	resp := ts.DoRequestAs(http.MethodGet, "/api/v1/groups/"+groupID+"/invitations", nil, secondUserToken)

	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestRevokeInvitation_OwnerSuccess(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	invitationID := invitation["id"].(string)

	resp := ts.Delete("/api/v1/invitations/" + invitationID)

	assert.Equal(t, http.StatusNoContent, resp.Code)
}

func TestRevokeInvitation_AlreadyAccepted(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	invitationID := invitation["id"].(string)
	token := invitation["token"].(string)

	acceptResp := ts.DoRequestAs(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, secondUserToken)
	require.Equal(t, http.StatusOK, acceptResp.Code)

	resp := ts.Delete("/api/v1/invitations/" + invitationID)

	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestRevokeInvitation_NotOwner(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	invitationID := invitation["id"].(string)

	resp := ts.DoRequestAs(http.MethodDelete, "/api/v1/invitations/"+invitationID, nil, secondUserToken)

	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestGetInvitationByToken_Success(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	token := invitation["token"].(string)

	resp := ts.Get("/api/v1/invitations/token/" + token)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "Test Group", result["group_name"])
	assert.Equal(t, "pending", result["status"])
}

func TestGetInvitationByToken_InvalidToken(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	resp := ts.Get("/api/v1/invitations/token/invalid-token-xyz")

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestAcceptInvitation_Success(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	token := invitation["token"].(string)

	resp := ts.DoRequestAs(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, secondUserToken)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestAcceptInvitation_GrantsGroupAccess(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	token := invitation["token"].(string)

	beforeAccess := ts.DoRequestAs(http.MethodGet, "/api/v1/groups/"+groupID, nil, secondUserToken)
	assert.Equal(t, http.StatusForbidden, beforeAccess.Code)

	acceptResp := ts.DoRequestAs(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, secondUserToken)
	require.Equal(t, http.StatusOK, acceptResp.Code)

	afterAccess := ts.DoRequestAs(http.MethodGet, "/api/v1/groups/"+groupID, nil, secondUserToken)
	assert.Equal(t, http.StatusOK, afterAccess.Code)
}

func TestAcceptInvitation_GrantsBudgetAccess(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	budgetResp := ts.Post("/api/v1/groups/"+groupID+"/budgets", map[string]interface{}{
		"name":       "Test Budget",
		"start_date": "2024-01-01",
		"end_date":   "2024-01-31",
	})
	require.Equal(t, http.StatusCreated, budgetResp.Code)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	token := invitation["token"].(string)

	beforeAccess := ts.DoRequestAs(http.MethodGet, "/api/v1/groups/"+groupID+"/budgets", nil, secondUserToken)
	assert.Equal(t, http.StatusForbidden, beforeAccess.Code)

	acceptResp := ts.DoRequestAs(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, secondUserToken)
	require.Equal(t, http.StatusOK, acceptResp.Code)

	afterAccess := ts.DoRequestAs(http.MethodGet, "/api/v1/groups/"+groupID+"/budgets", nil, secondUserToken)
	assert.Equal(t, http.StatusOK, afterAccess.Code)
}

func TestAcceptInvitation_Expired(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	invitationID := invitation["id"].(string)
	token := invitation["token"].(string)

	ctx := context.Background()
	_, err = ts.DB.Exec(ctx, 
		`UPDATE group_invitations SET expires_at = $1 WHERE external_id = $2`,
		time.Now().Add(-1*time.Hour), invitationID)
	require.NoError(t, err)

	resp := ts.DoRequestAs(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, secondUserToken)

	assert.Equal(t, http.StatusGone, resp.Code)
}

func TestAcceptInvitation_Revoked(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	invitationID := invitation["id"].(string)
	token := invitation["token"].(string)

	ts.Delete("/api/v1/invitations/" + invitationID)

	resp := ts.DoRequestAs(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, secondUserToken)

	assert.Equal(t, http.StatusGone, resp.Code)
}

func TestAcceptInvitation_AlreadyAccepted(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	secondUserToken, err := ts.CreateSecondUser()
	require.NoError(t, err)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	token := invitation["token"].(string)

	firstAccept := ts.DoRequestAs(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, secondUserToken)
	require.Equal(t, http.StatusOK, firstAccept.Code)

	secondAccept := ts.DoRequestAs(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, secondUserToken)

	assert.Equal(t, http.StatusConflict, secondAccept.Code)
}

func TestAcceptInvitation_AlreadyMember(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	token := invitation["token"].(string)

	resp := ts.Post("/api/v1/invitations/token/" + token + "/accept", nil)

	assert.Equal(t, http.StatusConflict, resp.Code)
}

func TestAcceptInvitation_InvalidToken(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	resp := ts.Post("/api/v1/invitations/token/invalid-token/accept", nil)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestAcceptInvitation_Unauthorized(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.Cleanup(t)
	defer ts.CleanupTestData(t)

	groupResp := ts.Post("/api/v1/groups", map[string]interface{}{
		"name": "Test Group",
	})
	require.Equal(t, http.StatusCreated, groupResp.Code)

	var group map[string]interface{}
	json.Unmarshal(groupResp.Body.Bytes(), &group)
	groupID := group["id"].(string)

	invResp := ts.Post("/api/v1/groups/"+groupID+"/invitations", map[string]interface{}{
		"role": "member",
	})
	require.Equal(t, http.StatusCreated, invResp.Code)

	var invitation map[string]interface{}
	json.Unmarshal(invResp.Body.Bytes(), &invitation)
	token := invitation["token"].(string)

	resp := ts.DoRequest(http.MethodPost, "/api/v1/invitations/token/"+token+"/accept", nil, http.Header{})

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
