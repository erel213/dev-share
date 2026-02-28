package integration_tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateWorkspace_Success(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	workspace, status := CreateWorkspace(t, auth, "Test Workspace", "A test workspace", adminID)

	if status != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", status)
	}

	if workspace.ID == uuid.Nil {
		t.Error("expected non-nil workspace ID")
	}

	if workspace.Name != "Test Workspace" {
		t.Errorf("expected name 'Test Workspace', got '%s'", workspace.Name)
	}

	if workspace.Description != "A test workspace" {
		t.Errorf("expected description 'A test workspace', got '%s'", workspace.Description)
	}

	if workspace.AdminID != adminID {
		t.Errorf("expected admin ID %s, got %s", adminID, workspace.AdminID)
	}

	if workspace.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	if workspace.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestCreateWorkspace_ValidationErrors(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}

	tests := []struct {
		name        string
		wsName      string
		description string
		adminID     uuid.UUID
		wantStatus  int
	}{
		{
			name:        "missing name",
			wsName:      "",
			description: "test",
			adminID:     uuid.New(),
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "name too short",
			wsName:      "ab",
			description: "test",
			adminID:     uuid.New(),
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "nil admin ID",
			wsName:      "Valid Name",
			description: "test",
			adminID:     uuid.Nil,
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, status := CreateWorkspace(t, auth, tt.wsName, tt.description, tt.adminID)
			if status != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, status)
			}
		})
	}
}

func TestGetWorkspace_Success(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	created, _ := CreateWorkspace(t, auth, "Get Test Workspace", "Description", adminID)

	fetched, status := GetWorkspace(t, auth, created.ID)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}

	if fetched.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, fetched.ID)
	}

	if fetched.Name != created.Name {
		t.Errorf("expected name '%s', got '%s'", created.Name, fetched.Name)
	}

	if fetched.Description != created.Description {
		t.Errorf("expected description '%s', got '%s'", created.Description, fetched.Description)
	}

	if fetched.AdminID != created.AdminID {
		t.Errorf("expected admin ID %s, got %s", created.AdminID, fetched.AdminID)
	}
}

func TestGetWorkspace_NotFound(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	randomID := uuid.New()
	_, status := GetWorkspace(t, auth, randomID)

	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
}

func TestGetWorkspacesByAdmin_Success(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()

	CreateWorkspace(t, auth, "Admin Workspace 1", "First workspace", adminID)
	CreateWorkspace(t, auth, "Admin Workspace 2", "Second workspace", adminID)

	workspaces, status := GetWorkspacesByAdmin(t, auth, adminID)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}

	if len(workspaces) < 2 {
		t.Errorf("expected at least 2 workspaces, got %d", len(workspaces))
	}

	for _, ws := range workspaces {
		if ws.AdminID != adminID {
			t.Errorf("expected all workspaces to have admin ID %s, got %s", adminID, ws.AdminID)
		}
	}
}

func TestGetWorkspacesByAdmin_Empty(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	randomAdminID := uuid.New()

	workspaces, status := GetWorkspacesByAdmin(t, auth, randomAdminID)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}

	if len(workspaces) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(workspaces))
	}
}

func TestUpdateWorkspace_Success(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	created, _ := CreateWorkspace(t, auth, "Original Name", "Original Description", adminID)

	// SQLite CURRENT_TIMESTAMP has second-level precision; wait to ensure a distinct updated_at.
	time.Sleep(1 * time.Second)

	updated, status := UpdateWorkspace(t, auth, created.ID, "Updated Name", "Updated Description")

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", updated.Name)
	}

	if updated.Description != "Updated Description" {
		t.Errorf("expected description 'Updated Description', got '%s'", updated.Description)
	}

	if updated.ID != created.ID {
		t.Errorf("ID should not change: expected %s, got %s", created.ID, updated.ID)
	}

	if updated.AdminID != created.AdminID {
		t.Errorf("AdminID should not change: expected %s, got %s", created.AdminID, updated.AdminID)
	}

	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Error("UpdatedAt should be later than original")
	}
}

func TestUpdateWorkspace_NotFound(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	randomID := uuid.New()
	_, status := UpdateWorkspace(t, auth, randomID, "New Name", "New Description")

	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
}

func TestDeleteWorkspace_Success(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	created, _ := CreateWorkspace(t, auth, "To Delete", "Will be deleted", adminID)

	status := DeleteWorkspace(t, auth, created.ID)

	if status != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", status)
	}

	_, getStatus := GetWorkspace(t, auth, created.ID)
	if getStatus != http.StatusNotFound {
		t.Errorf("expected workspace to be deleted (404), got status %d", getStatus)
	}
}

func TestDeleteWorkspace_NotFound(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	randomID := uuid.New()
	status := DeleteWorkspace(t, auth, randomID)

	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
}

func TestListWorkspaces_Success(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()

	CreateWorkspace(t, auth, "List Test 1", "First", adminID)
	CreateWorkspace(t, auth, "List Test 2", "Second", adminID)
	CreateWorkspace(t, auth, "List Test 3", "Third", adminID)

	workspaces, status := ListWorkspaces(t, auth, 10, 0, "", "")

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}

	if len(workspaces) < 3 {
		t.Errorf("expected at least 3 workspaces, got %d", len(workspaces))
	}
}

func TestListWorkspaces_Pagination(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()

	for i := range 5 {
		CreateWorkspace(t, auth, "Pagination Test "+string(rune('A'+i)), "Test workspace", adminID)
	}

	page1, status1 := ListWorkspaces(t, auth, 2, 0, "created_at", "DESC")
	if status1 != http.StatusOK {
		t.Fatalf("expected status 200 for page 1, got %d", status1)
	}

	if len(page1) != 2 {
		t.Errorf("expected 2 workspaces on page 1, got %d", len(page1))
	}

	page2, status2 := ListWorkspaces(t, auth, 2, 2, "created_at", "DESC")
	if status2 != http.StatusOK {
		t.Fatalf("expected status 200 for page 2, got %d", status2)
	}

	if len(page2) != 2 {
		t.Errorf("expected 2 workspaces on page 2, got %d", len(page2))
	}

	if len(page1) > 0 && len(page2) > 0 && page1[0].ID == page2[0].ID {
		t.Error("pages should not have overlapping workspaces")
	}
}
