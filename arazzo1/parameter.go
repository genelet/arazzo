package arazzo1

import (
	"encoding/json"
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
	Name string `json:"name"`

	// In is the named location of the parameter (path, query, header, cookie).
	// Required when used in operation steps.
	In ParameterIn `json:"in,omitempty"`

	// Value is the value to pass in the parameter (required).
	// Can be string, boolean, object, array, number, or null.
	Value any `json:"value"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-"`
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
