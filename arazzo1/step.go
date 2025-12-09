package arazzo1

import (
	"encoding/json"
)

// Step describes a single workflow step which MAY be a call to an API operation
// (OpenAPI Operation Object or another Workflow Object).
type Step struct {
	// StepId is a unique string to represent the step (required)
	StepId string `json:"stepId"`

	// Description of the step. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`

	// OperationId is the name of an existing, resolvable operation,
	// as defined with a unique operationId and existing within one of the sourceDescriptions.
	// Mutually exclusive with OperationPath and WorkflowId.
	OperationId string `json:"operationId,omitempty"`

	// OperationPath is a reference to a Source combined with a JSON Pointer to reference an operation.
	// Mutually exclusive with OperationId and WorkflowId.
	OperationPath string `json:"operationPath,omitempty"`

	// WorkflowId is the workflowId referencing an existing workflow within the Arazzo description.
	// Mutually exclusive with OperationId and OperationPath.
	WorkflowId string `json:"workflowId,omitempty"`

	// Parameters is a list of parameters that MUST be passed to an operation or workflow
	// as referenced by operationId, operationPath, or workflowId.
	// The schema varies based on whether the target is an operation (requires "in") or workflow.
	Parameters []any `json:"parameters,omitempty"`

	// RequestBody is the request body to pass to an operation.
	RequestBody *RequestBody `json:"requestBody,omitempty"`

	// SuccessCriteria is a list of assertions to determine the success of the step.
	SuccessCriteria []*Criterion `json:"successCriteria,omitempty"`

	// OnSuccess is an array of success action objects that specify what to do upon step success.
	OnSuccess []*SuccessActionOrReusable `json:"onSuccess,omitempty"`

	// OnFailure is an array of failure action objects that specify what to do upon step failure.
	OnFailure []*FailureActionOrReusable `json:"onFailure,omitempty"`

	// Outputs is a map between a friendly name and a dynamic output value
	// defined using a runtime expression.
	// Pattern for keys: ^[a-zA-Z0-9\.\-_]+$
	Outputs map[string]string `json:"outputs,omitempty"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-"`
}

type stepAlias Step

var stepKnownFields = []string{
	"stepId",
	"description",
	"operationId",
	"operationPath",
	"workflowId",
	"parameters",
	"requestBody",
	"successCriteria",
	"onSuccess",
	"onFailure",
	"outputs",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Step) UnmarshalJSON(data []byte) error {
	var alias stepAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*s = Step(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Extensions = extractExtensions(raw, stepKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (s Step) MarshalJSON() ([]byte, error) {
	alias := stepAlias(s)
	return marshalWithExtensions(&alias, s.Extensions)
}

// IsOperationStep returns true if this step references an operation (via operationId or operationPath).
func (s *Step) IsOperationStep() bool {
	return s.OperationId != "" || s.OperationPath != ""
}

// IsWorkflowStep returns true if this step references another workflow.
func (s *Step) IsWorkflowStep() bool {
	return s.WorkflowId != ""
}
