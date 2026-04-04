package integration_tests

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func setupGroupTests(t *testing.T) (AuthContext, uuid.UUID) {
	t.Helper()

	adminResp, status := InitializeAdmin(
		t,
		"Group Admin",
		"group-admin@example.com",
		"StrongP@ssw0rd123",
		"Group Workspace",
		"Group test workspace",
		"",
	)
	if status != http.StatusCreated {
		t.Fatalf("failed to init admin: status %d", status)
	}

	auth := AuthContext{
		UserID:      adminResp.AdminUserID,
		UserName:    "Group Admin",
		Role:        "admin",
		WorkspaceID: adminResp.WorkspaceID,
	}

	return auth, adminResp.WorkspaceID
}

func teardownGroupTests(t *testing.T, workspaceID uuid.UUID) {
	t.Helper()
	TearDownGroups(t, workspaceID)
	TearDownWorkspace(t, "Group Workspace")
}

// --- Create Group ---

func TestCreateGroup_Success(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, status := CreateGroup(t, auth, "Engineering", "Backend engineers", false)
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}

	if group.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got '%s'", group.Name)
	}
	if group.Description != "Backend engineers" {
		t.Errorf("expected description 'Backend engineers', got '%s'", group.Description)
	}
	if group.AccessAllTemplates != false {
		t.Errorf("expected access_all_templates false, got true")
	}
	if group.WorkspaceID != workspaceID {
		t.Errorf("expected workspace_id %s, got %s", workspaceID, group.WorkspaceID)
	}
	if group.ID == uuid.Nil {
		t.Error("expected non-nil group ID")
	}
}

func TestCreateGroup_AccessAllTemplates(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, status := CreateGroup(t, auth, "Admins", "Full access group", true)
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}

	if group.AccessAllTemplates != true {
		t.Errorf("expected access_all_templates true, got false")
	}
}

func TestCreateGroup_DuplicateName(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	_, status := CreateGroup(t, auth, "Engineering", "First", false)
	if status != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", status)
	}

	_, status = CreateGroup(t, auth, "Engineering", "Second", false)
	if status == http.StatusCreated {
		t.Errorf("duplicate name: expected error status, got 201")
	}
}

func TestCreateGroup_ValidationError(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	// Name too short
	_, status := CreateGroup(t, auth, "ab", "", false)
	if status != http.StatusBadRequest {
		t.Errorf("short name: expected 400, got %d", status)
	}
}

func TestCreateGroup_NonAdminForbidden(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	invite, invStatus := AdminInviteUser(t, auth, "Editor User", "group-editor@example.com", "editor")
	if invStatus != http.StatusCreated {
		t.Fatalf("setup invite: expected 201, got %d", invStatus)
	}

	editorAuth := AuthContext{
		UserID:      invite.UserID,
		UserName:    "Editor User",
		Role:        "editor",
		WorkspaceID: workspaceID,
	}

	_, status := CreateGroup(t, editorAuth, "Editors", "Editor group", false)
	if status != http.StatusForbidden {
		t.Errorf("non-admin create: expected 403, got %d", status)
	}
}

// --- Get Group ---

func TestGetGroup_Success(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	created, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)

	group, status := GetGroup(t, auth, created.ID)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	if group.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, group.ID)
	}
	if group.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got '%s'", group.Name)
	}
}

func TestGetGroup_NotFound(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	_, status := GetGroup(t, auth, uuid.New())
	if status != http.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

// --- List Groups ---

func TestListGroups_Success(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	CreateGroup(t, auth, "Engineering", "Engineers", false)
	CreateGroup(t, auth, "Design", "Designers", true)

	groups, status := ListGroups(t, auth)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestListGroups_Empty(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	groups, status := ListGroups(t, auth)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	if groups != nil && len(groups) != 0 {
		t.Errorf("expected empty list, got %d groups", len(groups))
	}
}

// --- Update Group ---

func TestUpdateGroup_Success(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	created, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)

	updated, status := UpdateGroup(t, auth, created.ID, map[string]interface{}{
		"name":                 "Platform Engineering",
		"description":          "Platform team",
		"access_all_templates": true,
	})
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	if updated.Name != "Platform Engineering" {
		t.Errorf("expected name 'Platform Engineering', got '%s'", updated.Name)
	}
	if updated.Description != "Platform team" {
		t.Errorf("expected description 'Platform team', got '%s'", updated.Description)
	}
	if updated.AccessAllTemplates != true {
		t.Errorf("expected access_all_templates true, got false")
	}
}

func TestUpdateGroup_NotFound(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	_, status := UpdateGroup(t, auth, uuid.New(), map[string]interface{}{
		"name": "Nonexistent",
	})
	if status != http.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

// --- Delete Group ---

func TestDeleteGroup_Success(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	created, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)

	status := DeleteGroup(t, auth, created.ID)
	if status != http.StatusNoContent {
		t.Errorf("expected 204, got %d", status)
	}

	// Verify group is gone
	_, getStatus := GetGroup(t, auth, created.ID)
	if getStatus != http.StatusNotFound {
		t.Errorf("deleted group: expected 404, got %d", getStatus)
	}
}

func TestDeleteGroup_NotFound(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	status := DeleteGroup(t, auth, uuid.New())
	if status != http.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

// --- Membership ---

func TestGroupMembers_AddAndList(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)

	invite1, _ := AdminInviteUser(t, auth, "User One", "group-user1@example.com", "user")
	invite2, _ := AdminInviteUser(t, auth, "User Two", "group-user2@example.com", "user")

	status := AddGroupMembers(t, auth, group.ID, []uuid.UUID{invite1.UserID, invite2.UserID})
	if status != http.StatusNoContent {
		t.Fatalf("add members: expected 204, got %d", status)
	}

	members, listStatus := GetGroupMembers(t, auth, group.ID)
	if listStatus != http.StatusOK {
		t.Fatalf("list members: expected 200, got %d", listStatus)
	}

	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
}

func TestGroupMembers_AddDuplicate(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)
	invite, _ := AdminInviteUser(t, auth, "User One", "group-dup-member@example.com", "user")

	AddGroupMembers(t, auth, group.ID, []uuid.UUID{invite.UserID})

	// Adding same user again should not error (ON CONFLICT DO NOTHING)
	status := AddGroupMembers(t, auth, group.ID, []uuid.UUID{invite.UserID})
	if status != http.StatusNoContent {
		t.Errorf("duplicate add: expected 204, got %d", status)
	}

	// Should still be only 1 member
	members, _ := GetGroupMembers(t, auth, group.ID)
	if len(members) != 1 {
		t.Errorf("expected 1 member after duplicate add, got %d", len(members))
	}
}

