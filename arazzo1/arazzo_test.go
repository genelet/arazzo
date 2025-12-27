package arazzo1

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMinimalArazzoRoundTrip(t *testing.T) {
	input := `{
		"arazzo": "1.0.0",
		"info": {
			"title": "Test Workflow",
			"version": "1.0.0"
		},
		"sourceDescriptions": [
			{
				"name": "petstore",
				"url": "https://example.com/openapi.json",
				"type": "openapi"
			}
		],
		"workflows": [
			{
				"workflowId": "get-pet",
				"steps": [
					{
						"stepId": "get-pet-step",
						"operationId": "getPet"
					}
				]
			}
		]
	}`

	var arazzo Arazzo
	if err := json.Unmarshal([]byte(input), &arazzo); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify basic structure
	if arazzo.Arazzo != "1.0.0" {
		t.Errorf("Expected arazzo version 1.0.0, got %s", arazzo.Arazzo)
	}
	if arazzo.Info == nil || arazzo.Info.Title != "Test Workflow" {
		t.Error("Info not properly parsed")
	}
	if len(arazzo.SourceDescriptions) != 1 {
		t.Errorf("Expected 1 source description, got %d", len(arazzo.SourceDescriptions))
	}
	if len(arazzo.Workflows) != 1 {
		t.Errorf("Expected 1 workflow, got %d", len(arazzo.Workflows))
	}

	// Marshal back
	output, err := json.Marshal(&arazzo)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal again to verify round-trip
	var arazzo2 Arazzo
	if err := json.Unmarshal(output, &arazzo2); err != nil {
		t.Fatalf("Failed to unmarshal round-trip: %v", err)
	}

	if arazzo2.Arazzo != arazzo.Arazzo {
		t.Error("Round-trip failed: arazzo version mismatch")
	}
	if arazzo2.Info.Title != arazzo.Info.Title {
		t.Error("Round-trip failed: info title mismatch")
	}
}

func TestArazzoWithExtensions(t *testing.T) {
	input := `{
		"arazzo": "1.0.0",
		"info": {
			"title": "Test",
			"version": "1.0.0",
			"x-custom-info": "info extension"
		},
		"sourceDescriptions": [
			{
				"name": "api",
				"url": "https://example.com/api.json",
				"x-custom-source": {"key": "value"}
			}
		],
		"workflows": [
			{
				"workflowId": "test-workflow",
				"steps": [
					{
						"stepId": "step1",
						"operationId": "testOp",
						"x-custom-step": 123
					}
				],
				"x-custom-workflow": true
			}
		],
		"x-custom-root": "root extension"
	}`

	var arazzo Arazzo
	if err := json.Unmarshal([]byte(input), &arazzo); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check root extension
	if arazzo.Extensions["x-custom-root"] != "root extension" {
		t.Error("Root extension not preserved")
	}

	// Check info extension
	if arazzo.Info.Extensions["x-custom-info"] != "info extension" {
		t.Error("Info extension not preserved")
	}

	// Check source description extension
	if arazzo.SourceDescriptions[0].Extensions["x-custom-source"] == nil {
		t.Error("Source description extension not preserved")
	}

	// Check workflow extension
	if arazzo.Workflows[0].Extensions["x-custom-workflow"] != true {
		t.Error("Workflow extension not preserved")
	}

	// Check step extension
	if arazzo.Workflows[0].Steps[0].Extensions["x-custom-step"] != float64(123) {
		t.Error("Step extension not preserved")
	}

	// Marshal and verify extensions are preserved
	output, err := json.Marshal(&arazzo)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var arazzo2 Arazzo
	if err := json.Unmarshal(output, &arazzo2); err != nil {
		t.Fatalf("Failed to unmarshal round-trip: %v", err)
	}

	if arazzo2.Extensions["x-custom-root"] != "root extension" {
		t.Error("Round-trip failed: root extension not preserved")
	}
}

