package integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Response structs
type WorkspaceResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AdminID     uuid.UUID `json:"admin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserResponse struct {
	Message string    `json:"message"`
	UserID  uuid.UUID `json:"user_id"`
}

type ErrorResponse struct {
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AuthContext holds the identity claims for generating a per-request JWT token.
type AuthContext struct {
	UserID      uuid.UUID
	UserName    string
	WorkspaceID uuid.UUID
}

// Teardown

func TearDownWorkspace(t *testing.T, workspaceName string) {
	t.Helper()
	DbConnection.Exec("DELETE FROM workspaces WHERE name = $1", workspaceName)
}

// GetWorkspaceFromDB fetches a workspace directly from the DB, bypassing HTTP auth.
func GetWorkspaceFromDB(t *testing.T, id uuid.UUID) *WorkspaceResponse {
	t.Helper()
	var w WorkspaceResponse
	err := DbConnection.QueryRow(
		"SELECT id, name, description, admin_id, created_at, updated_at FROM workspaces WHERE id = $1",
		id,
	).Scan(&w.ID, &w.Name, &w.Description, &w.AdminID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		t.Fatalf("GetWorkspaceFromDB: %v", err)
	}
	return &w
}

// addAuth generates a fresh JWT token from auth and attaches it as a cookie.
func addAuth(t *testing.T, req *http.Request, auth AuthContext) {
	t.Helper()
	token, err := jwtSvc.GenerateToken(
		auth.UserID.String(),
		auth.UserName,
		auth.WorkspaceID.String(),
	)
	if err != nil {
		t.Fatalf("addAuth: failed to generate JWT token: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
}

// Workspace helpers

func CreateWorkspace(t *testing.T, auth AuthContext, name, description string, adminID uuid.UUID) (*WorkspaceResponse, int) {
	t.Helper()

	payload := map[string]interface{}{
		"name":        name,
		"description": description,
		"admin_id":    adminID,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, BaseURL+"/api/v1/workspaces", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var workspace WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
			t.Fatalf("failed to decode workspace response: %v", err)
		}
		return &workspace, resp.StatusCode
	}

	return nil, resp.StatusCode
}

func GetWorkspace(t *testing.T, auth AuthContext, id uuid.UUID) (*WorkspaceResponse, int) {
	t.Helper()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/workspaces/%s", BaseURL, id), nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get workspace: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var workspace WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
			t.Fatalf("failed to decode workspace response: %v", err)
		}
		return &workspace, resp.StatusCode
	}

	return nil, resp.StatusCode
}

func GetWorkspacesByAdmin(t *testing.T, auth AuthContext, adminID uuid.UUID) ([]*WorkspaceResponse, int) {
	t.Helper()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/workspaces/admin/%s", BaseURL, adminID), nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get workspaces by admin: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var workspaces []*WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&workspaces); err != nil {
			t.Fatalf("failed to decode workspaces response: %v", err)
		}
		return workspaces, resp.StatusCode
	}

	return nil, resp.StatusCode
}

func UpdateWorkspace(t *testing.T, auth AuthContext, id uuid.UUID, name, description string) (*WorkspaceResponse, int) {
	t.Helper()

	payload := map[string]interface{}{}
	if name != "" {
		payload["name"] = name
	}
	if description != "" {
		payload["description"] = description
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/workspaces/%s", BaseURL, id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to update workspace: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var workspace WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
			t.Fatalf("failed to decode workspace response: %v", err)
		}
		return &workspace, resp.StatusCode
	}

	return nil, resp.StatusCode
}

func DeleteWorkspace(t *testing.T, auth AuthContext, id uuid.UUID) int {
	t.Helper()

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/workspaces/%s", BaseURL, id), nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to delete workspace: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

func ListWorkspaces(t *testing.T, auth AuthContext, limit, offset int, sortBy, order string) ([]*WorkspaceResponse, int) {
	t.Helper()

	url := fmt.Sprintf("%s/api/v1/workspaces?limit=%d&offset=%d", BaseURL, limit, offset)
	if sortBy != "" {
		url += "&sort_by=" + sortBy
	}
	if order != "" {
		url += "&order=" + order
	}

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to list workspaces: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var workspaces []*WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&workspaces); err != nil {
			t.Fatalf("failed to decode workspaces response: %v", err)
		}
		return workspaces, resp.StatusCode
	}

	return nil, resp.StatusCode
}

// User helpers

func CreateUser(t *testing.T, name, email, password string, workspaceID uuid.UUID) (*UserResponse, int) {
	t.Helper()

	payload := map[string]interface{}{
		"name":         name,
		"email":        email,
		"password":     password,
		"workspace_id": workspaceID,
	}

	body, _ := json.Marshal(payload)
	resp, err := HTTPClient.Post(
		BaseURL+"/api/v1/users",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var user UserResponse
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode user response: %v", err)
		}
		return &user, resp.StatusCode
	}

	return nil, resp.StatusCode
}

// ReadErrorResponse reads and decodes an error response
func ReadErrorResponse(t *testing.T, resp *http.Response) *ErrorResponse {
	t.Helper()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read error response: %v", err)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(bodyBytes, &errResp); err != nil {
		t.Logf("raw error response: %s", string(bodyBytes))
		t.Fatalf("failed to decode error response: %v", err)
	}

	return &errResp
}

// Template helpers

type TemplateResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func CreateTemplate(t *testing.T, auth AuthContext, name string, workspaceID uuid.UUID, path string) (*TemplateResponse, int) {
	t.Helper()

	payload := map[string]interface{}{
		"name":         name,
		"workspace_id": workspaceID,
		"path":         path,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, BaseURL+"/api/v1/templates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var template TemplateResponse
		if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
			t.Fatalf("failed to decode template response: %v", err)
		}
		return &template, resp.StatusCode
	}

	return nil, resp.StatusCode
}

func GetTemplate(t *testing.T, auth AuthContext, id uuid.UUID) (*TemplateResponse, int) {
	t.Helper()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/templates/%s", BaseURL, id), nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get template: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var template TemplateResponse
		if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
			t.Fatalf("failed to decode template response: %v", err)
		}
		return &template, resp.StatusCode
	}

	return nil, resp.StatusCode
}

func GetTemplatesByWorkspace(t *testing.T, auth AuthContext, workspaceID uuid.UUID) ([]*TemplateResponse, int) {
	t.Helper()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/templates/workspace/%s", BaseURL, workspaceID), nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get templates by workspace: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var templates []*TemplateResponse
		if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
			t.Fatalf("failed to decode templates response: %v", err)
		}
		return templates, resp.StatusCode
	}

	return nil, resp.StatusCode
}

func UpdateTemplate(t *testing.T, auth AuthContext, id uuid.UUID, name, path string) (*TemplateResponse, int) {
	t.Helper()

	payload := map[string]interface{}{}
	if name != "" {
		payload["name"] = name
	}
	if path != "" {
		payload["path"] = path
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/templates/%s", BaseURL, id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to update template: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var template TemplateResponse
		if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
			t.Fatalf("failed to decode template response: %v", err)
		}
		return &template, resp.StatusCode
	}

	return nil, resp.StatusCode
}

func DeleteTemplate(t *testing.T, auth AuthContext, id uuid.UUID) int {
	t.Helper()

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/templates/%s", BaseURL, id), nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to delete template: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode
}

func ListTemplates(t *testing.T, auth AuthContext, limit, offset int, sortBy, order string) ([]*TemplateResponse, int) {
	t.Helper()

	url := fmt.Sprintf("%s/api/v1/templates?limit=%d&offset=%d", BaseURL, limit, offset)
	if sortBy != "" {
		url += "&sort_by=" + sortBy
	}
	if order != "" {
		url += "&order=" + order
	}

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	addAuth(t, req, auth)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to list templates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var templates []*TemplateResponse
		if err := json.NewDecoder(resp.Body).Decode(&templates); err != nil {
			t.Fatalf("failed to decode templates response: %v", err)
		}
		return templates, resp.StatusCode
	}

	return nil, resp.StatusCode
}

// Admin helpers

type AdminInitResponse struct {
	Message     string    `json:"message"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	AdminUserID uuid.UUID `json:"admin_user_id"`
}

func InitializeAdmin(t *testing.T, adminName, adminEmail, adminPassword, workspaceName, workspaceDescription, token string) (*AdminInitResponse, int) {
	t.Helper()

	payload := map[string]interface{}{
		"admin_name":            adminName,
		"admin_email":           adminEmail,
		"admin_password":        adminPassword,
		"workspace_name":        workspaceName,
		"workspace_description": workspaceDescription,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, BaseURL+"/admin/init", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Admin-Init-Token", token)
	}

	resp, err := HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("failed to initialize admin: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		var adminResp AdminInitResponse
		if err := json.NewDecoder(resp.Body).Decode(&adminResp); err != nil {
			t.Fatalf("failed to decode admin init response: %v", err)
		}
		return &adminResp, resp.StatusCode
	}

	return nil, resp.StatusCode
}
