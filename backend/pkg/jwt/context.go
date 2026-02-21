package jwt

import "context"

type claimsKeyType string

const claimsContextKey claimsKeyType = "claims"

// WithClaims returns a new context carrying the provided JWT claims.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// ClaimsFromContext extracts JWT claims stored by WithClaims.
// Returns (nil, false) if no claims are present.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}
