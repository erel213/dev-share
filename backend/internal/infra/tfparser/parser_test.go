package tfparser

import (
	"testing"
)

func TestParseSimpleVariable(t *testing.T) {
	content := []byte(`
variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}
`)
	parser := NewHCLParser()
	vars, err := parser.ParseVariables(content, "variables.tf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(vars))
	}

	v := vars[0]
	if v.Key != "region" {
		t.Errorf("expected key 'region', got %q", v.Key)
	}
	if v.Description != "AWS region" {
		t.Errorf("expected description 'AWS region', got %q", v.Description)
	}
	if v.VarType != "string" {
		t.Errorf("expected type 'string', got %q", v.VarType)
	}
	if v.IsRequired {
		t.Error("expected IsRequired=false (has default)")
	}
	if v.IsSensitive {
		t.Error("expected IsSensitive=false")
	}
}

func TestParseSensitiveVariable(t *testing.T) {
	content := []byte(`
variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}
`)
	parser := NewHCLParser()
	vars, err := parser.ParseVariables(content, "variables.tf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(vars))
	}

	v := vars[0]
	if !v.IsSensitive {
		t.Error("expected IsSensitive=true")
	}
	if !v.IsRequired {
		t.Error("expected IsRequired=true (no default)")
	}
}

func TestParseMultipleVariables(t *testing.T) {
	content := []byte(`
variable "region" {
  description = "AWS region"
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
}

variable "api_key" {
  sensitive = true
}
`)
	parser := NewHCLParser()
	vars, err := parser.ParseVariables(content, "variables.tf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 3 {
		t.Fatalf("expected 3 variables, got %d", len(vars))
	}
}

func TestParseVariableWithNoAttributes(t *testing.T) {
	content := []byte(`
variable "simple" {}
`)
	parser := NewHCLParser()
	vars, err := parser.ParseVariables(content, "variables.tf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 variable, got %d", len(vars))
	}

	v := vars[0]
	if v.Key != "simple" {
		t.Errorf("expected key 'simple', got %q", v.Key)
	}
	if v.VarType != "string" {
		t.Errorf("expected default type 'string', got %q", v.VarType)
	}
	if !v.IsRequired {
		t.Error("expected IsRequired=true")
	}
}

func TestParseNoVariables(t *testing.T) {
	content := []byte(`
resource "aws_instance" "example" {
  ami           = "ami-12345"
  instance_type = "t2.micro"
}
`)
	parser := NewHCLParser()
	vars, err := parser.ParseVariables(content, "main.tf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(vars))
	}
}

func TestParseMalformedHCL(t *testing.T) {
	content := []byte(`this is not valid HCL {{{`)
	parser := NewHCLParser()
	_, err := parser.ParseVariables(content, "bad.tf")
	if err == nil {
		t.Fatal("expected error for malformed HCL")
	}
}

func TestParseEmptyContent(t *testing.T) {
	parser := NewHCLParser()
	vars, err := parser.ParseVariables([]byte(""), "empty.tf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(vars))
	}
}

func TestParseVariableRequiredWithoutDefault(t *testing.T) {
	content := []byte(`
variable "required_var" {
  type = string
}
`)
	parser := NewHCLParser()
	vars, err := parser.ParseVariables(content, "variables.tf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !vars[0].IsRequired {
		t.Error("variable without default should be required")
	}
}

func TestParseNonSensitiveExplicit(t *testing.T) {
	content := []byte(`
variable "public_var" {
  sensitive = false
}
`)
	parser := NewHCLParser()
	vars, err := parser.ParseVariables(content, "variables.tf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars[0].IsSensitive {
		t.Error("expected IsSensitive=false")
	}
}