func TestCompleteArazzoDocument(t *testing.T) {
	input := `{
		"arazzo": "1.0.0",
		"info": {
			"title": "Pet Store Workflow",
			"summary": "Workflows for pet store operations",
			"description": "Complete workflows for managing pets",
			"version": "1.0.0"
		},
		"sourceDescriptions": [
			{
				"name": "petstore",
				"url": "./openapi.json",
				"type": "openapi"
			},
			{
				"name": "inventory",
				"url": "./inventory-arazzo.json",
				"type": "arazzo"
			}
		],
		"workflows": [
			{
				"workflowId": "create-and-get-pet",
				"summary": "Create a pet and retrieve it",
				"description": "This workflow creates a new pet and then retrieves it",
				"inputs": {
					"type": "object",
					"properties": {
						"petName": {"type": "string"},
						"petStatus": {"type": "string", "enum": ["available", "pending", "sold"]}
					},
					"required": ["petName"]
				},
				"dependsOn": ["setup-workflow"],
				"steps": [
					{
						"stepId": "create-pet",
						"operationId": "addPet",
						"requestBody": {
							"contentType": "application/json",
							"payload": {
								"name": "$inputs.petName",
								"status": "$inputs.petStatus"
							},
							"replacements": [
								{
									"target": "/id",
									"value": "$randomUUID()"
								}
							]
						},
						"successCriteria": [
							{
								"condition": "$statusCode == 201"
							},
							{
								"context": "$response.body",
								"condition": "$.id",
								"type": "jsonpath"
							}
						],
						"onSuccess": [
							{
								"name": "continue",
								"type": "goto",
								"stepId": "get-pet"
							}
						],
						"onFailure": [
							{
								"name": "retry-create",
								"type": "retry",
								"retryAfter": 1.5,
								"retryLimit": 3
							},
							{
								"name": "end-on-failure",
								"type": "end",
								"criteria": [
									{
										"condition": "$statusCode >= 500"
									}
								]
							}
						],
						"outputs": {
							"petId": "$response.body.id"
						}
					},
					{
						"stepId": "get-pet",
						"operationPath": "petstore#/paths/~1pet~1{petId}/get",
						"parameters": [
							{
								"name": "petId",
								"in": "path",
								"value": "$steps.create-pet.outputs.petId"
							}
						]
					}
				],
				"successActions": [
					{
						"name": "log-success",
						"type": "end"
					}
				],
				"failureActions": [
					{
						"name": "log-failure",
						"type": "end"
					}
				],
				"outputs": {
					"createdPetId": "$steps.create-pet.outputs.petId"
				},
				"parameters": [
					{
						"name": "api-key",
						"in": "header",
						"value": "$inputs.apiKey"
					}
				]
			},
			{
				"workflowId": "nested-workflow",
				"steps": [
					{
						"stepId": "call-workflow",
						"workflowId": "create-and-get-pet",
						"parameters": [
							{
								"name": "petName",
								"value": "Fluffy"
							}
						]
					}
				]
			}
		],
		"components": {
			"inputs": {
				"PetInput": {
					"type": "object",
					"properties": {
						"name": {"type": "string"},
						"tag": {"type": "string"}
					}
				}
			},
			"parameters": {
				"ApiKeyHeader": {
					"name": "X-API-Key",
					"in": "header",
					"value": "$inputs.apiKey"
				}
			},
			"successActions": {
				"LogSuccess": {
					"name": "log-success",
					"type": "end"
				}
			},
			"failureActions": {
				"RetryOnce": {
					"name": "retry-once",
					"type": "retry",
					"retryLimit": 1,
					"retryAfter": 2.0
				}
			}
		}
	}`

	var arazzo Arazzo
	if err := json.Unmarshal([]byte(input), &arazzo); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify structure
	if arazzo.Info.Summary != "Workflows for pet store operations" {
		t.Error("Info summary not parsed correctly")
	}

	if len(arazzo.SourceDescriptions) != 2 {
		t.Errorf("Expected 2 source descriptions, got %d", len(arazzo.SourceDescriptions))
	}

	if arazzo.SourceDescriptions[1].Type != SourceDescriptionTypeArazzo {
		t.Error("Source description type not parsed correctly")
	}

	workflow := arazzo.Workflows[0]
	if workflow.Summary != "Create a pet and retrieve it" {
		t.Error("Workflow summary not parsed correctly")
	}

	if len(workflow.DependsOn) != 1 || workflow.DependsOn[0] != "setup-workflow" {
		t.Error("Workflow dependsOn not parsed correctly")
	}

	step := workflow.Steps[0]
	if step.RequestBody == nil {
		t.Error("Request body not parsed")
	}
	if step.RequestBody.ContentType != "application/json" {
		t.Error("Request body content type not parsed correctly")
	}
	if len(step.RequestBody.Replacements) != 1 {
		t.Error("Request body replacements not parsed correctly")
	}

	if len(step.SuccessCriteria) != 2 {
		t.Errorf("Expected 2 success criteria, got %d", len(step.SuccessCriteria))
	}
	if step.SuccessCriteria[1].Type != CriterionTypeJSONPath {
		t.Error("Criterion type not parsed correctly")
	}

	if len(step.OnSuccess) != 1 {
		t.Error("onSuccess not parsed correctly")
	}
	if step.OnSuccess[0].SuccessAction.Type != SuccessActionTypeGoto {
		t.Error("Success action type not parsed correctly")
	}

	if len(step.OnFailure) != 2 {
		t.Error("onFailure not parsed correctly")
	}
	failureAction := step.OnFailure[0].FailureAction
	if failureAction.Type != FailureActionTypeRetry {
		t.Error("Failure action type not parsed correctly")
	}
	if failureAction.RetryAfter == nil || *failureAction.RetryAfter != 1.5 {
		t.Error("RetryAfter not parsed correctly")
	}
	if failureAction.RetryLimit == nil || *failureAction.RetryLimit != 3 {
		t.Error("RetryLimit not parsed correctly")
	}

	// Check second step with operationPath
	step2 := workflow.Steps[1]
	if step2.OperationPath != "petstore#/paths/~1pet~1{petId}/get" {
		t.Error("OperationPath not parsed correctly")
	}

	// Check components
	if arazzo.Components == nil {
		t.Fatal("Components not parsed")
	}
	if len(arazzo.Components.Inputs) != 1 {
		t.Error("Components inputs not parsed correctly")
	}
	if len(arazzo.Components.Parameters) != 1 {
		t.Error("Components parameters not parsed correctly")
	}
	if arazzo.Components.Parameters["ApiKeyHeader"].In != ParameterInHeader {
		t.Error("Component parameter 'in' not parsed correctly")
	}
	if len(arazzo.Components.SuccessActions) != 1 {
		t.Error("Components successActions not parsed correctly")
	}
	if len(arazzo.Components.FailureActions) != 1 {
		t.Error("Components failureActions not parsed correctly")
	}

	// Marshal back and verify round-trip
	output, err := json.Marshal(&arazzo)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var arazzo2 Arazzo
	if err := json.Unmarshal(output, &arazzo2); err != nil {
		t.Fatalf("Failed to unmarshal round-trip: %v", err)
	}

	// Re-marshal and compare
	output2, err := json.Marshal(&arazzo2)
	if err != nil {
		t.Fatalf("Failed to marshal second time: %v", err)
	}

	if string(output) != string(output2) {
		t.Error("Round-trip marshaling produced different results")
	}
}

