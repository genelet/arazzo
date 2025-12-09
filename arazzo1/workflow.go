package arazzo1

import (
	"encoding/json"
)

// Workflow describes the steps to be taken across one or more APIs to achieve an objective.
type Workflow struct {
	// WorkflowId is a unique string to represent the workflow (required)
	WorkflowId string `json:"workflowId" yaml:"workflowId" hcl:"workflowId,label"`

	// Summary is a summary of the purpose or objective of the workflow
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty" hcl:"summary,optional"`

	// Description of the workflow. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty" hcl:"description,optional"`

	// Inputs is a JSON Schema 2020-12 object representing the input parameters used by this workflow
	Inputs any `json:"inputs,omitempty" yaml:"inputs,omitempty" hcl:"inputs,optional"`

	// DependsOn is a list of workflows that MUST be completed before this workflow can be processed
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty" hcl:"dependsOn,optional"`

	// Steps is an ordered list of steps where each step represents a call to an API operation
	// or to another workflow (required, minItems: 1)
	Steps []*Step `json:"steps" yaml:"steps" hcl:"step,block"`

	// SuccessActions is a list of success actions that are applicable for all steps
	// described under this workflow
	SuccessActions []*SuccessActionOrReusable `json:"successActions,omitempty" yaml:"successActions,omitempty" hcl:"successAction,block"`

	// FailureActions is a list of failure actions that are applicable for all steps
	// described under this workflow
	FailureActions []*FailureActionOrReusable `json:"failureActions,omitempty" yaml:"failureActions,omitempty" hcl:"failureAction,block"`

	// Outputs is a map between a friendly name and a dynamic output value
	// Pattern for keys: ^[a-zA-Z0-9\.\-_]+$
	Outputs map[string]string `json:"outputs,omitempty" yaml:"outputs,omitempty" hcl:"outputs,optional"`

	// Parameters is a list of parameters that are applicable for all steps
	// described under this workflow
	Parameters []*ParameterOrReusable `json:"parameters,omitempty" yaml:"parameters,omitempty" hcl:"parameter,block"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type workflowAlias Workflow

var workflowKnownFields = []string{
	"workflowId",
	"summary",
	"description",
	"inputs",
	"dependsOn",
	"steps",
	"successActions",
	"failureActions",
	"outputs",
	"parameters",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (w *Workflow) UnmarshalJSON(data []byte) error {
	var alias workflowAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*w = Workflow(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	w.Extensions = extractExtensions(raw, workflowKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (w Workflow) MarshalJSON() ([]byte, error) {
	alias := workflowAlias(w)
	return marshalWithExtensions(&alias, w.Extensions)
}

// SuccessActionOrReusable represents either a SuccessAction or a ReusableObject.
type SuccessActionOrReusable struct {
	SuccessAction *SuccessAction  `hcl:"successAction,block"`
	Reusable      *ReusableObject `hcl:"reusable,block"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *SuccessActionOrReusable) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as ReusableObject first (check for "reference" field)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, hasReference := raw["reference"]; hasReference {
		s.Reusable = &ReusableObject{}
		return json.Unmarshal(data, s.Reusable)
	}

	s.SuccessAction = &SuccessAction{}
	return json.Unmarshal(data, s.SuccessAction)
}

// MarshalJSON implements the json.Marshaler interface.
func (s SuccessActionOrReusable) MarshalJSON() ([]byte, error) {
	if s.Reusable != nil {
		return json.Marshal(s.Reusable)
	}
	return json.Marshal(s.SuccessAction)
}

// FailureActionOrReusable represents either a FailureAction or a ReusableObject.
type FailureActionOrReusable struct {
	FailureAction *FailureAction  `hcl:"failureAction,block"`
	Reusable      *ReusableObject `hcl:"reusable,block"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *FailureActionOrReusable) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as ReusableObject first (check for "reference" field)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, hasReference := raw["reference"]; hasReference {
		f.Reusable = &ReusableObject{}
		return json.Unmarshal(data, f.Reusable)
	}

	f.FailureAction = &FailureAction{}
	return json.Unmarshal(data, f.FailureAction)
}

// MarshalJSON implements the json.Marshaler interface.
func (f FailureActionOrReusable) MarshalJSON() ([]byte, error) {
	if f.Reusable != nil {
		return json.Marshal(f.Reusable)
	}
	return json.Marshal(f.FailureAction)
}

// ParameterOrReusable represents either a Parameter or a ReusableObject.
type ParameterOrReusable struct {
	Parameter *Parameter      `hcl:"parameter,block"`
	Reusable  *ReusableObject `hcl:"reusable,block"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *ParameterOrReusable) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as ReusableObject first (check for "reference" field)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, hasReference := raw["reference"]; hasReference {
		p.Reusable = &ReusableObject{}
		return json.Unmarshal(data, p.Reusable)
	}

	p.Parameter = &Parameter{}
	return json.Unmarshal(data, p.Parameter)
}

// MarshalJSON implements the json.Marshaler interface.
func (p ParameterOrReusable) MarshalJSON() ([]byte, error) {
	if p.Reusable != nil {
		return json.Marshal(p.Reusable)
	}
	return json.Marshal(p.Parameter)
}
