package middleware

import (
	"context"
	"time"

	domainerrors "backend/internal/domain/errors"
	"backend/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type contextKeyType string

const ClaimsKey contextKeyType = "claims"

// RequireAuth returns a Fiber middleware that validates the JWT token
// from the cookie defined in cfg and stores the claims in context locals.
func RequireAuth(jwtService *jwt.Service, cfg jwt.CookieConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies(cfg.Name)
		if tokenString == "" {
			return domainerrors.Unauthorized("missing auth cookie")
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			return err
		}

		c.Locals(ClaimsKey, claims)
		return c.Next()
	}
}

// SetTokenCookie writes the JWT token as a cookie on the response using the
// settings from cfg.
func SetTokenCookie(c *fiber.Ctx, token string, cfg jwt.CookieConfig) {
	c.Cookie(&fiber.Cookie{
		Name:     cfg.Name,
		Value:    token,
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		MaxAge:   cfg.MaxAge,
		Expires:  cfg.Expires,
		Secure:   cfg.Secure,
		HTTPOnly: cfg.HTTPOnly,
		SameSite: cfg.SameSite,
	})
}

// ClearTokenCookie expires the JWT cookie immediately, effectively logging the
// user out on the client side.
func ClearTokenCookie(c *fiber.Ctx, cfg jwt.CookieConfig) {
	c.Cookie(&fiber.Cookie{
		Name:     cfg.Name,
		Path:     cfg.Path,
		Domain:   cfg.Domain,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		Secure:   cfg.Secure,
		HTTPOnly: cfg.HTTPOnly,
		SameSite: cfg.SameSite,
	})
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
