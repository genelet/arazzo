package arazzo1

import (
	"testing"
)

func TestUnmarshalHCLBasicWorkflow(t *testing.T) {
	hclData := `
summary = "Test workflow"
description = "A test workflow for HCL unmarshaling"
dependsOn = ["other-workflow"]
outputs = {
  result = "$steps.step1.outputs.data"
}

step "step1" {
  operationId = "getUser"
  description = "Get a user"
  outputs = {
    data = "$response.body"
  }
}

step "step2" {
  operationPath = "api.get./items"
  description = "List items"
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	// Verify workflowId from label
	if w.WorkflowId != "test-workflow" {
		t.Errorf("Expected workflowId 'test-workflow', got %q", w.WorkflowId)
	}

	// Verify summary
	if w.Summary != "Test workflow" {
		t.Errorf("Expected summary 'Test workflow', got %q", w.Summary)
	}

	// Verify description
	if w.Description != "A test workflow for HCL unmarshaling" {
		t.Errorf("Expected description, got %q", w.Description)
	}

	// Verify dependsOn
	if len(w.DependsOn) != 1 || w.DependsOn[0] != "other-workflow" {
		t.Errorf("Expected dependsOn ['other-workflow'], got %v", w.DependsOn)
	}

	// Verify outputs
	if w.Outputs["result"] != "$steps.step1.outputs.data" {
		t.Errorf("Expected output result, got %v", w.Outputs)
	}

	// Verify steps
	if len(w.Steps) != 2 {
		t.Fatalf("Expected 2 steps, got %d", len(w.Steps))
	}

	// Verify first step
	if w.Steps[0].StepId != "step1" {
		t.Errorf("Expected stepId 'step1', got %q", w.Steps[0].StepId)
	}
	if w.Steps[0].OperationId != "getUser" {
		t.Errorf("Expected operationId 'getUser', got %q", w.Steps[0].OperationId)
	}

	// Verify second step
	if w.Steps[1].StepId != "step2" {
		t.Errorf("Expected stepId 'step2', got %q", w.Steps[1].StepId)
	}
	if w.Steps[1].OperationPath != "api.get./items" {
		t.Errorf("Expected operationPath 'api.get./items', got %q", w.Steps[1].OperationPath)
	}
}

func TestUnmarshalHCLWithRequestBody(t *testing.T) {
	hclData := `
step "createUser" {
  operationId = "createUser"

  requestBody {
    contentType = "application/json"
    payload = {
      name = "test"
      active = true
    }
  }
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	step := w.Steps[0]
	if step.RequestBody == nil {
		t.Fatal("Expected requestBody")
	}

	if step.RequestBody.ContentType != "application/json" {
		t.Errorf("Expected contentType 'application/json', got %q", step.RequestBody.ContentType)
	}

	payload, ok := step.RequestBody.Payload.(map[string]any)
	if !ok {
		t.Fatalf("Expected payload to be map[string]any, got %T", step.RequestBody.Payload)
	}

	if payload["name"] != "test" {
		t.Errorf("Expected payload.name 'test', got %v", payload["name"])
	}

	if payload["active"] != true {
		t.Errorf("Expected payload.active true, got %v", payload["active"])
	}
}

func TestUnmarshalHCLWithRequestBodyReplacements(t *testing.T) {
	hclData := `
step "createUser" {
  operationId = "createUser"

  requestBody {
    contentType = "application/json"
    payload = {
      name = "test"
    }

    replacement {
      target = "/name"
      value = "replace"
    }

    replacement {
      target = "/active"
      value = "true"
    }
  }
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	step := w.Steps[0]
	if step.RequestBody == nil {
		t.Fatal("Expected requestBody")
	}

	if len(step.RequestBody.Replacements) != 2 {
		t.Fatalf("Expected 2 replacements, got %d", len(step.RequestBody.Replacements))
	}

	if step.RequestBody.Replacements[0].Target != "/name" {
		t.Errorf("Expected first target '/name', got %q", step.RequestBody.Replacements[0].Target)
	}
	if step.RequestBody.Replacements[0].Value != "replace" {
		t.Errorf("Expected first value 'replace', got %q", step.RequestBody.Replacements[0].Value)
	}
	if step.RequestBody.Replacements[1].Target != "/active" {
		t.Errorf("Expected second target '/active', got %q", step.RequestBody.Replacements[1].Target)
	}
	if step.RequestBody.Replacements[1].Value != "true" {
		t.Errorf("Expected second value 'true', got %q", step.RequestBody.Replacements[1].Value)
	}
}

func TestUnmarshalHCLWithSuccessCriteria(t *testing.T) {
	hclData := `
step "checkStatus" {
  operationId = "getStatus"

  successCriterion {
    condition = "$statusCode == 200"
    type = "simple"
  }

  successCriterion {
    context = "$response.body"
    condition = "$.status"
    type = "jsonpath"
  }
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	step := w.Steps[0]
	if len(step.SuccessCriteria) != 2 {
		t.Fatalf("Expected 2 success criteria, got %d", len(step.SuccessCriteria))
	}

	// First criterion
	if step.SuccessCriteria[0].Condition != "$statusCode == 200" {
		t.Errorf("Expected condition '$statusCode == 200', got %q", step.SuccessCriteria[0].Condition)
	}
	if step.SuccessCriteria[0].Type != CriterionTypeSimple {
		t.Errorf("Expected type 'simple', got %q", step.SuccessCriteria[0].Type)
	}

	// Second criterion
	if step.SuccessCriteria[1].Context != "$response.body" {
		t.Errorf("Expected context '$response.body', got %q", step.SuccessCriteria[1].Context)
	}
	if step.SuccessCriteria[1].Type != CriterionTypeJSONPath {
		t.Errorf("Expected type 'jsonpath', got %q", step.SuccessCriteria[1].Type)
	}
}

func TestUnmarshalHCLWithCriterionExpressionType(t *testing.T) {
	hclData := `
step "checkStatus" {
  operationId = "getStatus"

  successCriterion {
    condition = "$.status"
    expressionType {
      type = "jsonpath"
      version = "draft-goessner-dispatch-jsonpath-00"
    }
  }

  successCriterion {
    condition = "$.status"
    type = "xpath"
    version = "xpath-20"
  }
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	step := w.Steps[0]
	if len(step.SuccessCriteria) != 2 {
		t.Fatalf("Expected 2 success criteria, got %d", len(step.SuccessCriteria))
	}

	first := step.SuccessCriteria[0]
	if first.ExpressionType == nil {
		t.Fatal("Expected expression type on first criterion")
	}
	if first.ExpressionType.Type != CriterionTypeJSONPath {
		t.Errorf("Expected first expression type 'jsonpath', got %q", first.ExpressionType.Type)
	}
	if first.ExpressionType.Version != "draft-goessner-dispatch-jsonpath-00" {
		t.Errorf("Expected first expression version 'draft-goessner-dispatch-jsonpath-00', got %q", first.ExpressionType.Version)
	}

	second := step.SuccessCriteria[1]
	if second.ExpressionType == nil {
		t.Fatal("Expected expression type on second criterion")
	}
	if second.ExpressionType.Type != CriterionTypeXPath {
		t.Errorf("Expected second expression type 'xpath', got %q", second.ExpressionType.Type)
	}
	if second.ExpressionType.Version != "xpath-20" {
		t.Errorf("Expected second expression version 'xpath-20', got %q", second.ExpressionType.Version)
	}
}

func TestUnmarshalHCLWithActions(t *testing.T) {
	hclData := `
step "process" {
  operationId = "processItem"

  onSuccess "continue" {
    type = "goto"
    stepId = "nextStep"
  }

  onFailure "retry" {
    type = "retry"
    retryAfter = 1.5
    retryLimit = 3
  }

  onFailure "giveUp" {
    type = "end"
  }
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	step := w.Steps[0]

	// Check onSuccess
	if len(step.OnSuccess) != 1 {
		t.Fatalf("Expected 1 onSuccess action, got %d", len(step.OnSuccess))
	}
	successAction := step.OnSuccess[0].SuccessAction
	if successAction.Name != "continue" {
		t.Errorf("Expected name 'continue', got %q", successAction.Name)
	}
	if successAction.Type != SuccessActionTypeGoto {
		t.Errorf("Expected type 'goto', got %q", successAction.Type)
	}
	if successAction.StepId != "nextStep" {
		t.Errorf("Expected stepId 'nextStep', got %q", successAction.StepId)
	}

	// Check onFailure
	if len(step.OnFailure) != 2 {
		t.Fatalf("Expected 2 onFailure actions, got %d", len(step.OnFailure))
	}

	retryAction := step.OnFailure[0].FailureAction
	if retryAction.Name != "retry" {
		t.Errorf("Expected name 'retry', got %q", retryAction.Name)
	}
	if retryAction.Type != FailureActionTypeRetry {
		t.Errorf("Expected type 'retry', got %q", retryAction.Type)
	}
	if retryAction.RetryAfter == nil || *retryAction.RetryAfter != 1.5 {
		t.Errorf("Expected retryAfter 1.5, got %v", retryAction.RetryAfter)
	}
	if retryAction.RetryLimit == nil || *retryAction.RetryLimit != 3 {
		t.Errorf("Expected retryLimit 3, got %v", retryAction.RetryLimit)
	}

	endAction := step.OnFailure[1].FailureAction
	if endAction.Type != FailureActionTypeEnd {
		t.Errorf("Expected type 'end', got %q", endAction.Type)
	}
}

func TestUnmarshalHCLWithActionCriteria(t *testing.T) {
	hclData := `
step "process" {
  operationId = "processItem"

  onSuccess "continue" {
    type = "goto"
    stepId = "nextStep"

    criterion {
      condition = "$statusCode == 200"
      type = "simple"
    }
  }

  onFailure "retry" {
    type = "retry"
    retryLimit = 3

    criterion {
      condition = "$statusCode == 500"
      type = "simple"
    }
  }
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	step := w.Steps[0]
	if len(step.OnSuccess) != 1 {
		t.Fatalf("Expected 1 onSuccess action, got %d", len(step.OnSuccess))
	}
	if len(step.OnSuccess[0].SuccessAction.Criteria) != 1 {
		t.Fatalf("Expected 1 onSuccess criterion, got %d", len(step.OnSuccess[0].SuccessAction.Criteria))
	}
	if step.OnSuccess[0].SuccessAction.Criteria[0].Condition != "$statusCode == 200" {
		t.Errorf("Expected onSuccess condition '$statusCode == 200', got %q", step.OnSuccess[0].SuccessAction.Criteria[0].Condition)
	}

	if len(step.OnFailure) != 1 {
		t.Fatalf("Expected 1 onFailure action, got %d", len(step.OnFailure))
	}
	if len(step.OnFailure[0].FailureAction.Criteria) != 1 {
		t.Fatalf("Expected 1 onFailure criterion, got %d", len(step.OnFailure[0].FailureAction.Criteria))
	}
	if step.OnFailure[0].FailureAction.Criteria[0].Condition != "$statusCode == 500" {
		t.Errorf("Expected onFailure condition '$statusCode == 500', got %q", step.OnFailure[0].FailureAction.Criteria[0].Condition)
	}
}

func TestUnmarshalHCLWithReusableActions(t *testing.T) {
	hclData := `
step "process" {
  operationId = "processItem"

  onSuccess {
    reusable {
      reference = "$components.successActions.LogSuccess"
    }
  }

  onFailure {
    reusable {
      reference = "$components.failureActions.RetryOnce"
      value {
        retryLimit = 5
      }
    }
  }
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	step := w.Steps[0]
	if len(step.OnSuccess) != 1 {
		t.Fatalf("Expected 1 onSuccess action, got %d", len(step.OnSuccess))
	}
	if step.OnSuccess[0].Reusable == nil {
		t.Fatal("Expected reusable onSuccess action")
	}
	if step.OnSuccess[0].Reusable.Reference != "$components.successActions.LogSuccess" {
		t.Errorf("Expected onSuccess reference '$components.successActions.LogSuccess', got %q", step.OnSuccess[0].Reusable.Reference)
	}

	if len(step.OnFailure) != 1 {
		t.Fatalf("Expected 1 onFailure action, got %d", len(step.OnFailure))
	}
	if step.OnFailure[0].Reusable == nil {
		t.Fatal("Expected reusable onFailure action")
	}
	if step.OnFailure[0].Reusable.Reference != "$components.failureActions.RetryOnce" {
		t.Errorf("Expected onFailure reference '$components.failureActions.RetryOnce', got %q", step.OnFailure[0].Reusable.Reference)
	}
	if step.OnFailure[0].Reusable.Value == nil {
		t.Fatal("Expected onFailure reusable value override")
	}
}

func TestUnmarshalHCLWithInputs(t *testing.T) {
	hclData := `
inputs {
  type = "object"
  properties {
    username {
      type = "string"
    }
    limit {
      type = "integer"
    }
  }
  required = ["username"]
}

step "listUsers" {
  operationId = "listUsers"
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if w.Inputs == nil {
		t.Fatal("Expected inputs")
	}

	inputs, ok := w.Inputs.(map[string]any)
	if !ok {
		t.Fatalf("Expected inputs to be map[string]any, got %T", w.Inputs)
	}

	if inputs["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", inputs["type"])
	}
}

func TestUnmarshalHCLWithWorkflowActions(t *testing.T) {
	hclData := `
successAction "logSuccess" {
  type = "end"
}

failureAction "logFailure" {
  type = "end"
}

parameter "apiKey" {
  in = "header"
  value = "$inputs.apiKey"
}

step "main" {
  operationId = "mainOp"
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	// Check workflow-level success actions
	if len(w.SuccessActions) != 1 {
		t.Fatalf("Expected 1 workflow success action, got %d", len(w.SuccessActions))
	}
	if w.SuccessActions[0].SuccessAction.Name != "logSuccess" {
		t.Errorf("Expected name 'logSuccess', got %q", w.SuccessActions[0].SuccessAction.Name)
	}

	// Check workflow-level failure actions
	if len(w.FailureActions) != 1 {
		t.Fatalf("Expected 1 workflow failure action, got %d", len(w.FailureActions))
	}
	if w.FailureActions[0].FailureAction.Name != "logFailure" {
		t.Errorf("Expected name 'logFailure', got %q", w.FailureActions[0].FailureAction.Name)
	}

	// Check workflow-level parameters
	if len(w.Parameters) != 1 {
		t.Fatalf("Expected 1 workflow parameter, got %d", len(w.Parameters))
	}
	if w.Parameters[0].Parameter.Name != "apiKey" {
		t.Errorf("Expected name 'apiKey', got %q", w.Parameters[0].Parameter.Name)
	}
	if w.Parameters[0].Parameter.In != ParameterInHeader {
		t.Errorf("Expected in 'header', got %q", w.Parameters[0].Parameter.In)
	}
}

func TestUnmarshalHCLInvalidSyntax(t *testing.T) {
	hclData := `
step "invalid {
  missing closing quote
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err == nil {
		t.Error("Expected error for invalid HCL syntax")
	}
}

func TestUnmarshalHCLEmptyWorkflow(t *testing.T) {
	hclData := ``

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "empty-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed for empty workflow: %v", err)
	}

	if w.WorkflowId != "empty-workflow" {
		t.Errorf("Expected workflowId 'empty-workflow', got %q", w.WorkflowId)
	}
}

func TestUnmarshalHCLWithStepWorkflowId(t *testing.T) {
	hclData := `
step "callOther" {
  workflowId = "other-workflow"
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	if w.Steps[0].WorkflowId != "other-workflow" {
		t.Errorf("Expected workflowId 'other-workflow', got %q", w.Steps[0].WorkflowId)
	}
}

// Test ctyToGo conversion functions
func TestCtyConversions(t *testing.T) {
	// Test through the hclBlockToMap function behavior
	// These are indirect tests since ctyToGo is internal

	hclData := `
step "test" {
  operationId = "testOp"
  outputs = {
    stringVal = "hello"
    intVal = "42"
  }
}
`

	w := &Workflow{}
	err := w.UnmarshalHCL([]byte(hclData), "test-workflow")
	if err != nil {
		t.Fatalf("UnmarshalHCL failed: %v", err)
	}

	if len(w.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(w.Steps))
	}

	// String values in outputs map
	if w.Steps[0].Outputs["stringVal"] != "hello" {
		t.Errorf("Expected stringVal 'hello', got %v", w.Steps[0].Outputs["stringVal"])
	}
}
