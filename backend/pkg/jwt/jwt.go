package jwt

import (
	stderrors "errors"
	"os"
	"time"

	"backend/pkg/errors"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const (
	// MinSecretLength defines the minimum length for JWT secret
	MinSecretLength = 32

	// DefaultTokenDuration is the default expiration time for tokens (24 hours)
	DefaultTokenDuration = 24 * time.Hour
)

var (
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.WithCode(errors.CodeUnauthorized, "invalid token")

	// ErrExpiredToken is returned when the token has expired
	ErrExpiredToken = errors.WithCode(errors.CodeUnauthorized, "token has expired")

	// ErrInvalidSigningMethod is returned when the signing method is not expected
	ErrInvalidSigningMethod = errors.WithCode(errors.CodeUnauthorized, "invalid signing method")

	// ErrMissingSecret is returned when JWT_SECRET environment variable is not set
	ErrMissingSecret = errors.WithCode(errors.CodeInternal, "JWT_SECRET environment variable is not set")

	// ErrWeakSecret is returned when the secret is too short
	ErrWeakSecret = errors.WithCodef(errors.CodeInternal, "JWT_SECRET must be at least %d characters long", MinSecretLength)
)

// Claims represents the JWT claims structure containing user information
type Claims struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WorkspaceID string `json:"workspace_id"`
	jwtlib.RegisteredClaims
}

// Service handles JWT token operations
type Service struct {
	secret []byte
}

// NewService creates a new JWT service with the secret from environment variable
func NewService() (*Service, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, ErrMissingSecret
	}

	if len(secret) < MinSecretLength {
		return nil, ErrWeakSecret
	}

	return &Service{
		secret: []byte(secret),
	}, nil
}

// GenerateToken creates a new JWT token with the provided claims
// Returns the signed token string or an error if token generation fails
func (s *Service) GenerateToken(id, name, workspaceID string) (string, error) {
	now := time.Now()

	claims := Claims{
		ID:          id,
		Name:        name,
		WorkspaceID: workspaceID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(now.Add(DefaultTokenDuration)),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign token")
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token, returning the claims
// Returns an error if the token is invalid, expired, or uses an incorrect signing method
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (interface{}, error) {
		// Verify the signing method to prevent algorithm substitution attacks
		// Only accept HS256 specifically
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, ErrInvalidSigningMethod
		}
		return s.secret, nil
	})

	if err != nil {
		// Check if the error is due to token expiration
		if stderrors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
