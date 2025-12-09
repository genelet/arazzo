package arazzo1

import (
	"encoding/json"
	"strings"
)

// Components holds a set of reusable objects for different aspects of the Arazzo Specification.
// Component names must match pattern: ^[a-zA-Z0-9\.\-_]+$
type Components struct {
	// Inputs is an object to hold reusable JSON Schema 2020-12 schemas
	// to be referenced from workflow inputs.
	Inputs map[string]any `json:"inputs,omitempty" yaml:"inputs,omitempty" hcl:"inputs,optional"`

	// Parameters is an object to hold reusable Parameter Objects.
	Parameters map[string]*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty" hcl:"parameter,block"`

	// SuccessActions is an object to hold reusable Success Actions Objects.
	SuccessActions map[string]*SuccessAction `json:"successActions,omitempty" yaml:"successActions,omitempty" hcl:"successAction,block"`

	// FailureActions is an object to hold reusable Failure Actions Objects.
	FailureActions map[string]*FailureAction `json:"failureActions,omitempty" yaml:"failureActions,omitempty" hcl:"failureAction,block"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

var componentsKnownFields = []string{
	"inputs",
	"parameters",
	"successActions",
	"failureActions",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *Components) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Unmarshal each known field
	if inputsData, ok := raw["inputs"]; ok {
		if err := json.Unmarshal(inputsData, &c.Inputs); err != nil {
			return err
		}
	}

	if parametersData, ok := raw["parameters"]; ok {
		if err := json.Unmarshal(parametersData, &c.Parameters); err != nil {
			return err
		}
	}

	if successActionsData, ok := raw["successActions"]; ok {
		if err := json.Unmarshal(successActionsData, &c.SuccessActions); err != nil {
			return err
		}
	}

	if failureActionsData, ok := raw["failureActions"]; ok {
		if err := json.Unmarshal(failureActionsData, &c.FailureActions); err != nil {
			return err
		}
	}

	// Extract extensions
	c.Extensions = extractExtensions(raw, componentsKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (c Components) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)

	if len(c.Inputs) > 0 {
		result["inputs"] = c.Inputs
	}
	if len(c.Parameters) > 0 {
		result["parameters"] = c.Parameters
	}
	if len(c.SuccessActions) > 0 {
		result["successActions"] = c.SuccessActions
	}
	if len(c.FailureActions) > 0 {
		result["failureActions"] = c.FailureActions
	}

	// Add extensions
	for key, value := range c.Extensions {
		if strings.HasPrefix(key, "x-") {
			result[key] = value
		}
	}

	return json.Marshal(result)
}
