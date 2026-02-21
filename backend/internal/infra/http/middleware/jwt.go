package middleware

import (
	"context"
	"strings"

	domainerrors "backend/internal/domain/errors"
	"backend/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type contextKeyType string

const ClaimsKey contextKeyType = "claims"

// RequireAuth returns a Fiber middleware that validates the Bearer JWT token
// from the Authorization header and stores the claims in context locals.
func RequireAuth(jwtService *jwt.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" {
			return domainerrors.Unauthorized("missing authorization header")
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			return domainerrors.Unauthorized("authorization header must use Bearer scheme")
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			return domainerrors.Unauthorized("bearer token is empty")
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			return err
		}

		c.Locals(ClaimsKey, claims)
		return c.Next()
	}
}

// GetClaims retrieves the JWT claims stored by RequireAuth from the Fiber context.
// Returns (nil, false) if called on an unprotected route.
func GetClaims(c *fiber.Ctx) (*jwt.Claims, bool) {
	claims, ok := c.Locals(ClaimsKey).(*jwt.Claims)
	return claims, ok
}

// ContextWithClaims returns c.Context() enriched with JWT claims so the application
// layer can call jwt.ClaimsFromContext without any Fiber dependency.
// If no claims are present (unprotected route), the original context is returned unchanged.
func ContextWithClaims(c *fiber.Ctx) context.Context {
	claims, ok := GetClaims(c)
	if !ok {
		return c.Context()
	}
	return jwt.WithClaims(c.Context(), claims)
}
