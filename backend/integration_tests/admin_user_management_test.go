package integration_tests

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func setupAdminForUserMgmt(t *testing.T) (AuthContext, uuid.UUID) {
	t.Helper()

	adminResp, status := InitializeAdmin(
		t,
		"Mgmt Admin",
		"mgmt-admin@example.com",
		"StrongP@ssw0rd123",
		"Mgmt Workspace",
		"User management test workspace",
		"",
	)
	if status != http.StatusCreated {
		t.Fatalf("failed to init admin: status %d", status)
	}

	auth := AuthContext{
		UserID:      adminResp.AdminUserID,
		UserName:    "Mgmt Admin",
		Role:        "admin",
		WorkspaceID: adminResp.WorkspaceID,
	}

	return auth, adminResp.WorkspaceID
}

func teardownAdminForUserMgmt(t *testing.T) {
	t.Helper()
	TearDownWorkspace(t, "Mgmt Workspace")
}

// --- Invite User ---

func TestAdminInviteUser_Success(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	invite, status := AdminInviteUser(t, auth, "New User", "newuser@example.com", "editor")
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}

	if invite.Name != "New User" {
		t.Errorf("expected name 'New User', got '%s'", invite.Name)
	}
	if invite.Email != "newuser@example.com" {
		t.Errorf("expected email 'newuser@example.com', got '%s'", invite.Email)
	}
	if invite.Role != "editor" {
		t.Errorf("expected role 'editor', got '%s'", invite.Role)
	}
	if invite.Password == "" {
		t.Error("expected non-empty generated password")
	}
	if invite.UserID == uuid.Nil {
		t.Error("expected non-nil user ID")
	}

	// Verify login with generated password
	_, loginResp, loginStatus := LoginUser(t, "newuser@example.com", invite.Password)
	if loginStatus != http.StatusOK {
		t.Fatalf("expected login 200, got %d", loginStatus)
	}
	if loginResp.UserID != invite.UserID {
		t.Errorf("login user ID mismatch: expected %s, got %s", invite.UserID, loginResp.UserID)
	}
}

func TestAdminInviteUser_DuplicateEmail(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	_, status := AdminInviteUser(t, auth, "User One", "dup@example.com", "user")
	if status != http.StatusCreated {
		t.Fatalf("first invite: expected 201, got %d", status)
	}

	_, status = AdminInviteUser(t, auth, "User Two", "dup@example.com", "user")
	if status != http.StatusConflict {
		t.Errorf("duplicate email: expected 409, got %d", status)
	}
}

func TestAdminInviteUser_NonAdminForbidden(t *testing.T) {
	auth, workspaceID := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	// Create a non-admin user via invite
	invite, inviteStatus := AdminInviteUser(t, auth, "Editor User", "editor-forbidden@example.com", "editor")
	if inviteStatus != http.StatusCreated {
		t.Fatalf("setup invite: expected 201, got %d", inviteStatus)
	}

	editorAuth := AuthContext{
		UserID:      invite.UserID,
		UserName:    "Editor User",
		Role:        "editor",
		WorkspaceID: workspaceID,
	}

	_, status := AdminInviteUser(t, editorAuth, "Another User", "another@example.com", "user")
	if status != http.StatusForbidden {
		t.Errorf("non-admin invite: expected 403, got %d", status)
	}
}

func TestAdminInviteUser_InvalidEmail(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	_, status := AdminInviteUser(t, auth, "Bad Email", "not-an-email", "user")
	if status != http.StatusBadRequest {
		t.Errorf("invalid email: expected 400, got %d", status)
	}
}

func TestAdminInviteUser_InvalidRole(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	_, status := AdminInviteUser(t, auth, "Bad Role", "badrole@example.com", "superadmin")
	if status != http.StatusBadRequest {
		t.Errorf("invalid role: expected 400, got %d", status)
	}
}

// --- Reset Password ---

func TestAdminResetPassword_Success(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	invite, invStatus := AdminInviteUser(t, auth, "Reset Target", "reset@example.com", "user")
	if invStatus != http.StatusCreated {
		t.Fatalf("setup invite: expected 201, got %d", invStatus)
	}
	oldPassword := invite.Password

	reset, status := AdminResetUserPassword(t, auth, invite.UserID)
	if status != http.StatusOK {
		t.Fatalf("reset password: expected 200, got %d", status)
	}

	if reset.Password == "" {
		t.Error("expected non-empty new password")
	}
	if reset.Password == oldPassword {
		t.Error("new password should differ from old password")
	}

	// Old password should fail
	_, _, oldLoginStatus := LoginUser(t, "reset@example.com", oldPassword)
	if oldLoginStatus != http.StatusUnauthorized {
		t.Errorf("old password: expected 401, got %d", oldLoginStatus)
	}

	// New password should work
	_, _, newLoginStatus := LoginUser(t, "reset@example.com", reset.Password)
	if newLoginStatus != http.StatusOK {
		t.Errorf("new password: expected 200, got %d", newLoginStatus)
	}
}

