package jwt

import (
	"errors"
	"os"
	"testing"
	"time"

	apperrors "backend/pkg/errors"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	testSecret      = "this-is-a-very-secure-secret-key-for-testing-purposes"
	testShortSecret = "short"
	testUserID      = "user-123"
	testUserName    = "John Doe"
	testWorkspaceID = "workspace-456"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		envSecret   string
		shouldSetup bool
		wantErr     bool
		expectedErr *apperrors.Error
	}{
		{
			name:        "valid secret",
			envSecret:   testSecret,
			shouldSetup: true,
			wantErr:     false,
		},
		{
			name:        "missing secret",
			envSecret:   "",
			shouldSetup: true,
			wantErr:     true,
			expectedErr: ErrMissingSecret,
		},
		{
			name:        "weak secret",
			envSecret:   testShortSecret,
			shouldSetup: true,
			wantErr:     true,
			expectedErr: ErrWeakSecret,
		},
		{
			name:        "secret at minimum length",
			envSecret:   "12345678901234567890123456789012", // exactly 32 chars
			shouldSetup: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.shouldSetup {
				os.Setenv("JWT_SECRET", tt.envSecret)
				defer os.Unsetenv("JWT_SECRET")
			}

			service, err := NewService()

			if tt.wantErr {
				if err == nil {
					t.Error("NewService() expected error, got nil")
					return
				}
				if tt.expectedErr != nil && !errors.As(err, &tt.expectedErr) {
					t.Errorf("NewService() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewService() unexpected error = %v", err)
				return
			}

			if service == nil {
				t.Error("NewService() returned nil service")
				return
			}

			if string(service.secret) != tt.envSecret {
				t.Errorf("NewService() secret = %v, want %v", string(service.secret), tt.envSecret)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	// Setup service
	os.Setenv("JWT_SECRET", testSecret)
	defer os.Unsetenv("JWT_SECRET")

	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name        string
		id          string
		userName    string
		workspaceID string
		wantErr     bool
	}{
		{
			name:        "valid token generation",
			id:          testUserID,
			userName:    testUserName,
			workspaceID: testWorkspaceID,
			wantErr:     false,
		},
		{
			name:        "empty user ID",
			id:          "",
			userName:    testUserName,
			workspaceID: testWorkspaceID,
			wantErr:     false, // Should still generate token
		},
		{
			name:        "empty workspace ID",
			id:          testUserID,
			userName:    testUserName,
			workspaceID: "",
			wantErr:     false, // Should still generate token
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.id, tt.userName, tt.workspaceID)

			if tt.wantErr {
				if err == nil {
					t.Error("GenerateToken() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateToken() unexpected error = %v", err)
				return
			}

			if token == "" {
				t.Error("GenerateToken() returned empty token")
				return
			}

			// Verify the token can be parsed and contains correct claims
			claims, err := service.ValidateToken(token)
			if err != nil {
				t.Errorf("ValidateToken() failed for generated token: %v", err)
				return
			}

			if claims.ID != tt.id {
				t.Errorf("GenerateToken() ID = %v, want %v", claims.ID, tt.id)
			}
			if claims.Name != tt.userName {
				t.Errorf("GenerateToken() Name = %v, want %v", claims.Name, tt.userName)
			}
			if claims.WorkspaceID != tt.workspaceID {
				t.Errorf("GenerateToken() WorkspaceID = %v, want %v", claims.WorkspaceID, tt.workspaceID)
			}
		})
	}
}

func TestGenerateToken_Expiration(t *testing.T) {
	os.Setenv("JWT_SECRET", testSecret)
	defer os.Unsetenv("JWT_SECRET")

	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	token, err := service.GenerateToken(testUserID, testUserName, testWorkspaceID)
	if err != nil {
		t.Fatalf("GenerateToken() failed: %v", err)
	}

	// Parse token to check expiration
	parsedToken, err := jwtlib.ParseWithClaims(token, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		t.Fatal("Failed to cast claims")
	}

	// Check that expiration is set to approximately 24 hours from now
	expectedExpiration := time.Now().Add(DefaultTokenDuration)
	actualExpiration := claims.ExpiresAt.Time

	timeDiff := actualExpiration.Sub(expectedExpiration)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("Token expiration = %v, want approximately %v (diff: %v)", actualExpiration, expectedExpiration, timeDiff)
	}

	// Check IssuedAt is set
	if claims.IssuedAt == nil {
		t.Error("IssuedAt claim is not set")
	}

	// Check NotBefore is set
	if claims.NotBefore == nil {
		t.Error("NotBefore claim is not set")
	}
}

