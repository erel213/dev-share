package integration_tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// setupWorkspaceForTemplates creates a real workspace and returns an auth context
// whose WorkspaceID matches the created workspace (required by the template service's
// workspace-isolation checks).
func setupWorkspaceForTemplates(t *testing.T) (AuthContext, *WorkspaceResponse) {
	t.Helper()

	bootstrapAuth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Template Test User",
		WorkspaceID: uuid.New(),
	}
	workspace, status := CreateWorkspace(t, bootstrapAuth, "Template WS "+uuid.New().String()[:8], "Workspace for template tests", uuid.New())
	if status != http.StatusCreated {
		t.Fatalf("setupWorkspaceForTemplates: failed to create workspace, status %d", status)
	}

	auth := AuthContext{
		UserID:      bootstrapAuth.UserID,
		UserName:    bootstrapAuth.UserName,
		WorkspaceID: workspace.ID,
	}
	return auth, workspace
}

// --- Create ---

func TestCreateTemplate_Success(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)

	template, status := CreateTemplate(t, auth, "My Template", workspace.ID, "/path/to/template")

	if status != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", status)
	}

	if template.ID == uuid.Nil {
		t.Error("expected non-nil template ID")
	}
	if template.Name != "My Template" {
		t.Errorf("expected name 'My Template', got '%s'", template.Name)
	}
	if template.WorkspaceID != workspace.ID {
		t.Errorf("expected workspace ID %s, got %s", workspace.ID, template.WorkspaceID)
	}
	if template.Path != "/path/to/template" {
		t.Errorf("expected path '/path/to/template', got '%s'", template.Path)
	}
	if template.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if template.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestCreateTemplate_ValidationErrors(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)

	tests := []struct {
		name       string
		tmplName   string
		path       string
		wantStatus int
	}{
		{
			name:       "missing name",
			tmplName:   "",
			path:       "/valid/path",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "name too short",
			tmplName:   "ab",
			path:       "/valid/path",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing path",
			tmplName:   "Valid Name",
			path:       "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, status := CreateTemplate(t, auth, tt.tmplName, workspace.ID, tt.path)
			if status != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, status)
			}
		})
	}
}

func TestCreateTemplate_ForbiddenOtherWorkspace(t *testing.T) {
	auth, _ := setupWorkspaceForTemplates(t)
	_, otherWorkspace := setupWorkspaceForTemplates(t)

	// JWT workspace differs from the requested workspace_id → 403
	_, status := CreateTemplate(t, auth, "Forbidden Template", otherWorkspace.ID, "/path")

	if status != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", status)
	}
}

// --- Get ---

func TestGetTemplate_Success(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)
	created, _ := CreateTemplate(t, auth, "Get Test Template", workspace.ID, "/path/to/template")

	fetched, status := GetTemplate(t, auth, created.ID)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if fetched.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, fetched.ID)
	}
	if fetched.Name != created.Name {
		t.Errorf("expected name '%s', got '%s'", created.Name, fetched.Name)
	}
	if fetched.WorkspaceID != created.WorkspaceID {
		t.Errorf("expected workspace ID %s, got %s", created.WorkspaceID, fetched.WorkspaceID)
	}
	if fetched.Path != created.Path {
		t.Errorf("expected path '%s', got '%s'", created.Path, fetched.Path)
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	auth, _ := setupWorkspaceForTemplates(t)

	_, status := GetTemplate(t, auth, uuid.New())

	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
}

func TestGetTemplate_InvalidID(t *testing.T) {
	auth, _ := setupWorkspaceForTemplates(t)

	req, _ := http.NewRequest(http.MethodGet, BaseURL+"/api/v1/templates/not-a-uuid", nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestGetTemplate_ForbiddenOtherWorkspace(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)
	otherAuth, _ := setupWorkspaceForTemplates(t)

	created, _ := CreateTemplate(t, auth, "Workspace A Template", workspace.ID, "/path")

	_, status := GetTemplate(t, otherAuth, created.ID)

	if status != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", status)
	}
}

// --- GetByWorkspace ---

func TestGetTemplatesByWorkspace_Success(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)

	CreateTemplate(t, auth, "Workspace Template 1", workspace.ID, "/path/1")
	CreateTemplate(t, auth, "Workspace Template 2", workspace.ID, "/path/2")

	templates, status := GetTemplatesByWorkspace(t, auth, workspace.ID)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if len(templates) < 2 {
		t.Errorf("expected at least 2 templates, got %d", len(templates))
	}
	for _, tmpl := range templates {
		if tmpl.WorkspaceID != workspace.ID {
			t.Errorf("expected workspace ID %s, got %s", workspace.ID, tmpl.WorkspaceID)
		}
	}
}

func TestGetTemplatesByWorkspace_Empty(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)

	templates, status := GetTemplatesByWorkspace(t, auth, workspace.ID)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if len(templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(templates))
	}
}

func TestGetTemplatesByWorkspace_Forbidden(t *testing.T) {
	_, workspace := setupWorkspaceForTemplates(t)
	otherAuth, _ := setupWorkspaceForTemplates(t)

	_, status := GetTemplatesByWorkspace(t, otherAuth, workspace.ID)

	if status != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", status)
	}
}

