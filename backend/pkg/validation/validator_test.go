package validation

import (
	"testing"

	"backend/pkg/contracts"

	"github.com/google/uuid"
)

func TestValidator_ValidContract(t *testing.T) {
	validator := New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	validRequest := contracts.CreateLocalUser{
		Name:        "John Doe",
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		WorkspaceID: uuid.New(),
	}

	err := validator.Validate(validRequest)
	if err != nil {
		t.Errorf("Expected no error for valid request, got: %v", err)
	}
}

func TestValidator_RequiredField(t *testing.T) {
	validator := New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	invalidRequest := contracts.CreateLocalUser{
		// Name is missing
		Email:       "john@example.com",
		Password:    "SecurePass123!",
		WorkspaceID: uuid.New(),
	}

	err := validator.Validate(invalidRequest)
	if err == nil {
		t.Error("Expected validation error for missing required field")
		return
	}

	// Check that error metadata contains the field error
	metadata := err.GetMetadata()
	fields, ok := metadata["fields"].(map[string]string)
	if !ok {
		t.Errorf("Expected fields in metadata, got: %v", metadata)
		return
	}

	if _, exists := fields["name"]; !exists {
		t.Errorf("Expected 'name' field error, got fields: %v", fields)
	}
}

func TestValidator_EmailValidation(t *testing.T) {
	validator := New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	tests := []struct {
		name      string
		email     string
		wantError bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"invalid - no @", "notanemail", true},
		{"invalid - no domain", "test@", true},
		{"invalid - no local part", "@example.com", true},
		{"empty", "", true}, // Also fails 'required'
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := contracts.CreateLocalUser{
				Name:        "John Doe",
				Email:       tt.email,
				Password:    "SecurePass123!",
				WorkspaceID: uuid.New(),
			}

			err := validator.Validate(request)
			if tt.wantError && err == nil {
				t.Errorf("Expected validation error for email: %s", tt.email)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error for email: %s, got: %v", tt.email, err)
			}
		})
	}
}

func TestValidator_MinLengthValidation(t *testing.T) {
	validator := New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{"valid - 8 chars with requirements", "Pass123!", false},
		{"valid - longer", "VeryLongPassword123!", false},
		{"invalid - too short", "Pass1!", true},
		{"invalid - empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := contracts.CreateLocalUser{
				Name:        "John Doe",
				Email:       "john@example.com",
				Password:    tt.password,
				WorkspaceID: uuid.New(),
			}

			err := validator.Validate(request)
			if tt.wantError && err == nil {
				t.Errorf("Expected validation error for password: %s", tt.password)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error for password: %s, got: %v", tt.password, err)
			}
		})
	}
}

func TestValidator_UUID4Validation(t *testing.T) {
	validator := New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	tests := []struct {
		name        string
		workspaceID uuid.UUID
		wantError   bool
	}{
		{"valid UUID v4", uuid.New(), false},
		{"nil UUID", uuid.Nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := contracts.CreateLocalUser{
				Name:        "John Doe",
				Email:       "john@example.com",
				Password:    "SecurePass123!",
				WorkspaceID: tt.workspaceID,
			}

			err := validator.Validate(request)
			if tt.wantError && err == nil {
				t.Error("Expected validation error for invalid UUID")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestValidator_MultipleFieldErrors(t *testing.T) {
	validator := New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	// Request with multiple validation failures
	invalidRequest := contracts.CreateLocalUser{
		// Name is missing (required)
		Email:       "not-an-email", // invalid email
		Password:    "short",         // too short (min=8)
		WorkspaceID: uuid.Nil,        // invalid UUID
	}

	err := validator.Validate(invalidRequest)
	if err == nil {
		t.Fatal("Expected validation error for multiple invalid fields")
	}

	metadata := err.GetMetadata()
	fields, ok := metadata["fields"].(map[string]string)
	if !ok {
		t.Fatalf("Expected fields in metadata, got: %v", metadata)
	}

	// Should have errors for all invalid fields
	expectedFields := []string{"name", "email", "password", "workspace_id"}
	for _, field := range expectedFields {
		if _, exists := fields[field]; !exists {
			t.Errorf("Expected error for field '%s', got fields: %v", field, fields)
		}
	}
}

func TestValidator_StrongPassword(t *testing.T) {
	validator := New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	type testStruct struct {
		Password string `json:"password" validate:"required,min=8,strongpassword"`
	}

	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{"valid - all requirements", "SecurePass123!", false},
		{"valid - different special char", "MyP@ssw0rd", false},
		{"invalid - no uppercase", "securepass123!", true},
		{"invalid - no lowercase", "SECUREPASS123!", true},
		{"invalid - no digit", "SecurePass!!", true},
		{"invalid - no special char", "SecurePass123", true},
		{"invalid - too short", "Sec1!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := testStruct{Password: tt.password}
			err := validator.Validate(data)

			if tt.wantError && err == nil {
				t.Errorf("Expected validation error for password: %s", tt.password)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error for password: %s, got: %v", tt.password, err)
			}
		})
	}
}

func TestValidator_ErrorMessages(t *testing.T) {
	validator := New()
	if err := validator.RegisterDefaultCustomValidations(); err != nil {
		t.Fatalf("Failed to register custom validations: %v", err)
	}

	invalidRequest := contracts.CreateLocalUser{
		Name:        "", // required
		Email:       "bademail",
		Password:    "short",
		WorkspaceID: uuid.Nil,
	}

	err := validator.Validate(invalidRequest)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	metadata := err.GetMetadata()
	fields, ok := metadata["fields"].(map[string]string)
	if !ok {
		t.Fatalf("Expected fields in metadata")
	}

	// Verify error messages are human-readable
	if msg, exists := fields["name"]; exists {
		if msg == "" {
			t.Error("Expected non-empty error message for name")
		}
	}

	if msg, exists := fields["email"]; exists {
		if msg == "" {
			t.Error("Expected non-empty error message for email")
		}
	}
}
