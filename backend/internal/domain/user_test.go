package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewLocalUser(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "ValidPassword123!",
			expectError: false,
		},
		{
			name:        "short password",
			password:    "short",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: false,
		},
		{
			name:        "long password",
			password:    string(make([]byte, 100)),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localUser, err := NewLocalUser(tt.password)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && localUser.Password == "" {
				t.Error("expected hashed password but got empty string")
			}
			if !tt.expectError && localUser.Password == tt.password {
				t.Error("password should be hashed, not plain text")
			}
		})
	}
}

func TestNewBaseUser(t *testing.T) {
	name := "John Doe"
	email := "john@example.com"
	workspaceID := uuid.New()

	before := time.Now()
	baseUser := NewBaseUser(name, email, workspaceID)
	after := time.Now()

	if baseUser.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if baseUser.Name != name {
		t.Errorf("expected name %q, got %q", name, baseUser.Name)
	}
	if baseUser.Email != email {
		t.Errorf("expected email %q, got %q", email, baseUser.Email)
	}
	if baseUser.WorkspaceID != workspaceID {
		t.Errorf("expected workspace ID %v, got %v", workspaceID, baseUser.WorkspaceID)
	}
	if baseUser.CreatedAt.Before(before) || baseUser.CreatedAt.After(after) {
		t.Errorf("CreatedAt timestamp %v out of expected range [%v, %v]", baseUser.CreatedAt, before, after)
	}
	if baseUser.UpdatedAt.Before(before) || baseUser.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt timestamp %v out of expected range [%v, %v]", baseUser.UpdatedAt, before, after)
	}
}

func TestNewThirdPartyUser(t *testing.T) {
	tests := []struct {
		name          string
		oauthProvider string
		oauthID       string
	}{
		{
			name:          "github provider",
			oauthProvider: "github",
			oauthID:       "123456",
		},
		{
			name:          "google provider",
			oauthProvider: "google",
			oauthID:       "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thirdPartyUser, err := NewThirdPartyUser(tt.oauthProvider, tt.oauthID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if thirdPartyUser == nil {
				t.Fatal("expected non-nil third party user")
			}
			if string(thirdPartyUser.OauthProvider) != tt.oauthProvider {
				t.Errorf("expected oauth provider %q, got %q", tt.oauthProvider, thirdPartyUser.OauthProvider)
			}
			if thirdPartyUser.OauthID != tt.oauthID {
				t.Errorf("expected oauth ID %q, got %q", tt.oauthID, thirdPartyUser.OauthID)
			}
		})
	}
}

