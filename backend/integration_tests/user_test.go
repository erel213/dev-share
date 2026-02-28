package integration_tests

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestCreateUser_Success(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	workspace, _ := CreateWorkspace(t, auth, "User Test Workspace", "For user tests", adminID)

	user, status := CreateUser(t, "John Doe", "john@example.com", "SecureP@ssw0rd!", workspace.ID)

	if status != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", status)
	}

	if user.UserID == uuid.Nil {
		t.Error("expected non-nil user ID")
	}

	if user.Message != "User created successfully" {
		t.Errorf("expected message 'User created successfully', got '%s'", user.Message)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	workspace, _ := CreateWorkspace(t, auth, "Duplicate Email Test", "Workspace", adminID)

	CreateUser(t, "First User", "duplicate@example.com", "SecureP@ss1!", workspace.ID)

	_, status := CreateUser(t, "Second User", "duplicate@example.com", "DifferentP@ss1!", workspace.ID)

	if status != http.StatusConflict {
		t.Errorf("expected status 409 for duplicate email, got %d", status)
	}
}

func TestCreateUser_InvalidWorkspace(t *testing.T) {
	randomWorkspaceID := uuid.New()

	_, status := CreateUser(t, "Test User", "test@example.com", "ValidP@ssw0rd!", randomWorkspaceID)

	if status != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid workspace, got %d", status)
	}
}

func TestCreateUser_WeakPassword(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	workspace, _ := CreateWorkspace(t, auth, "Weak Password Test", "Workspace", adminID)

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "too short",
			password: "Pass1!",
		},
		{
			name:     "no uppercase",
			password: "password123!",
		},
		{
			name:     "no lowercase",
			password: "PASSWORD123!",
		},
		{
			name:     "no number",
			password: "Password!",
		},
		{
			name:     "no special char",
			password: "Password123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, status := CreateUser(t, "Test User", "weak"+tt.name+"@example.com", tt.password, workspace.ID)

			if status != http.StatusBadRequest {
				t.Errorf("expected status 400 for weak password (%s), got %d", tt.name, status)
			}
		})
	}
}

func TestCreateUser_ValidationErrors(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	workspace, _ := CreateWorkspace(t, auth, "Validation Test", "Workspace", adminID)

	tests := []struct {
		name       string
		userName   string
		email      string
		password   string
		wantStatus int
	}{
		{
			name:       "missing name",
			userName:   "",
			email:      "test@example.com",
			password:   "ValidP@ssw0rd!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "name too short",
			userName:   "a",
			email:      "test@example.com",
			password:   "ValidP@ssw0rd!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email",
			userName:   "Test User",
			email:      "not-an-email",
			password:   "ValidP@ssw0rd!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			userName:   "Test User",
			email:      "",
			password:   "ValidP@ssw0rd!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			userName:   "Test User",
			email:      "test@example.com",
			password:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, status := CreateUser(t, tt.userName, tt.email, tt.password, workspace.ID)

			if status != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, status)
			}
		})
	}
}

func TestCreateUser_CascadeDelete(t *testing.T) {
	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	workspace, _ := CreateWorkspace(t, auth, "Cascade Test", "Workspace for cascade delete", adminID)

	user, status := CreateUser(t, "Cascade User", "cascade@example.com", "CascadeP@ss1!", workspace.ID)
	if status != http.StatusCreated {
		t.Fatalf("failed to create user: status %d", status)
	}

	if user.UserID == uuid.Nil {
		t.Fatal("user ID is nil")
	}

	deleteStatus := DeleteWorkspace(t, auth, workspace.ID)
	if deleteStatus != http.StatusNoContent {
		t.Fatalf("failed to delete workspace: status %d", deleteStatus)
	}
}
