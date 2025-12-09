package arazzo1

import (
	"encoding/json"
)

// RequestBody represents the request body to pass to an operation
// as referenced by operationId or operationPath.
type RequestBody struct {
	// ContentType is the Content-Type for the request content.
	ContentType string `json:"contentType,omitempty" yaml:"contentType,omitempty" hcl:"contentType,optional"`

	// Payload is the actual payload (can be any JSON value).
	Payload any `json:"payload,omitempty" yaml:"payload,omitempty" hcl:"payload,optional"`

	// Replacements is a list of locations and values to set within a payload.
	Replacements []*PayloadReplacement `json:"replacements,omitempty" yaml:"replacements,omitempty" hcl:"replacement,block"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type requestBodyAlias RequestBody

var requestBodyKnownFields = []string{
	"contentType",
	"payload",
	"replacements",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *RequestBody) UnmarshalJSON(data []byte) error {
	var alias requestBodyAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*r = RequestBody(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Extensions = extractExtensions(raw, requestBodyKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (r RequestBody) MarshalJSON() ([]byte, error) {
	alias := requestBodyAlias(r)
	return marshalWithExtensions(&alias, r.Extensions)
}

// PayloadReplacement describes a location within a payload (e.g., a request body)
// and a value to set within the location.
type PayloadReplacement struct {
	// Target is a JSON Pointer or XPath Expression which MUST be resolved
	// against the request body (required).
	Target string `json:"target" yaml:"target" hcl:"target"`

	// Value is the value set within the target location (required).
	Value string `json:"value" yaml:"value" hcl:"value"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type payloadReplacementAlias PayloadReplacement

var payloadReplacementKnownFields = []string{
	"target",
	"value",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *PayloadReplacement) UnmarshalJSON(data []byte) error {
	var alias payloadReplacementAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*p = PayloadReplacement(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.Extensions = extractExtensions(raw, payloadReplacementKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (p PayloadReplacement) MarshalJSON() ([]byte, error) {
	alias := payloadReplacementAlias(p)
	return marshalWithExtensions(&alias, p.Extensions)
}