func TestLocalUser_CheckPassword(t *testing.T) {
	password := "MySecretPassword123!"
	localUser, err := NewLocalUser(password)
	if err != nil {
		t.Fatalf("failed to create local user: %v", err)
	}

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "correct password",
			password: password,
			expected: true,
		},
		{
			name:     "incorrect password",
			password: "WrongPassword",
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			expected: false,
		},
		{
			name:     "similar but different password",
			password: "MySecretPassword123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := localUser.CheckPassword(tt.password)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUserFactory_Create_LocalUser(t *testing.T) {
	factory := &UserFactory{}
	name := "John Doe"
	email := "john@example.com"
	password := "ValidPassword123!"
	workspaceID := uuid.New()

	userAggregate, err := factory.Create(nil, nil, name, email, &password, workspaceID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userAggregate.Name != name {
		t.Errorf("expected name %q, got %q", name, userAggregate.Name)
	}
	if userAggregate.Email != email {
		t.Errorf("expected email %q, got %q", email, userAggregate.Email)
	}
	if userAggregate.LocalUser == nil {
		t.Fatal("expected LocalUser to be set")
	}
	if userAggregate.ThirdPartyUser != nil {
		t.Error("expected ThirdPartyUser to be nil")
	}
	if !userAggregate.LocalUser.CheckPassword(password) {
		t.Error("password verification failed")
	}
}

func TestUserFactory_Create_ThirdPartyUser(t *testing.T) {
	factory := &UserFactory{}
	name := "Jane Doe"
	email := "jane@example.com"
	oauthProvider := OauthProviderGitHub
	oauthID := uuid.New()
	workspaceID := uuid.New()

	userAggregate, err := factory.Create(&oauthProvider, &oauthID, name, email, nil, workspaceID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userAggregate.Name != name {
		t.Errorf("expected name %q, got %q", name, userAggregate.Name)
	}
	if userAggregate.Email != email {
		t.Errorf("expected email %q, got %q", email, userAggregate.Email)
	}
	if userAggregate.ThirdPartyUser == nil {
		t.Fatal("expected ThirdPartyUser to be set")
	}
	if userAggregate.LocalUser != nil {
		t.Error("expected LocalUser to be nil")
	}
	if userAggregate.ThirdPartyUser.OauthProvider != oauthProvider {
		t.Errorf("expected oauth provider %q, got %q", oauthProvider, userAggregate.ThirdPartyUser.OauthProvider)
	}
}

func TestUserFactory_Create_NoAuthMethod(t *testing.T) {
	factory := &UserFactory{}
	name := "Test User"
	email := "test@example.com"
	workspaceID := uuid.New()

	userAggregate, err := factory.Create(nil, nil, name, email, nil, workspaceID)

	if err == nil {
		t.Fatal("expected error but got none")
	}
	if userAggregate.ID != uuid.Nil {
		t.Error("expected empty user aggregate")
	}
	if err.HTTPStatus() != 400 {
		t.Errorf("expected HTTP status 400, got %d", err.HTTPStatus())
	}
}

func TestUserFactory_Create_PartialOAuthCredentials(t *testing.T) {
	factory := &UserFactory{}
	name := "Test User"
	email := "test@example.com"
	workspaceID := uuid.New()

	t.Run("provider without ID", func(t *testing.T) {
		oauthProvider := OauthProviderGitHub
		userAggregate, err := factory.Create(&oauthProvider, nil, name, email, nil, workspaceID)

		if err == nil {
			t.Fatal("expected error but got none")
		}
		if userAggregate.ID != uuid.Nil {
			t.Error("expected empty user aggregate")
		}
	})

	t.Run("ID without provider", func(t *testing.T) {
		oauthID := uuid.New()
		userAggregate, err := factory.Create(nil, &oauthID, name, email, nil, workspaceID)

		if err == nil {
			t.Fatal("expected error but got none")
		}
		if userAggregate.ID != uuid.Nil {
			t.Error("expected empty user aggregate")
		}
	})
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "normal password",
			password: "Password123!",
		},
		{
			name:     "short password",
			password: "abc",
		},
		{
			name:     "long password",
			password: string(make([]byte, 100)),
		},
		{
			name:     "special characters",
			password: "!@#$%^&*()_+-={}[]|:;<>?,./",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashed, err := hashPassword(tt.password)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if hashed == "" {
				t.Error("expected non-empty hashed password")
			}
			if hashed == tt.password {
				t.Error("password should be hashed, not plain text")
			}

			valid, checkErr := verifyArgon2idHash(tt.password, hashed)
			if checkErr != nil {
				t.Errorf("error verifying password: %v", checkErr)
			}
			if !valid {
				t.Error("hashed password does not match original password")
			}

			if err != nil && err.HTTPStatus() != 500 {
				t.Errorf("expected HTTP status 500, got %d", err.HTTPStatus())
			}
		})
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hashed, err := hashPassword("")

	if err != nil {
		t.Errorf("unexpected error for empty password: %v", err)
	}
	if hashed == "" {
		t.Error("expected non-empty hashed password even for empty input")
	}
}

func TestOAuthProviderConstants(t *testing.T) {
	if OauthProviderGitHub != "github" {
		t.Errorf("expected OauthProviderGitHub to be 'github', got %q", OauthProviderGitHub)
	}
	if OauthProviderGoogle != "google" {
		t.Errorf("expected OauthProviderGoogle to be 'google', got %q", OauthProviderGoogle)
	}
}

func TestUserFactory_Create_BothAuthMethods(t *testing.T) {
	factory := &UserFactory{}
	name := "Test User"
	email := "test@example.com"
	password := "Password123!"
	oauthProvider := OauthProviderGitHub
	oauthID := uuid.New()
	workspaceID := uuid.New()

	userAggregate, err := factory.Create(&oauthProvider, &oauthID, name, email, &password, workspaceID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userAggregate.ThirdPartyUser == nil {
		t.Error("expected ThirdPartyUser to be set when both auth methods provided")
	}
	if userAggregate.LocalUser != nil {
		t.Error("expected LocalUser to be nil when OAuth credentials are provided")
	}
}
