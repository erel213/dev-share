package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	handlererrors "backend/internal/application/errors"
	"backend/internal/domain"
	"backend/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

const testSecret = "this-is-a-very-secure-secret-key-for-testing-purposes"

func setupTestApp(minRole domain.Role) *fiber.App {
	jwtService, _ := jwt.NewService(testSecret)

	app := fiber.New(fiber.Config{
		ErrorHandler: handlererrors.ErrorHandler(),
	})

	app.Use(RequireAuth(jwtService, jwt.DefaultCookieConfig()))
	app.Use(RequireRole(minRole))

	app.Get("/resource", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Post("/resource", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Put("/resource", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Delete("/resource", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	return app
}

func generateToken(t *testing.T, role string) string {
	t.Helper()
	svc, _ := jwt.NewService(testSecret)
	token, err := svc.GenerateToken("user-1", "Test User", role, "workspace-1")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return token
}

func doRequest(t *testing.T, app *fiber.App, method, path, token string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	return resp
}

func TestRequireRole_GETAllowsAllRoles(t *testing.T) {
	app := setupTestApp(domain.RoleAdmin)

	roles := []string{"admin", "editor", "user"}
	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			token := generateToken(t, role)
			resp := doRequest(t, app, http.MethodGet, "/resource", token)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("GET with role %q: expected 200, got %d", role, resp.StatusCode)
			}
		})
	}
}

func TestRequireRole_EditorMinRole(t *testing.T) {
	app := setupTestApp(domain.RoleEditor)

	tests := []struct {
		role           string
		method         string
		expectedStatus int
	}{
		// Admin can write
		{"admin", http.MethodPost, http.StatusOK},
		{"admin", http.MethodPut, http.StatusOK},
		{"admin", http.MethodDelete, http.StatusOK},
		// Editor can write
		{"editor", http.MethodPost, http.StatusOK},
		{"editor", http.MethodPut, http.StatusOK},
		{"editor", http.MethodDelete, http.StatusOK},
		// User cannot write
		{"user", http.MethodPost, http.StatusForbidden},
		{"user", http.MethodPut, http.StatusForbidden},
		{"user", http.MethodDelete, http.StatusForbidden},
		// User can still GET
		{"user", http.MethodGet, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.role+"_"+tt.method, func(t *testing.T) {
			token := generateToken(t, tt.role)
			resp := doRequest(t, app, tt.method, "/resource", token)

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("%s %s with role %q: expected %d, got %d (body: %s)",
					tt.method, "/resource", tt.role, tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

func TestRequireRole_AdminMinRole(t *testing.T) {
	app := setupTestApp(domain.RoleAdmin)

	tests := []struct {
		role           string
		method         string
		expectedStatus int
	}{
		// Admin can write
		{"admin", http.MethodPost, http.StatusOK},
		{"admin", http.MethodPut, http.StatusOK},
		{"admin", http.MethodDelete, http.StatusOK},
		// Editor cannot write
		{"editor", http.MethodPost, http.StatusForbidden},
		{"editor", http.MethodPut, http.StatusForbidden},
		{"editor", http.MethodDelete, http.StatusForbidden},
		// User cannot write
		{"user", http.MethodPost, http.StatusForbidden},
		{"user", http.MethodPut, http.StatusForbidden},
		{"user", http.MethodDelete, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.role+"_"+tt.method, func(t *testing.T) {
			token := generateToken(t, tt.role)
			resp := doRequest(t, app, tt.method, "/resource", token)

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("%s with role %q: expected %d, got %d (body: %s)",
					tt.method, tt.role, tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

func TestRequireRole_UserMinRole(t *testing.T) {
	app := setupTestApp(domain.RoleUser)

	// All roles can write when min role is user
	roles := []string{"admin", "editor", "user"}
	for _, role := range roles {
		t.Run(role+"_POST", func(t *testing.T) {
			token := generateToken(t, role)
			resp := doRequest(t, app, http.MethodPost, "/resource", token)

			if resp.StatusCode != http.StatusOK {
				t.Errorf("POST with role %q: expected 200, got %d", role, resp.StatusCode)
			}
		})
	}
}

func TestRequireRole_NoAuth(t *testing.T) {
	app := setupTestApp(domain.RoleEditor)

	resp := doRequest(t, app, http.MethodPost, "/resource", "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
}