func TestReusableObjects(t *testing.T) {
	input := `{
		"arazzo": "1.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"sourceDescriptions": [{"name": "api", "url": "./api.json"}],
		"workflows": [
			{
				"workflowId": "test",
				"steps": [
					{
						"stepId": "step1",
						"operationId": "testOp"
					}
				],
				"parameters": [
					{
						"reference": "$components.parameters.ApiKey",
						"value": "overridden-value"
					},
					{
						"name": "inline-param",
						"in": "query",
						"value": "inline-value"
					}
				],
				"successActions": [
					{
						"reference": "$components.successActions.LogSuccess"
					}
				],
				"failureActions": [
					{
						"reference": "$components.failureActions.RetryOnce",
						"value": {"retryLimit": 5}
					}
				]
			}
		]
	}`

	var arazzo Arazzo
	if err := json.Unmarshal([]byte(input), &arazzo); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	workflow := arazzo.Workflows[0]

	// Check reusable parameter
	if workflow.Parameters[0].Reusable == nil {
		t.Error("First parameter should be reusable")
	}
	if workflow.Parameters[0].Reusable.Reference != "$components.parameters.ApiKey" {
		t.Error("Reusable reference not parsed correctly")
	}
	if workflow.Parameters[0].Reusable.Value != "overridden-value" {
		t.Error("Reusable value override not parsed correctly")
	}

	// Check inline parameter
	if workflow.Parameters[1].Parameter == nil {
		t.Error("Second parameter should be inline")
	}
	if workflow.Parameters[1].Parameter.Name != "inline-param" {
		t.Error("Inline parameter name not parsed correctly")
	}

	// Check reusable success action
	if workflow.SuccessActions[0].Reusable == nil {
		t.Error("Success action should be reusable")
	}

	// Check reusable failure action with value override
	if workflow.FailureActions[0].Reusable == nil {
		t.Error("Failure action should be reusable")
	}
	if workflow.FailureActions[0].Reusable.Value == nil {
		t.Error("Failure action value override not parsed")
	}

	// Round-trip test
	output, err := json.Marshal(&arazzo)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var arazzo2 Arazzo
	if err := json.Unmarshal(output, &arazzo2); err != nil {
		t.Fatalf("Failed to unmarshal round-trip: %v", err)
	}

	if arazzo2.Workflows[0].Parameters[0].Reusable.Reference != "$components.parameters.ApiKey" {
		t.Error("Round-trip failed for reusable parameter")
	}
}

