package arazzo1

import (
	"encoding/json"
)

// SuccessActionType represents the type of success action to take.
type SuccessActionType string

const (
	// SuccessActionTypeEnd indicates the workflow should end.
	SuccessActionTypeEnd SuccessActionType = "end"

	// SuccessActionTypeGoto indicates the workflow should jump to another step or workflow.
	SuccessActionTypeGoto SuccessActionType = "goto"
)

// SuccessAction describes a single success action which describes an action
// to take upon success of a workflow step.
type SuccessAction struct {
	// Name is the name of the success action (required).
	Name string `json:"name" yaml:"name" hcl:"name,label"`

	// Type is the type of action to take (required).
	// Must be "end" or "goto".
	Type SuccessActionType `json:"type" yaml:"type" hcl:"type"`

	// WorkflowId is the workflowId referencing an existing workflow within the Arazzo description
	// to transfer to upon success of the step.
	// Required when Type is "goto" (mutually exclusive with StepId).
	WorkflowId string `json:"workflowId,omitempty" yaml:"workflowId,omitempty" hcl:"workflowId,optional"`

	// StepId is the stepId to transfer to upon success of the step.
	// Required when Type is "goto" (mutually exclusive with WorkflowId).
	StepId string `json:"stepId,omitempty" yaml:"stepId,omitempty" hcl:"stepId,optional"`

	// Criteria is a list of assertions to determine if this action SHALL be executed.
	Criteria []*Criterion `json:"criteria,omitempty" yaml:"criteria,omitempty" hcl:"criterion,block"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type successActionAlias SuccessAction

var successActionKnownFields = []string{
	"name",
	"type",
	"workflowId",
	"stepId",
	"criteria",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *SuccessAction) UnmarshalJSON(data []byte) error {
	var alias successActionAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*s = SuccessAction(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Extensions = extractExtensions(raw, successActionKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (s SuccessAction) MarshalJSON() ([]byte, error) {
	alias := successActionAlias(s)
	return marshalWithExtensions(&alias, s.Extensions)
}

// FailureActionType represents the type of failure action to take.
type FailureActionType string

const (
	// FailureActionTypeEnd indicates the workflow should end.
	FailureActionTypeEnd FailureActionType = "end"

	// FailureActionTypeGoto indicates the workflow should jump to another step or workflow.
	FailureActionTypeGoto FailureActionType = "goto"

	// FailureActionTypeRetry indicates the step should be retried.
	FailureActionTypeRetry FailureActionType = "retry"
)

// FailureAction describes a single failure action which describes an action
// to take upon failure of a workflow step.
type FailureAction struct {
	// Name is the name of the failure action (required).
	Name string `json:"name" yaml:"name" hcl:"name,label"`

	// Type is the type of action to take (required).
	// Must be "end", "goto", or "retry".
	Type FailureActionType `json:"type" yaml:"type" hcl:"type"`

	// WorkflowId is the workflowId referencing an existing workflow within the Arazzo description
	// to transfer to upon failure of the step.
	// Required when Type is "goto" (mutually exclusive with StepId).
	WorkflowId string `json:"workflowId,omitempty" yaml:"workflowId,omitempty" hcl:"workflowId,optional"`

	// StepId is the stepId to transfer to upon failure of the step.
	// Required when Type is "goto" (mutually exclusive with WorkflowId).
	StepId string `json:"stepId,omitempty" yaml:"stepId,omitempty" hcl:"stepId,optional"`

	// RetryAfter is a non-negative decimal indicating the seconds to delay after the step failure
	// before another attempt SHALL be made.
	RetryAfter *float64 `json:"retryAfter,omitempty" yaml:"retryAfter,omitempty" hcl:"retryAfter,optional"`

	// RetryLimit is a non-negative integer indicating how many attempts to retry the step
	// MAY be attempted before failing the overall step.
	RetryLimit *int `json:"retryLimit,omitempty" yaml:"retryLimit,omitempty" hcl:"retryLimit,optional"`

	// Criteria is a list of assertions to determine if this action SHALL be executed.
	Criteria []*Criterion `json:"criteria,omitempty" yaml:"criteria,omitempty" hcl:"criterion,block"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type failureActionAlias FailureAction

var failureActionKnownFields = []string{
	"name",
	"type",
	"workflowId",
	"stepId",
	"retryAfter",
	"retryLimit",
	"criteria",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *FailureAction) UnmarshalJSON(data []byte) error {
	var alias failureActionAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*f = FailureAction(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Extensions = extractExtensions(raw, failureActionKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (f FailureAction) MarshalJSON() ([]byte, error) {
	alias := failureActionAlias(f)
	return marshalWithExtensions(&alias, f.Extensions)
}
