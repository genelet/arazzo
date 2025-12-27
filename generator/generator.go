package generator

import (
	"github.com/genelet/arazzo/arazzo1"
	"github.com/genelet/oas/openapi31"
)

// Generator represents a generator config.
type Generator struct {
	Provider   *Provider           `yaml:"provider" json:"provider" hcl:"provider,block"`
	Workflows  []*WorkflowSpec     `yaml:"workflows" json:"workflows" hcl:"workflow,block"`
	Components *arazzo1.Components `yaml:"components,omitempty" json:"components,omitempty" hcl:"components,block"`
	Extensions map[string]any      `yaml:"extensions,omitempty" json:"extensions,omitempty" hcl:"extensions,optional"`

	// Internal
	openapiDoc *openapi31.OpenAPI
}

// Provider represents the provider configuration.
type Provider struct {
	Name       string                 `yaml:"name" json:"name" hcl:"name"`
	ServerURL  string                 `yaml:"server_url" json:"server_url" hcl:"server_url"`
	Appendices map[string]interface{} `yaml:"appendices" json:"appendices" hcl:"appendices,optional"` // Reserves Info details
	Extensions map[string]any         `yaml:"extensions,omitempty" json:"extensions,omitempty" hcl:"extensions,optional"`
}

// WorkflowSpec defines a workflow in the generator.
type WorkflowSpec struct {
	WorkflowId     string                             `yaml:"workflow_id" json:"workflowId" hcl:"workflow_id,label"`
	Summary        string                             `yaml:"summary" json:"summary" hcl:"summary,optional"`
	Description    string                             `yaml:"description" json:"description" hcl:"description,optional"`
	Inputs         interface{}                        `yaml:"inputs" json:"inputs" hcl:"inputs,block"`
	Outputs        map[string]string                  `yaml:"outputs" json:"outputs" hcl:"outputs,block"`
	DependsOn      []string                           `yaml:"depends_on" json:"dependsOn" hcl:"depends_on,optional"`
	Parameters     []*arazzo1.ParameterOrReusable     `yaml:"parameters" json:"parameters" hcl:"parameter,block"`
	SuccessActions []*arazzo1.SuccessActionOrReusable `yaml:"success_actions" json:"successActions" hcl:"success_action,block"`
	FailureActions []*arazzo1.FailureActionOrReusable `yaml:"failure_actions" json:"failureActions" hcl:"failure_action,block"`
	Steps          []*OperationSpec                   `yaml:"steps" json:"steps" hcl:"step,block"`
	Extensions     map[string]any                     `yaml:"extensions,omitempty" json:"extensions,omitempty" hcl:"extensions,optional"`
}

// OperationSpec defines an operation to be included in the Arazzo workflow.
type OperationSpec struct {
	Name string `yaml:"name" json:"name" hcl:"name,label"` // Acts as label in HCL/YAML list item

	// High Fidelity Fields
	Description     string                             `yaml:"description" json:"description" hcl:"description,optional"`
	Parameters      []interface{}                      `yaml:"parameters" json:"parameters" hcl:"parameter,block"`
	RequestBody     *arazzo1.RequestBody               `yaml:"request_body" json:"requestBody" hcl:"request_body,block"`
	SuccessCriteria []*arazzo1.Criterion               `yaml:"success_criteria" json:"successCriteria" hcl:"success_criterion,block"`
	OnSuccess       []*arazzo1.SuccessActionOrReusable `yaml:"on_success" json:"onSuccess" hcl:"on_success,block"`
	OnFailure       []*arazzo1.FailureActionOrReusable `yaml:"on_failure" json:"onFailure" hcl:"on_failure,block"`
	Outputs         map[string]string                  `yaml:"outputs" json:"outputs" hcl:"outputs,block"`
	OperationPath   string                             `yaml:"operation_path" json:"operationPath" hcl:"operation_path,optional"`
	OperationId     string                             `yaml:"operation_id" json:"operationId" hcl:"operation_id,optional"`
	WorkflowId      string                             `yaml:"workflow_id" json:"workflowId" hcl:"workflow_id,optional"`
	Extensions      map[string]any                     `yaml:"extensions,omitempty" json:"extensions,omitempty" hcl:"extensions,optional"`
}
