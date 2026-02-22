package jwt

import "time"

const DefaultCookieName = "access_token"

// CookieConfig holds framework-agnostic settings for the JWT cookie.
// Use DefaultCookieConfig for secure production defaults.
type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	MaxAge   int       // seconds; takes priority when non-zero
	Expires  time.Time // used when MaxAge is 0
	Secure   bool
	HTTPOnly bool
	SameSite string // "Strict", "Lax", or "None"
}

// DefaultCookieConfig returns a CookieConfig with secure defaults:
// HttpOnly=true, Secure=true, SameSite=Strict, 24h MaxAge, path "/".
func DefaultCookieConfig() CookieConfig {
	return CookieConfig{
		Name:     DefaultCookieName,
		Path:     "/",
		MaxAge:   int(DefaultTokenDuration.Seconds()),
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	}
}
