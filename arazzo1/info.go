package arazzo1

import (
	"encoding/json"
)

// Info provides metadata about the Arazzo description.
type Info struct {
	// Title is a human readable title of the Arazzo Description (required)
	Title string `json:"title" yaml:"title" hcl:"title"`

	// Summary is a short summary of the Arazzo Description
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty" hcl:"summary,optional"`

	// Description of the purpose of the workflows defined.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty" hcl:"description,optional"`

	// Version is the version identifier of the Arazzo document (required)
	// This is distinct from the Arazzo Specification version.
	Version string `json:"version" yaml:"version" hcl:"version"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type infoAlias Info

var infoKnownFields = []string{
	"title",
	"summary",
	"description",
	"version",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (i *Info) UnmarshalJSON(data []byte) error {
	var alias infoAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*i = Info(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	i.Extensions = extractExtensions(raw, infoKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (i Info) MarshalJSON() ([]byte, error) {
	alias := infoAlias(i)
	return marshalWithExtensions(&alias, i.Extensions)
}