func TestGroupMembers_Remove(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)
	invite, _ := AdminInviteUser(t, auth, "User One", "group-rm-member@example.com", "user")

	AddGroupMembers(t, auth, group.ID, []uuid.UUID{invite.UserID})

	status := RemoveGroupMember(t, auth, group.ID, invite.UserID)
	if status != http.StatusNoContent {
		t.Errorf("remove member: expected 204, got %d", status)
	}

	members, _ := GetGroupMembers(t, auth, group.ID)
	if len(members) != 0 {
		t.Errorf("expected 0 members after remove, got %d", len(members))
	}
}

func TestGroupMembers_RemoveNotFound(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)

	status := RemoveGroupMember(t, auth, group.ID, uuid.New())
	if status != http.StatusNotFound {
		t.Errorf("remove non-member: expected 404, got %d", status)
	}
}

// --- Template Access ---

func TestGroupTemplateAccess_AddAndList(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)

	tmpl1, _ := CreateTemplate(t, auth, "tmpl-one", workspaceID, map[string]string{"main.tf": "resource {}"})
	tmpl2, _ := CreateTemplate(t, auth, "tmpl-two", workspaceID, map[string]string{"main.tf": "resource {}"})

	status := AddGroupTemplateAccess(t, auth, group.ID, []uuid.UUID{tmpl1.ID, tmpl2.ID})
	if status != http.StatusNoContent {
		t.Fatalf("add template access: expected 204, got %d", status)
	}

	templateIDs, listStatus := GetGroupTemplateAccess(t, auth, group.ID)
	if listStatus != http.StatusOK {
		t.Fatalf("list template access: expected 200, got %d", listStatus)
	}

	if len(templateIDs) != 2 {
		t.Fatalf("expected 2 template accesses, got %d", len(templateIDs))
	}
}

func TestGroupTemplateAccess_Remove(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)
	tmpl, _ := CreateTemplate(t, auth, "tmpl-rm", workspaceID, map[string]string{"main.tf": "resource {}"})

	AddGroupTemplateAccess(t, auth, group.ID, []uuid.UUID{tmpl.ID})

	status := RemoveGroupTemplateAccess(t, auth, group.ID, tmpl.ID)
	if status != http.StatusNoContent {
		t.Errorf("remove template access: expected 204, got %d", status)
	}

	templateIDs, _ := GetGroupTemplateAccess(t, auth, group.ID)
	if len(templateIDs) != 0 {
		t.Errorf("expected 0 template accesses after remove, got %d", len(templateIDs))
	}
}

func TestGroupTemplateAccess_RemoveNotFound(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)

	status := RemoveGroupTemplateAccess(t, auth, group.ID, uuid.New())
	if status != http.StatusNotFound {
		t.Errorf("remove non-existent access: expected 404, got %d", status)
	}
}

// --- Cascade Deletion ---

