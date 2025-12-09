package arazzo1

import (
	"encoding/json"
)

// SourceDescriptionType represents the type of source description.
type SourceDescriptionType string

const (
	// SourceDescriptionTypeArazzo indicates the source is an Arazzo document.
	SourceDescriptionTypeArazzo SourceDescriptionType = "arazzo"

	// SourceDescriptionTypeOpenAPI indicates the source is an OpenAPI document.
	SourceDescriptionTypeOpenAPI SourceDescriptionType = "openapi"
)

// SourceDescription describes a source description (such as an OpenAPI description)
// that will be referenced by one or more workflows described within an Arazzo description.
type SourceDescription struct {
	// Name is a unique name for the source description (required)
	// Pattern: ^[A-Za-z0-9_\-]+$
	Name string `json:"name"`

	// URL is a URL to a source description to be used by a workflow (required)
	// Format: uri-reference
	URL string `json:"url"`

	// Type is the type of source description (arazzo or openapi)
	Type SourceDescriptionType `json:"type,omitempty"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-"`
}

type sourceDescriptionAlias SourceDescription

var sourceDescriptionKnownFields = []string{
	"name",
	"url",
	"type",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *SourceDescription) UnmarshalJSON(data []byte) error {
	var alias sourceDescriptionAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*s = SourceDescription(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Extensions = extractExtensions(raw, sourceDescriptionKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (s SourceDescription) MarshalJSON() ([]byte, error) {
	alias := sourceDescriptionAlias(s)
	return marshalWithExtensions(&alias, s.Extensions)
}