func TestAdminResetPassword_NotFound(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	_, status := AdminResetUserPassword(t, auth, uuid.New())
	if status != http.StatusNotFound {
		t.Errorf("not found: expected 404, got %d", status)
	}
}

func TestAdminResetPassword_NonAdminForbidden(t *testing.T) {
	auth, workspaceID := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	invite, invStatus := AdminInviteUser(t, auth, "Target", "target-reset@example.com", "user")
	if invStatus != http.StatusCreated {
		t.Fatalf("setup invite: expected 201, got %d", invStatus)
	}

	userAuth := AuthContext{
		UserID:      invite.UserID,
		UserName:    "Target",
		Role:        "user",
		WorkspaceID: workspaceID,
	}

	_, status := AdminResetUserPassword(t, userAuth, auth.UserID)
	if status != http.StatusForbidden {
		t.Errorf("non-admin reset: expected 403, got %d", status)
	}
}

// --- List Users ---

func TestAdminListUsers_Success(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	// Invite a user
	if _, s := AdminInviteUser(t, auth, "Listed User", "listed@example.com", "editor"); s != http.StatusCreated {
		t.Fatalf("setup invite: expected 201, got %d", s)
	}

	users, status := AdminListUsers(t, auth)
	if status != http.StatusOK {
		t.Fatalf("list users: expected 200, got %d", status)
	}

	if len(users) < 2 {
		t.Fatalf("expected at least 2 users (admin + invited), got %d", len(users))
	}

	// Verify invited user is in the list
	found := false
	for _, u := range users {
		if u.Email == "listed@example.com" {
			found = true
			if u.Role != "editor" {
				t.Errorf("expected role 'editor', got '%s'", u.Role)
			}
		}
	}
	if !found {
		t.Error("invited user not found in list")
	}
}

func TestAdminListUsers_NonAdminForbidden(t *testing.T) {
	auth, workspaceID := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	invite, invStatus := AdminInviteUser(t, auth, "Non-admin Lister", "lister@example.com", "user")
	if invStatus != http.StatusCreated {
		t.Fatalf("setup invite: expected 201, got %d", invStatus)
	}

	userAuth := AuthContext{
		UserID:      invite.UserID,
		UserName:    "Non-admin Lister",
		Role:        "user",
		WorkspaceID: workspaceID,
	}

	_, status := AdminListUsers(t, userAuth)
	if status != http.StatusForbidden {
		t.Errorf("non-admin list: expected 403, got %d", status)
	}
}

// --- Delete User ---

func TestAdminDeleteUser_Success(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	invite, invStatus := AdminInviteUser(t, auth, "Delete Me", "deleteme@example.com", "user")
	if invStatus != http.StatusCreated {
		t.Fatalf("setup invite: expected 201, got %d", invStatus)
	}

	status := AdminDeleteUser(t, auth, invite.UserID)
	if status != http.StatusNoContent {
		t.Errorf("delete: expected 204, got %d", status)
	}

	// Verify user is gone — login should fail
	_, _, loginStatus := LoginUser(t, "deleteme@example.com", invite.Password)
	if loginStatus != http.StatusUnauthorized {
		t.Errorf("deleted user login: expected 401, got %d", loginStatus)
	}
}

func TestAdminDeleteUser_SelfDeletion(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	status := AdminDeleteUser(t, auth, auth.UserID)
	if status != http.StatusBadRequest {
		t.Errorf("self-delete: expected 400, got %d", status)
	}
}

func TestAdminDeleteUser_NonAdminForbidden(t *testing.T) {
	auth, workspaceID := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	invite, invStatus := AdminInviteUser(t, auth, "Editor Deleter", "editor-deleter@example.com", "editor")
	if invStatus != http.StatusCreated {
		t.Fatalf("setup invite: expected 201, got %d", invStatus)
	}

	editorAuth := AuthContext{
		UserID:      invite.UserID,
		UserName:    "Editor Deleter",
		Role:        "editor",
		WorkspaceID: workspaceID,
	}

	status := AdminDeleteUser(t, editorAuth, auth.UserID)
	if status != http.StatusForbidden {
		t.Errorf("non-admin delete: expected 403, got %d", status)
	}
}

func TestAdminDeleteUser_NotFound(t *testing.T) {
	auth, _ := setupAdminForUserMgmt(t)
	defer teardownAdminForUserMgmt(t)

	status := AdminDeleteUser(t, auth, uuid.New())
	if status != http.StatusNotFound {
		t.Errorf("not found delete: expected 404, got %d", status)
	}
}