func TestValidateToken(t *testing.T) {
	os.Setenv("JWT_SECRET", testSecret)
	defer os.Unsetenv("JWT_SECRET")

	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Generate a valid token for testing
	validToken, err := service.GenerateToken(testUserID, testUserName, testWorkspaceID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Generate an expired token
	expiredClaims := Claims{
		ID:          testUserID,
		Name:        testUserName,
		WorkspaceID: testWorkspaceID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwtlib.NewNumericDate(time.Now().Add(-25 * time.Hour)),
		},
	}
	expiredToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte(testSecret))

	// Generate a token with wrong signing method
	wrongMethodClaims := Claims{
		ID:          testUserID,
		Name:        testUserName,
		WorkspaceID: testWorkspaceID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	wrongMethodToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS512, wrongMethodClaims)
	wrongMethodTokenString, _ := wrongMethodToken.SignedString([]byte(testSecret))

	// Generate a token with wrong secret
	wrongSecretToken := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, Claims{
		ID:          testUserID,
		Name:        testUserName,
		WorkspaceID: testWorkspaceID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	wrongSecretTokenString, _ := wrongSecretToken.SignedString([]byte("wrong-secret-key-that-is-different"))

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		expectedErr *apperrors.Error
		validateClaims func(*testing.T, *Claims)
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
			validateClaims: func(t *testing.T, claims *Claims) {
				if claims.ID != testUserID {
					t.Errorf("ID = %v, want %v", claims.ID, testUserID)
				}
				if claims.Name != testUserName {
					t.Errorf("Name = %v, want %v", claims.Name, testUserName)
				}
				if claims.WorkspaceID != testWorkspaceID {
					t.Errorf("WorkspaceID = %v, want %v", claims.WorkspaceID, testWorkspaceID)
				}
			},
		},
		{
			name:        "expired token",
			token:       expiredTokenString,
			wantErr:     true,
			expectedErr: ErrExpiredToken,
		},
		{
			name:        "invalid token format",
			token:       "invalid.token.format",
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "empty token",
			token:       "",
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "malformed token",
			token:       "not-a-jwt-token",
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name:        "wrong signing method",
			token:       wrongMethodTokenString,
			wantErr:     true,
			expectedErr: ErrInvalidSigningMethod,
		},
		{
			name:        "wrong secret",
			token:       wrongSecretTokenString,
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("ValidateToken() expected error, got nil")
					return
				}
				if tt.expectedErr != nil {
					var appErr *apperrors.Error
					if !errors.As(err, &appErr) {
						t.Errorf("ValidateToken() error is not *apperrors.Error: %v", err)
						return
					}
					if appErr.Code() != tt.expectedErr.Code() {
						t.Errorf("ValidateToken() error code = %v, want %v", appErr.Code(), tt.expectedErr.Code())
					}
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() unexpected error = %v", err)
				return
			}

			if claims == nil {
				t.Error("ValidateToken() returned nil claims")
				return
			}

			if tt.validateClaims != nil {
				tt.validateClaims(t, claims)
			}
		})
	}
}

func TestValidateToken_TokenTampering(t *testing.T) {
	os.Setenv("JWT_SECRET", testSecret)
	defer os.Unsetenv("JWT_SECRET")

	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Generate a valid token
	validToken, err := service.GenerateToken(testUserID, testUserName, testWorkspaceID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Tamper with the token by changing a character
	if len(validToken) > 10 {
		tamperedToken := validToken[:len(validToken)-5] + "X" + validToken[len(validToken)-4:]

		_, err = service.ValidateToken(tamperedToken)
		if err == nil {
			t.Error("ValidateToken() should reject tampered token")
			return
		}

		var appErr *apperrors.Error
		if !errors.As(err, &appErr) {
			t.Errorf("ValidateToken() error is not *apperrors.Error: %v", err)
			return
		}

		if appErr.Code() != apperrors.CodeUnauthorized {
			t.Errorf("ValidateToken() error code = %v, want %v", appErr.Code(), apperrors.CodeUnauthorized)
		}
	}
}

func TestService_MultipleTokens(t *testing.T) {
	os.Setenv("JWT_SECRET", testSecret)
	defer os.Unsetenv("JWT_SECRET")

	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Generate multiple tokens and verify they're all valid
	// Note: Tokens with identical claims at the same second will be identical,
	// which is expected behavior. We add delays to ensure different timestamps.
	tokens := make(map[string]bool)
	for i := 0; i < 5; i++ {
		token, err := service.GenerateToken(testUserID, testUserName, testWorkspaceID)
		if err != nil {
			t.Errorf("GenerateToken() iteration %d failed: %v", i, err)
			continue
		}

		tokens[token] = true

		// Validate each token
		claims, err := service.ValidateToken(token)
		if err != nil {
			t.Errorf("ValidateToken() iteration %d failed: %v", i, err)
			continue
		}

		if claims.ID != testUserID {
			t.Errorf("Iteration %d: ID = %v, want %v", i, claims.ID, testUserID)
		}

		// Wait for at least 1 second to ensure different IssuedAt times
		// JWT timestamps have second precision
		time.Sleep(1100 * time.Millisecond)
	}

	// Verify we got unique tokens (since we waited between generations)
	if len(tokens) != 5 {
		t.Errorf("Expected 5 unique tokens, got %d", len(tokens))
	}
}
