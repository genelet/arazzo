package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewGeneratorFromArazzo(t *testing.T) {
	// 1. Create temporary OpenAPI file
	openapiYAML := `
openapi: 3.0.0
info:
  title: Reverse API
  version: 1.0.0
servers:
  - url: http://api.reverse.com
paths:
  /items/{id}:
    get:
      operationId: getItem
      responses:
        '200':
          description: OK
`
	tmpDir := t.TempDir()
	openapiFile := filepath.Join(tmpDir, "openapi_rev.yaml")
	if err := os.WriteFile(openapiFile, []byte(openapiYAML), 0644); err != nil {
		t.Fatalf("failed to write openapi file: %v", err)
	}

	// 2. Create temporary Arazzo file
	arazzoJSON := `
{
  "arazzo": "1.0.0",
  "info": {
    "title": "Reverse Test Arazzo",
    "version": "1.0.0"
  },
  "sourceDescriptions": [
    {
      "name": "Reverse Test Arazzo",
      "url": "openapi_rev.yaml",
      "type": "openapi"
    }
  ],
  "workflows": [
    {
      "workflowId": "main",
      "steps": [
        {
          "stepId": "getMyItem",
          "operationId": "$source.getItem"
        }
      ]
    }
  ]
}
`
	arazzoFile := filepath.Join(tmpDir, "arazzo_rev.json")
	if err := os.WriteFile(arazzoFile, []byte(arazzoJSON), 0644); err != nil {
		t.Fatalf("failed to write arazzo file: %v", err)
	}

	// 3. Run NewGeneratorFromArazzo
	gen, err := NewGeneratorFromArazzo(arazzoFile, openapiFile)
	if err != nil {
		t.Fatalf("NewGeneratorFromArazzo failed: %v", err)
	}

	// 4. Verify results
	if gen.Provider.Name != "Reverse Test Arazzo" {
		t.Errorf("expected provider name 'Reverse Test Arazzo', got '%s'", gen.Provider.Name)
	}
	if gen.Provider.ServerURL != "http://api.reverse.com" {
		t.Errorf("expected server URL 'http://api.reverse.com', got '%s'", gen.Provider.ServerURL)
	}

	if len(gen.Workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(gen.Workflows))
	}
	wf := gen.Workflows[0]

	if len(wf.Steps) != 1 {
		t.Fatalf("expected 1 HTTP operation, got %d", len(wf.Steps))
	}

	op := wf.Steps[0]
	if op.Name != "getMyItem" {
		t.Errorf("expected operation name 'getMyItem', got '%s'", op.Name)
	}
}
