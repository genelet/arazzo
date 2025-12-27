package arazzo1

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ParameterIn represents the location of a parameter.
type ParameterIn string

const (
	// ParameterInPath indicates the parameter is in the URL path.
	ParameterInPath ParameterIn = "path"

	// ParameterInQuery indicates the parameter is in the query string.
	ParameterInQuery ParameterIn = "query"

	// ParameterInHeader indicates the parameter is in the HTTP headers.
	ParameterInHeader ParameterIn = "header"

	// ParameterInCookie indicates the parameter is in cookies.
	ParameterInCookie ParameterIn = "cookie"
)

// Parameter describes a single step parameter.
type Parameter struct {
	// Name is the name of the parameter (required)
	Name string `json:"name" yaml:"name" hcl:"name,label"`

	// In is the named location of the parameter (path, query, header, cookie).
	// Required when used in operation steps.
	In ParameterIn `json:"in,omitempty" yaml:"in,omitempty" hcl:"in,optional"`

	// Value is the value to pass in the parameter (required).
	// Can be string, boolean, object, array, number, or null.
	Value any `json:"value" yaml:"value" hcl:"value"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type parameterAlias Parameter

var parameterKnownFields = []string{
	"name",
	"in",
	"value",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *Parameter) UnmarshalJSON(data []byte) error {
	var alias parameterAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*p = Parameter(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.Extensions = extractExtensions(raw, parameterKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (p Parameter) MarshalJSON() ([]byte, error) {
	alias := parameterAlias(p)
	return marshalWithExtensions(&alias, p.Extensions)
}

// UnmarshalHCL implements the dethcl.Unmarshaler interface.
// This custom unmarshaler handles the Value field which is typed as `any`
// and needs special handling to parse HCL values (especially numbers) into Go values.
func (p *Parameter) UnmarshalHCL(data []byte, labels ...string) error {
	// Parse HCL
	file, diags := hclsyntax.ParseConfig(data, "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("parsing HCL: %w", diags)
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return fmt.Errorf("unexpected HCL body type: %T", file.Body)
	}

	// Set label (name) if provided
	if len(labels) > 0 {
		p.Name = labels[0]
	}

	// Process attributes
	for name, attr := range body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return fmt.Errorf("attribute %q: %w", name, diags)
		}

		switch name {
		case "name":
			p.Name = val.AsString()
		case "in":
			p.In = ParameterIn(val.AsString())
		case "value":
			p.Value = ctyToGo(val)
		}
	}

	return nil
}
