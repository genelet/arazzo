package arazzo1

import (
	"encoding/json"
)

// Arazzo represents the root object of an Arazzo 1.0.x document.
// It describes workflows that span multiple APIs.
type Arazzo struct {
	// Arazzo is the version number of the Arazzo Specification (pattern: ^1\.0\.\d+(-.+)?$)
	Arazzo string `json:"arazzo"`

	// Info provides metadata about the Arazzo description
	Info *Info `json:"info"`

	// SourceDescriptions is a list of source descriptions such as Arazzo or OpenAPI
	SourceDescriptions []*SourceDescription `json:"sourceDescriptions"`

	// Workflows is a list of workflows
	Workflows []*Workflow `json:"workflows"`

	// Components holds a set of reusable objects for different aspects of the Arazzo Specification
	Components *Components `json:"components,omitempty"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-"`
}

type arazzoAlias Arazzo

var arazzoKnownFields = []string{
	"arazzo",
	"info",
	"sourceDescriptions",
	"workflows",
	"components",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *Arazzo) UnmarshalJSON(data []byte) error {
	var alias arazzoAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*a = Arazzo(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	a.Extensions = extractExtensions(raw, arazzoKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (a Arazzo) MarshalJSON() ([]byte, error) {
	alias := arazzoAlias(a)
	return marshalWithExtensions(&alias, a.Extensions)
}
