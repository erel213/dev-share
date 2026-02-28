package integration_tests

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// TestAdminInit_SuccessAndConflict tests both successful initialization and conflict scenarios
// in a single test to ensure proper sequencing
func TestAdminInit_SuccessAndConflict(t *testing.T) {
	t.Run("Successful initialization", func(t *testing.T) {
		adminResp, status := InitializeAdmin(
			t,
			"Admin User",
			"admin@example.com",
			"StrongP@ssw0rd123",
			"My Workspace",
			"Initial workspace",
			"",
		)

		if status != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", status)
		}

		if adminResp.Message != "System initialized successfully" {
			t.Errorf("expected message 'System initialized successfully', got '%s'", adminResp.Message)
		}

		if adminResp.WorkspaceID == uuid.Nil {
			t.Error("expected non-nil workspace ID")
		}

		if adminResp.AdminUserID == uuid.Nil {
			t.Error("expected non-nil admin user ID")
		}

		// Verify workspace was created — use DB directly (HTTP workspace endpoint requires auth)
		workspace := GetWorkspaceFromDB(t, adminResp.WorkspaceID)

		if workspace.Name != "My Workspace" {
			t.Errorf("expected workspace name 'My Workspace', got '%s'", workspace.Name)
		}

		if workspace.AdminID != adminResp.AdminUserID {
			t.Errorf("expected workspace admin_id %s, got %s", adminResp.AdminUserID, workspace.AdminID)
		}
	})

	t.Run("Already initialized conflict", func(t *testing.T) {
		// Second initialization should fail
		_, status := InitializeAdmin(
			t,
			"Second Admin",
			"second@example.com",
			"StrongP@ssw0rd456",
			"Second Workspace",
			"Second workspace",
			"",
		)

		if status != http.StatusConflict {
			t.Errorf("expected status 409 Conflict, got %d", status)
		}
	})
	// Workspace teardown
	TearDownWorkspace(t, "My Workspace")
}

func TestAdminInit_InvalidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"too short", "short"},
		{"no special chars", "password123"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, status := InitializeAdmin(
				t,
				"Admin User",
				"admin@example.com",
				tt.password,
				"My Workspace",
				"Initial workspace",
				"",
			)

			if status != http.StatusBadRequest {
				t.Errorf("expected status 400 Bad Request for %s, got %d", tt.name, status)
			}
		})
	}
}

func TestAdminInit_InvalidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"invalid format", "not-an-email"},
		{"missing @", "adminexample.com"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, status := InitializeAdmin(
				t,
				"Admin User",
				tt.email,
				"StrongP@ssw0rd123",
				"My Workspace",
				"Initial workspace",
				"",
			)

			if status != http.StatusBadRequest {
				t.Errorf("expected status 400 Bad Request for %s, got %d", tt.name, status)
			}
		})
	}
}

func TestAdminInit_MissingFields(t *testing.T) {
	tests := []struct {
		name           string
		adminName      string
		adminEmail     string
		adminPassword  string
		workspaceName  string
		workspaceDesc  string
		expectedStatus int
	}{
		{
			name:           "missing admin name",
			adminName:      "",
			adminEmail:     "admin@example.com",
			adminPassword:  "StrongP@ssw0rd123",
			workspaceName:  "Workspace",
			workspaceDesc:  "Description",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing admin email",
			adminName:      "Admin User",
			adminEmail:     "",
			adminPassword:  "StrongP@ssw0rd123",
			workspaceName:  "Workspace",
			workspaceDesc:  "Description",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing admin password",
			adminName:      "Admin User",
			adminEmail:     "admin@example.com",
			adminPassword:  "",
			workspaceName:  "Workspace",
			workspaceDesc:  "Description",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing workspace name",
			adminName:      "Admin User",
			adminEmail:     "admin@example.com",
			adminPassword:  "StrongP@ssw0rd123",
			workspaceName:  "",
			workspaceDesc:  "Description",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "workspace description optional",
			adminName:      "Admin User",
			adminEmail:     "admin@example.com",
			adminPassword:  "StrongP@ssw0rd123",
			workspaceName:  "Workspace",
			workspaceDesc:  "",
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus == http.StatusCreated {
				defer TearDownWorkspace(t, tt.workspaceName)
			}
			_, status := InitializeAdmin(
				t,
				tt.adminName,
				tt.adminEmail,
				tt.adminPassword,
				tt.workspaceName,
				tt.workspaceDesc,
				"",
			)

			if status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

func TestAdminInit_AdminNameValidation(t *testing.T) {
	tests := []struct {
		name           string
		adminName      string
		expectedStatus int
	}{
		{"too short", "A", http.StatusBadRequest},
		{"valid min length", "Ad", http.StatusCreated},
		{"valid max length", "A very long admin name that is exactly one hundred characters long for testing maximum length", http.StatusCreated},
		{"too long", "A very long admin name that exceeds one hundred characters and should fail validation because it is too lengthy", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus == http.StatusCreated {
				defer TearDownWorkspace(t, "My Workspace")
			}
			_, status := InitializeAdmin(
				t,
				tt.adminName,
				"admin@example.com",
				"StrongP@ssw0rd123",
				"My Workspace",
				"Initial workspace",
				"",
			)

			if status != tt.expectedStatus {
				t.Errorf("expected status %d for %s, got %d", tt.expectedStatus, tt.name, status)
			}
		})
	}
}

func TestAdminInit_WorkspaceNameValidation(t *testing.T) {
	tests := []struct {
		name           string
		workspaceName  string
		expectedStatus int
	}{
		{"too short", "AB", http.StatusBadRequest},
		{"valid min length", "ABC", http.StatusCreated},
		{"valid max length", "A workspace name that is exactly one hundred characters long for testing the maximum length allowed", http.StatusCreated},
		{"too long", "A workspace name that exceeds one hundred characters and should fail validation because it is way too lengthy", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedStatus == http.StatusCreated {
				defer TearDownWorkspace(t, tt.workspaceName)
			}
			_, status := InitializeAdmin(
				t,
				"Admin User",
				"admin@example.com",
				"StrongP@ssw0rd123",
				tt.workspaceName,
				"Initial workspace",
				"",
			)

			if status != tt.expectedStatus {
				t.Errorf("expected status %d for %s, got %d", tt.expectedStatus, tt.name, status)
			}
		})
	}
}