// --- Update ---

func TestUpdateTemplate_Success(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)
	created, _ := CreateTemplate(t, auth, "Original Name", workspace.ID, "/original/path")

	updated, status := UpdateTemplate(t, auth, created.ID, "Updated Name", "/updated/path")

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", updated.Name)
	}
	if updated.Path != "/updated/path" {
		t.Errorf("expected path '/updated/path', got '%s'", updated.Path)
	}
	if updated.ID != created.ID {
		t.Errorf("ID should not change: expected %s, got %s", created.ID, updated.ID)
	}
	if updated.WorkspaceID != created.WorkspaceID {
		t.Errorf("WorkspaceID should not change: expected %s, got %s", created.WorkspaceID, updated.WorkspaceID)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Error("UpdatedAt should be later than original")
	}
}

func TestUpdateTemplate_PartialUpdate(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)
	created, _ := CreateTemplate(t, auth, "Partial Update Template", workspace.ID, "/original/path")

	// Update only the name; path should remain unchanged.
	updated, status := UpdateTemplate(t, auth, created.ID, "New Name Only", "")

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if updated.Name != "New Name Only" {
		t.Errorf("expected name 'New Name Only', got '%s'", updated.Name)
	}
	if updated.Path != created.Path {
		t.Errorf("path should not change: expected '%s', got '%s'", created.Path, updated.Path)
	}
}

func TestUpdateTemplate_NotFound(t *testing.T) {
	auth, _ := setupWorkspaceForTemplates(t)

	_, status := UpdateTemplate(t, auth, uuid.New(), "New Name", "/new/path")

	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
}

func TestUpdateTemplate_ForbiddenOtherWorkspace(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)
	otherAuth, _ := setupWorkspaceForTemplates(t)

	created, _ := CreateTemplate(t, auth, "Other WS Template", workspace.ID, "/path")

	_, status := UpdateTemplate(t, otherAuth, created.ID, "Forbidden Update", "/forbidden/path")

	if status != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", status)
	}
}

// --- Delete ---

func TestDeleteTemplate_Success(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)
	created, _ := CreateTemplate(t, auth, "To Delete Template", workspace.ID, "/path/to/delete")

	status := DeleteTemplate(t, auth, created.ID)

	if status != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", status)
	}

	_, getStatus := GetTemplate(t, auth, created.ID)
	if getStatus != http.StatusNotFound {
		t.Errorf("expected template to be deleted (404), got status %d", getStatus)
	}
}

func TestDeleteTemplate_NotFound(t *testing.T) {
	auth, _ := setupWorkspaceForTemplates(t)

	status := DeleteTemplate(t, auth, uuid.New())

	if status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", status)
	}
}

func TestDeleteTemplate_ForbiddenOtherWorkspace(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)
	otherAuth, _ := setupWorkspaceForTemplates(t)

	created, _ := CreateTemplate(t, auth, "Forbidden Delete Template", workspace.ID, "/path")

	status := DeleteTemplate(t, otherAuth, created.ID)

	if status != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", status)
	}
}

// --- List ---

func TestListTemplates_Success(t *testing.T) {
	auth, workspace := setupWorkspaceForTemplates(t)

	CreateTemplate(t, auth, "List Template 1", workspace.ID, "/path/1")
	CreateTemplate(t, auth, "List Template 2", workspace.ID, "/path/2")
	CreateTemplate(t, auth, "List Template 3", workspace.ID, "/path/3")

	templates, status := ListTemplates(t, auth, 10, 0, "", "")

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if len(templates) < 3 {
		t.Errorf("expected at least 3 templates, got %d", len(templates))
	}
}

func TestListTemplates_FilteredByWorkspace(t *testing.T) {
	authA, workspaceA := setupWorkspaceForTemplates(t)
	authB, workspaceB := setupWorkspaceForTemplates(t)

	CreateTemplate(t, authA, "Workspace A Template", workspaceA.ID, "/path/a")
	CreateTemplate(t, authB, "Workspace B Template", workspaceB.ID, "/path/b")

	// Listing from workspace A's context must not include workspace B's templates.
	templates, status := ListTemplates(t, authA, 100, 0, "", "")

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	for _, tmpl := range templates {
		if tmpl.WorkspaceID == workspaceB.ID {
			t.Errorf("workspace B template should not appear in workspace A's list")
		}
	}
}

func TestListTemplates_InvalidSortBy(t *testing.T) {
	auth, _ := setupWorkspaceForTemplates(t)

	tests := []struct {
		name       string
		sortBy     string
		order      string
		wantStatus int
	}{
		{
			name:       "invalid sort_by value",
			sortBy:     "invalid_field",
			order:      "ASC",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid order value",
			sortBy:     "name",
			order:      "RANDOM",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/v1/templates?sort_by=%s&order=%s", BaseURL, tt.sortBy, tt.order)
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			addAuth(t, req, auth)

			resp, err := HTTPClient.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}