func TestDeleteGroup_CascadesMemberships(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	group, _ := CreateGroup(t, auth, "Engineering", "Engineers", false)
	invite, _ := AdminInviteUser(t, auth, "Cascade User", "group-cascade@example.com", "user")
	AddGroupMembers(t, auth, group.ID, []uuid.UUID{invite.UserID})

	tmpl, _ := CreateTemplate(t, auth, "tmpl-cascade", workspaceID, map[string]string{"main.tf": "resource {}"})
	AddGroupTemplateAccess(t, auth, group.ID, []uuid.UUID{tmpl.ID})

	// Delete the group
	status := DeleteGroup(t, auth, group.ID)
	if status != http.StatusNoContent {
		t.Fatalf("delete group: expected 204, got %d", status)
	}

	// Verify memberships and template access are gone via DB
	var memberCount int
	DbConnection.QueryRow("SELECT COUNT(*) FROM group_memberships WHERE group_id = ?", group.ID).Scan(&memberCount)
	if memberCount != 0 {
		t.Errorf("expected 0 memberships after cascade delete, got %d", memberCount)
	}

	var accessCount int
	DbConnection.QueryRow("SELECT COUNT(*) FROM group_template_access WHERE group_id = ?", group.ID).Scan(&accessCount)
	if accessCount != 0 {
		t.Errorf("expected 0 template accesses after cascade delete, got %d", accessCount)
	}
}

// --- Template Access Filtering ---

func TestTemplateListFiltering_UserWithGroup(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	// Create two templates
	tmpl1, _ := CreateTemplate(t, auth, "tmpl-visible", workspaceID, map[string]string{"main.tf": "resource {}"})
	CreateTemplate(t, auth, "tmpl-hidden", workspaceID, map[string]string{"main.tf": "resource {}"})

	// Create a group with access to only tmpl1
	group, _ := CreateGroup(t, auth, "Limited", "Limited access", false)
	AddGroupTemplateAccess(t, auth, group.ID, []uuid.UUID{tmpl1.ID})

	// Create a non-admin user and add to the group
	invite, _ := AdminInviteUser(t, auth, "Limited User", "group-limited@example.com", "user")
	AddGroupMembers(t, auth, group.ID, []uuid.UUID{invite.UserID})

	userAuth := AuthContext{
		UserID:      invite.UserID,
		UserName:    "Limited User",
		Role:        "user",
		WorkspaceID: workspaceID,
	}

	// User should only see tmpl1
	templates, status := ListTemplates(t, userAuth, 50, 0, "", "")
	if status != http.StatusOK {
		t.Fatalf("list templates: expected 200, got %d", status)
	}

	if len(templates) != 1 {
		t.Fatalf("expected 1 template visible to user, got %d", len(templates))
	}
	if templates[0].ID != tmpl1.ID {
		t.Errorf("expected visible template %s, got %s", tmpl1.ID, templates[0].ID)
	}
}

func TestTemplateListFiltering_UserWithAccessAllGroup(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	CreateTemplate(t, auth, "tmpl-a", workspaceID, map[string]string{"main.tf": "resource {}"})
	CreateTemplate(t, auth, "tmpl-b", workspaceID, map[string]string{"main.tf": "resource {}"})

	// Create a group with access_all_templates=true
	group, _ := CreateGroup(t, auth, "Full Access", "All access", true)

	invite, _ := AdminInviteUser(t, auth, "Full User", "group-full@example.com", "user")
	AddGroupMembers(t, auth, group.ID, []uuid.UUID{invite.UserID})

	userAuth := AuthContext{
		UserID:      invite.UserID,
		UserName:    "Full User",
		Role:        "user",
		WorkspaceID: workspaceID,
	}

	templates, status := ListTemplates(t, userAuth, 50, 0, "", "")
	if status != http.StatusOK {
		t.Fatalf("list templates: expected 200, got %d", status)
	}

	if len(templates) != 2 {
		t.Errorf("expected 2 templates for access_all user, got %d", len(templates))
	}
}

func TestTemplateListFiltering_UserWithNoGroup(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	CreateTemplate(t, auth, "tmpl-no-group", workspaceID, map[string]string{"main.tf": "resource {}"})

	invite, _ := AdminInviteUser(t, auth, "No Group User", "group-none@example.com", "user")

	userAuth := AuthContext{
		UserID:      invite.UserID,
		UserName:    "No Group User",
		Role:        "user",
		WorkspaceID: workspaceID,
	}

	templates, status := ListTemplates(t, userAuth, 50, 0, "", "")
	if status != http.StatusOK {
		t.Fatalf("list templates: expected 200, got %d", status)
	}

	if len(templates) != 0 {
		t.Errorf("expected 0 templates for user with no group, got %d", len(templates))
	}
}

func TestTemplateListFiltering_AdminSeesAll(t *testing.T) {
	auth, workspaceID := setupGroupTests(t)
	defer teardownGroupTests(t, workspaceID)

	CreateTemplate(t, auth, "tmpl-admin-a", workspaceID, map[string]string{"main.tf": "resource {}"})
	CreateTemplate(t, auth, "tmpl-admin-b", workspaceID, map[string]string{"main.tf": "resource {}"})

	// Admin has no groups, but should still see all templates
	templates, status := ListTemplates(t, auth, 50, 0, "", "")
	if status != http.StatusOK {
		t.Fatalf("list templates: expected 200, got %d", status)
	}

	if len(templates) != 2 {
		t.Errorf("expected admin to see 2 templates, got %d", len(templates))
	}
}
