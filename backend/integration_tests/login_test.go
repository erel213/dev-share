package integration_tests

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func setupUserForLogin(t *testing.T, email, password string) uuid.UUID {
	t.Helper()

	auth := AuthContext{
		UserID:      uuid.New(),
		UserName:    "Test User",
		WorkspaceID: uuid.New(),
	}
	adminID := uuid.New()
	workspace, status := CreateWorkspace(t, auth, "Login Test Workspace "+uuid.New().String()[:8], "For login tests", adminID)
	if status != http.StatusCreated {
		t.Fatalf("failed to create workspace: status %d", status)
	}

	user, status := CreateUser(t, "Login User", email, password, workspace.ID)
	if status != http.StatusCreated {
		t.Fatalf("failed to create user: status %d", status)
	}

	return user.UserID
}

func TestLogin_Success(t *testing.T) {
	email := "login-success@example.com"
	password := "SecureP@ssw0rd!"
	userID := setupUserForLogin(t, email, password)

	resp, login, status := LoginUser(t, email, password)

	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}

	if login.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, login.UserID)
	}

	// Verify access_token cookie is set
	var foundCookie bool
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "access_token" {
			foundCookie = true
			if cookie.Value == "" {
				t.Error("expected non-empty access_token cookie")
			}
			break
		}
	}
	if !foundCookie {
		t.Error("expected access_token cookie to be set")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	email := "login-wrongpass@example.com"
	password := "SecureP@ssw0rd!"
	setupUserForLogin(t, email, password)

	_, _, status := LoginUser(t, email, "WrongP@ssw0rd!")

	if status != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", status)
	}
}

func TestLogin_NonExistentEmail(t *testing.T) {
	_, _, status := LoginUser(t, "nonexistent@example.com", "SomeP@ssw0rd!")

	if status != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", status)
	}
}

func TestLogin_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		password   string
		wantStatus int
	}{
		{
			name:       "missing email",
			email:      "",
			password:   "SecureP@ssw0rd!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email",
			email:      "not-an-email",
			password:   "SecureP@ssw0rd!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			email:      "test@example.com",
			password:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, status := LoginUser(t, tt.email, tt.password)

			if status != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, status)
			}
		})
	}
}

func TestLogin_SameErrorForAllFailures(t *testing.T) {
	// Both wrong password and non-existent email should return the same
	// status code to avoid leaking whether an email exists.
	email := "login-sameerror@example.com"
	password := "SecureP@ssw0rd!"
	setupUserForLogin(t, email, password)

	_, _, wrongPassStatus := LoginUser(t, email, "WrongP@ssw0rd!")
	_, _, noUserStatus := LoginUser(t, "nobody@example.com", "SomeP@ssw0rd!")

	if wrongPassStatus != noUserStatus {
		t.Errorf("expected same status for wrong password (%d) and non-existent email (%d)", wrongPassStatus, noUserStatus)
	}

	if wrongPassStatus != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", wrongPassStatus)
	}
}
