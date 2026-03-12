package tfparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type ParsedVariable struct {
	Key         string
	Description string
	VarType     string
	Default     string
	IsSensitive bool
	IsRequired  bool
}

type TFParser interface {
	// ParseVariables parses the content of a variables.tf file
	// and returns the extracted variable definitions.
	// The filename parameter is used only for error reporting.
	ParseVariables(content []byte, filename string) ([]ParsedVariable, error)
}

type HCLParser struct{}

func NewHCLParser() *HCLParser {
	return &HCLParser{}
}

func (p *HCLParser) ParseVariables(content []byte, filename string) ([]ParsedVariable, error) {
	file, diags := hclsyntax.ParseConfig(content, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse %s: %s", filename, diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, nil
	}

	var variables []ParsedVariable
	for _, block := range body.Blocks {
		if block.Type != "variable" || len(block.Labels) == 0 {
			continue
		}

		v := ParsedVariable{
			Key:        block.Labels[0],
			VarType:    "string",
			IsRequired: true,
		}

		attrs, attrDiags := block.Body.JustAttributes()
		if attrDiags.HasErrors() {
			variables = append(variables, v)
			continue
		}

		if attr, exists := attrs["description"]; exists {
			val, diags := attr.Expr.Value(nil)
			if !diags.HasErrors() {
				v.Description = val.AsString()
			}
		}

		if attr, exists := attrs["type"]; exists {
			v.VarType = string(content[attr.Expr.Range().Start.Byte:attr.Expr.Range().End.Byte])
		}

		if attr, exists := attrs["default"]; exists {
			v.IsRequired = false
			val, diags := attr.Expr.Value(nil)
			if !diags.HasErrors() {
				v.Default = val.GoString()
			}
		}

		if attr, exists := attrs["sensitive"]; exists {
			val, diags := attr.Expr.Value(nil)
			if !diags.HasErrors() && val.True() {
				v.IsSensitive = true
			}
		}

		variables = append(variables, v)
	}

	return variables, nil
}
