package arazzo1

import (
	"encoding/json"
	"fmt"
	"testing"
)

// generateLargeDocument creates a large Arazzo document for benchmarking
func generateLargeDocument(numWorkflows, numStepsPerWorkflow int) *Arazzo {
	workflows := make([]*Workflow, numWorkflows)
	for i := 0; i < numWorkflows; i++ {
		steps := make([]*Step, numStepsPerWorkflow)
		for j := 0; j < numStepsPerWorkflow; j++ {
			steps[j] = &Step{
				StepId:      fmt.Sprintf("step-%d-%d", i, j),
				OperationId: fmt.Sprintf("operation-%d-%d", i, j),
				Description: fmt.Sprintf("This is step %d in workflow %d with a longer description to add some payload", j, i),
				Parameters: []any{
					map[string]any{
						"name":  "param1",
						"in":    "query",
						"value": "$inputs.value1",
					},
					map[string]any{
						"name":  "param2",
						"in":    "header",
						"value": fmt.Sprintf("$steps.step-%d-%d.outputs.token", i, j),
					},
				},
				RequestBody: &RequestBody{
					ContentType: "application/json",
					Payload: map[string]any{
						"field1": "$inputs.data",
						"field2": 123,
						"field3": true,
						"nested": map[string]any{
							"inner": "value",
						},
					},
				},
				SuccessCriteria: []*Criterion{
					{
						Condition: "$statusCode == 200",
						Type:      CriterionTypeSimple,
					},
					{
						Context:   "$response.body",
						Condition: "$.data.id",
						Type:      CriterionTypeJSONPath,
					},
				},
				OnSuccess: []*SuccessActionOrReusable{
					{
						SuccessAction: &SuccessAction{
							Name: "continue",
							Type: SuccessActionTypeGoto,
							StepId: func() string {
								if j < numStepsPerWorkflow-1 {
									return fmt.Sprintf("step-%d-%d", i, j+1)
								}
								return ""
							}(),
						},
					},
				},
				OnFailure: []*FailureActionOrReusable{
					{
						FailureAction: &FailureAction{
							Name:       "retry",
							Type:       FailureActionTypeRetry,
							RetryAfter: func() *float64 { v := 1.0; return &v }(),
							RetryLimit: func() *int { v := 3; return &v }(),
						},
					},
				},
				Outputs: map[string]string{
					"result": "$response.body.data",
					"token":  "$response.header.Authorization",
				},
			}
		}

		workflows[i] = &Workflow{
			WorkflowId:  fmt.Sprintf("workflow-%d", i),
			Summary:     fmt.Sprintf("Workflow %d for testing", i),
			Description: "A workflow with multiple steps for performance testing",
			DependsOn: func() []string {
				if i > 0 {
					return []string{fmt.Sprintf("workflow-%d", i-1)}
				}
				return nil
			}(),
			Steps: steps,
			Outputs: map[string]string{
				"finalResult": fmt.Sprintf("$steps.step-%d-%d.outputs.result", i, numStepsPerWorkflow-1),
			},
		}
	}

	return &Arazzo{
		Arazzo: "1.0.0",
		Info: &Info{
			Title:       "Large Benchmark Document",
			Summary:     "A large document for performance testing",
			Description: "This document contains many workflows and steps to test parsing and marshaling performance",
			Version:     "1.0.0",
		},
		SourceDescriptions: []*SourceDescription{
			{
				Name: "api",
				URL:  "https://api.example.com/openapi.json",
				Type: SourceDescriptionTypeOpenAPI,
			},
		},
		Workflows: workflows,
		Components: &Components{
			Parameters: map[string]*Parameter{
				"AuthHeader": {
					Name:  "Authorization",
					In:    ParameterInHeader,
					Value: "$inputs.token",
				},
			},
			SuccessActions: map[string]*SuccessAction{
				"EndSuccess": {
					Name: "end-success",
					Type: SuccessActionTypeEnd,
				},
			},
			FailureActions: map[string]*FailureAction{
				"RetryOnce": {
					Name:       "retry-once",
					Type:       FailureActionTypeRetry,
					RetryLimit: func() *int { v := 1; return &v }(),
				},
			},
		},
	}
}

func BenchmarkMarshalSmall(b *testing.B) {
	doc := generateLargeDocument(1, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalSmall(b *testing.B) {
	doc := generateLargeDocument(1, 2)
	data, _ := json.Marshal(doc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result Arazzo
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalMedium(b *testing.B) {
	doc := generateLargeDocument(5, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalMedium(b *testing.B) {
	doc := generateLargeDocument(5, 10)
	data, _ := json.Marshal(doc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result Arazzo
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalLarge(b *testing.B) {
	doc := generateLargeDocument(20, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalLarge(b *testing.B) {
	doc := generateLargeDocument(20, 50)
	data, _ := json.Marshal(doc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result Arazzo
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateSmall(b *testing.B) {
	doc := generateLargeDocument(1, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.Validate()
	}
}

func BenchmarkValidateMedium(b *testing.B) {
	doc := generateLargeDocument(5, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.Validate()
	}
}

func BenchmarkValidateLarge(b *testing.B) {
	doc := generateLargeDocument(20, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.Validate()
	}
}

func BenchmarkRoundTripSmall(b *testing.B) {
	doc := generateLargeDocument(1, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(doc)
		if err != nil {
			b.Fatal(err)
		}
		var result Arazzo
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTripMedium(b *testing.B) {
	doc := generateLargeDocument(5, 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(doc)
		if err != nil {
			b.Fatal(err)
		}
		var result Arazzo
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTripLarge(b *testing.B) {
	doc := generateLargeDocument(20, 50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(doc)
		if err != nil {
			b.Fatal(err)
		}
		var result Arazzo
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMarshalWithExtensions tests marshaling with many extensions
func BenchmarkMarshalWithExtensions(b *testing.B) {
	doc := generateLargeDocument(5, 10)
	// Add extensions at various levels
	doc.Extensions = map[string]any{
		"x-root-ext":  "value",
		"x-root-ext2": map[string]any{"nested": true},
	}
	doc.Info.Extensions = map[string]any{
		"x-info-ext": "info value",
	}
	for _, wf := range doc.Workflows {
		wf.Extensions = map[string]any{
			"x-workflow-ext": wf.WorkflowId,
		}
		for _, step := range wf.Steps {
			step.Extensions = map[string]any{
				"x-step-ext": step.StepId,
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnmarshalWithExtensions tests unmarshaling with many extensions
func BenchmarkUnmarshalWithExtensions(b *testing.B) {
	doc := generateLargeDocument(5, 10)
	doc.Extensions = map[string]any{
		"x-root-ext":  "value",
		"x-root-ext2": map[string]any{"nested": true},
	}
	doc.Info.Extensions = map[string]any{
		"x-info-ext": "info value",
	}
	for _, wf := range doc.Workflows {
		wf.Extensions = map[string]any{
			"x-workflow-ext": wf.WorkflowId,
		}
		for _, step := range wf.Steps {
			step.Extensions = map[string]any{
				"x-step-ext": step.StepId,
			}
		}
	}
	data, _ := json.Marshal(doc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result Arazzo
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}