func TestCriterionWithExpressionType(t *testing.T) {
	input := `{
		"arazzo": "1.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"sourceDescriptions": [{"name": "api", "url": "./api.json"}],
		"workflows": [
			{
				"workflowId": "test",
				"steps": [
					{
						"stepId": "step1",
						"operationId": "testOp",
						"successCriteria": [
							{
								"context": "$response.body",
								"condition": "$.data[*].id",
								"type": "jsonpath",
								"version": "draft-goessner-dispatch-jsonpath-00"
							},
							{
								"context": "$response.body",
								"condition": "//element/@attr",
								"type": "xpath",
								"version": "xpath-30"
							},
							{
								"condition": "$statusCode == 200",
								"type": "simple"
							}
						]
					}
				]
			}
		]
	}`

	var arazzo Arazzo
	if err := json.Unmarshal([]byte(input), &arazzo); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	criteria := arazzo.Workflows[0].Steps[0].SuccessCriteria

	// Check JSONPath criterion with expression type
	if criteria[0].ExpressionType == nil {
		t.Error("JSONPath criterion should have expression type")
	}
	if criteria[0].ExpressionType.Type != CriterionTypeJSONPath {
		t.Error("JSONPath criterion type incorrect")
	}
	if criteria[0].ExpressionType.Version != "draft-goessner-dispatch-jsonpath-00" {
		t.Error("JSONPath criterion version incorrect")
	}

	// Check XPath criterion with expression type
	if criteria[1].ExpressionType == nil {
		t.Error("XPath criterion should have expression type")
	}
	if criteria[1].ExpressionType.Version != "xpath-30" {
		t.Error("XPath criterion version incorrect")
	}

	// Check simple criterion (no expression type)
	if criteria[2].Type != CriterionTypeSimple {
		t.Error("Simple criterion type incorrect")
	}

	// Round-trip
	output, err := json.Marshal(&arazzo)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var arazzo2 Arazzo
	if err := json.Unmarshal(output, &arazzo2); err != nil {
		t.Fatalf("Failed to unmarshal round-trip: %v", err)
	}

	if arazzo2.Workflows[0].Steps[0].SuccessCriteria[0].ExpressionType.Version != "draft-goessner-dispatch-jsonpath-00" {
		t.Error("Round-trip failed for criterion expression type")
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		arazzo      *Arazzo
		expectValid bool
		expectError string
	}{
		{
			name: "valid minimal",
			arazzo: &Arazzo{
				Arazzo: "1.0.0",
				Info:   &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{
					{Name: "api", URL: "./api.json"},
				},
				Workflows: []*Workflow{
					{
						WorkflowId: "test",
						Steps: []*Step{
							{StepId: "step1", OperationId: "op1"},
						},
					},
				},
			},
			expectValid: true,
		},
		{
			name: "missing arazzo version",
			arazzo: &Arazzo{
				Info: &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{
					{Name: "api", URL: "./api.json"},
				},
				Workflows: []*Workflow{
					{WorkflowId: "test", Steps: []*Step{{StepId: "step1", OperationId: "op1"}}},
				},
			},
			expectValid: false,
			expectError: "arazzo",
		},
		{
			name: "invalid arazzo version pattern",
			arazzo: &Arazzo{
				Arazzo: "2.0.0",
				Info:   &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{
					{Name: "api", URL: "./api.json"},
				},
				Workflows: []*Workflow{
					{WorkflowId: "test", Steps: []*Step{{StepId: "step1", OperationId: "op1"}}},
				},
			},
			expectValid: false,
			expectError: "arazzo",
		},
		{
			name: "missing info",
			arazzo: &Arazzo{
				Arazzo: "1.0.0",
				SourceDescriptions: []*SourceDescription{
					{Name: "api", URL: "./api.json"},
				},
				Workflows: []*Workflow{
					{WorkflowId: "test", Steps: []*Step{{StepId: "step1", OperationId: "op1"}}},
				},
			},
			expectValid: false,
			expectError: "info",
		},
		{
			name: "empty source descriptions",
			arazzo: &Arazzo{
				Arazzo:             "1.0.0",
				Info:               &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{},
				Workflows: []*Workflow{
					{WorkflowId: "test", Steps: []*Step{{StepId: "step1", OperationId: "op1"}}},
				},
			},
			expectValid: false,
			expectError: "sourceDescriptions",
		},
		{
			name: "empty workflows",
			arazzo: &Arazzo{
				Arazzo: "1.0.0",
				Info:   &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{
					{Name: "api", URL: "./api.json"},
				},
				Workflows: []*Workflow{},
			},
			expectValid: false,
			expectError: "workflows",
		},
		{
			name: "step missing operation reference",
			arazzo: &Arazzo{
				Arazzo: "1.0.0",
				Info:   &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{
					{Name: "api", URL: "./api.json"},
				},
				Workflows: []*Workflow{
					{WorkflowId: "test", Steps: []*Step{{StepId: "step1"}}},
				},
			},
			expectValid: false,
			expectError: "operationId, operationPath, or workflowId",
		},
		{
			name: "step with multiple operation references",
			arazzo: &Arazzo{
				Arazzo: "1.0.0",
				Info:   &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{
					{Name: "api", URL: "./api.json"},
				},
				Workflows: []*Workflow{
					{
						WorkflowId: "test",
						Steps:      []*Step{{StepId: "step1", OperationId: "op1", WorkflowId: "wf1"}},
					},
				},
			},
			expectValid: false,
			expectError: "only one",
		},
		{
			name: "invalid source name pattern",
			arazzo: &Arazzo{
				Arazzo: "1.0.0",
				Info:   &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{
					{Name: "invalid name with spaces", URL: "./api.json"},
				},
				Workflows: []*Workflow{
					{WorkflowId: "test", Steps: []*Step{{StepId: "step1", OperationId: "op1"}}},
				},
			},
			expectValid: false,
			expectError: "name",
		},
		{
			name: "goto action without target",
			arazzo: &Arazzo{
				Arazzo: "1.0.0",
				Info:   &Info{Title: "Test", Version: "1.0.0"},
				SourceDescriptions: []*SourceDescription{
					{Name: "api", URL: "./api.json"},
				},
				Workflows: []*Workflow{
					{
						WorkflowId: "test",
						Steps: []*Step{
							{
								StepId:      "step1",
								OperationId: "op1",
								OnSuccess: []*SuccessActionOrReusable{
									{SuccessAction: &SuccessAction{Name: "goto-nowhere", Type: SuccessActionTypeGoto}},
								},
							},
						},
					},
				},
			},
			expectValid: false,
			expectError: "workflowId or stepId",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.arazzo.Validate()
			if tc.expectValid && !result.Valid() {
				t.Errorf("Expected valid, got errors: %s", result.Error())
			}
			if !tc.expectValid && result.Valid() {
				t.Error("Expected invalid, but validation passed")
			}
			if !tc.expectValid && tc.expectError != "" {
				errorStr := result.Error()
				if errorStr == "" || !strings.Contains(errorStr, tc.expectError) {
					t.Errorf("Expected error containing '%s', got: %s", tc.expectError, errorStr)
				}
			}
		})
	}
}

func TestStepHelperMethods(t *testing.T) {
	step1 := &Step{StepId: "s1", OperationId: "op1"}
	if !step1.IsOperationStep() {
		t.Error("Step with operationId should be operation step")
	}
	if step1.IsWorkflowStep() {
		t.Error("Step with operationId should not be workflow step")
	}

	step2 := &Step{StepId: "s2", OperationPath: "api#/paths/~1pets/get"}
	if !step2.IsOperationStep() {
		t.Error("Step with operationPath should be operation step")
	}

	step3 := &Step{StepId: "s3", WorkflowId: "workflow1"}
	if step3.IsOperationStep() {
		t.Error("Step with workflowId should not be operation step")
	}
	if !step3.IsWorkflowStep() {
		t.Error("Step with workflowId should be workflow step")
	}
}
