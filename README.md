# Arazzo Go Parser

A Go library for parsing, validating, and generating [Arazzo 1.0](https://spec.openapis.org/arazzo/v1.0.0) documents. Arazzo is an OpenAPI Initiative specification for describing workflows that span multiple APIs.

[![GoDoc](https://godoc.org/github.com/genelet/arazzo?status.svg)](https://godoc.org/github.com/genelet/arazzo)

## Installation

```bash
go get github.com/genelet/arazzo
```

## Features

- Full support for Arazzo 1.0.x specification
- Marshal/Unmarshal JSON with proper round-trip preservation
- **HCL format support** - Convert between JSON and HCL representations
- Specification extensions (`x-*`) support on all objects
- Comprehensive validation with detailed error paths
- Type-safe constants for enum values

## Quick Start

### Parsing an Arazzo Document

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/genelet/arazzo/arazzo1"
)

func main() {
    // Read the Arazzo document
    data, err := os.ReadFile("workflow.arazzo.json")
    if err != nil {
        log.Fatal(err)
    }

    // Parse it
    var doc arazzo1.Arazzo
    if err := json.Unmarshal(data, &doc); err != nil {
        log.Fatal(err)
    }

    // Access the parsed data
    fmt.Printf("Title: %s\n", doc.Info.Title)
    fmt.Printf("Version: %s\n", doc.Info.Version)
    fmt.Printf("Workflows: %d\n", len(doc.Workflows))

    for _, wf := range doc.Workflows {
        fmt.Printf("  - %s: %d steps\n", wf.WorkflowId, len(wf.Steps))
    }
}
```

### Validating a Document

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/genelet/arazzo/arazzo1"
)

func main() {
    data := []byte(`{
        "arazzo": "1.0.0",
        "info": {"title": "My Workflow", "version": "1.0.0"},
        "sourceDescriptions": [
            {"name": "petstore", "url": "./openapi.json", "type": "openapi"}
        ],
        "workflows": [
            {
                "workflowId": "get-pet",
                "steps": [
                    {"stepId": "fetch", "operationId": "getPetById"}
                ]
            }
        ]
    }`)

    var doc arazzo1.Arazzo
    if err := json.Unmarshal(data, &doc); err != nil {
        log.Fatal(err)
    }

    // Validate the document
    result := doc.Validate()
    if !result.Valid() {
        fmt.Println("Validation errors:")
        for _, err := range result.Errors {
            fmt.Printf("  %s: %s\n", err.Path, err.Message)
        }
        return
    }

    fmt.Println("Document is valid!")
}
```

### Creating a Document Programmatically

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/genelet/arazzo/arazzo1"
)

func main() {
    // Build the document
    doc := &arazzo1.Arazzo{
        Arazzo: "1.0.0",
        Info: &arazzo1.Info{
            Title:       "Pet Store Workflow",
            Summary:     "Workflows for managing pets",
            Description: "A collection of workflows for the Pet Store API",
            Version:     "1.0.0",
        },
        SourceDescriptions: []*arazzo1.SourceDescription{
            {
                Name: "petstore",
                URL:  "https://petstore3.swagger.io/api/v3/openapi.json",
                Type: arazzo1.SourceDescriptionTypeOpenAPI,
            },
        },
        Workflows: []*arazzo1.Workflow{
            {
                WorkflowId:  "create-and-get-pet",
                Summary:     "Create a pet and retrieve it",
                Description: "This workflow creates a new pet and then fetches it by ID",
                Steps: []*arazzo1.Step{
                    {
                        StepId:      "create-pet",
                        OperationId: "addPet",
                        RequestBody: &arazzo1.RequestBody{
                            ContentType: "application/json",
                            Payload: map[string]any{
                                "name":   "$inputs.petName",
                                "status": "available",
                            },
                        },
                        SuccessCriteria: []*arazzo1.Criterion{
                            {Condition: "$statusCode == 200"},
                        },
                        Outputs: map[string]string{
                            "petId": "$response.body.id",
                        },
                    },
                    {
                        StepId:      "get-pet",
                        OperationId: "getPetById",
                        Parameters: []any{
                            &arazzo1.Parameter{
                                Name:  "petId",
                                In:    arazzo1.ParameterInPath,
                                Value: "$steps.create-pet.outputs.petId",
                            },
                        },
                    },
                },
                Outputs: map[string]string{
                    "pet": "$steps.get-pet.outputs.response",
                },
            },
        },
    }

    // Validate before serializing
    if result := doc.Validate(); !result.Valid() {
        log.Fatalf("Invalid document: %s", result.Error())
    }

    // Marshal to JSON
    output, err := json.MarshalIndent(doc, "", "  ")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(output))
}
```

### Working with Extensions

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/genelet/arazzo/arazzo1"
)

func main() {
    // Parse a document with extensions
    data := []byte(`{
        "arazzo": "1.0.0",
        "info": {
            "title": "My API",
            "version": "1.0.0",
            "x-logo": "https://example.com/logo.png"
        },
        "sourceDescriptions": [{"name": "api", "url": "./api.json"}],
        "workflows": [{
            "workflowId": "test",
            "steps": [{"stepId": "s1", "operationId": "op1"}],
            "x-timeout": 30,
            "x-retry-config": {"maxRetries": 3, "backoff": "exponential"}
        }],
        "x-generator": "my-tool-v1.0"
    }`)

    var doc arazzo1.Arazzo
    if err := json.Unmarshal(data, &doc); err != nil {
        log.Fatal(err)
    }

    // Access extensions
    fmt.Printf("Generator: %v\n", doc.Extensions["x-generator"])
    fmt.Printf("Logo: %v\n", doc.Info.Extensions["x-logo"])
    fmt.Printf("Timeout: %v\n", doc.Workflows[0].Extensions["x-timeout"])

    // Add new extensions
    doc.Extensions["x-custom"] = "custom value"
    doc.Workflows[0].Extensions["x-priority"] = "high"

    // Extensions are preserved when marshaling
    output, _ := json.MarshalIndent(doc, "", "  ")
    fmt.Println(string(output))
}
```

### Working with Success and Failure Actions

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/genelet/arazzo/arazzo1"
)

func main() {
    retryAfter := 2.0
    retryLimit := 3

    workflow := &arazzo1.Workflow{
        WorkflowId: "resilient-workflow",
        Steps: []*arazzo1.Step{
            {
                StepId:      "api-call",
                OperationId: "fetchData",
                SuccessCriteria: []*arazzo1.Criterion{
                    {Condition: "$statusCode == 200"},
                },
                OnSuccess: []*arazzo1.SuccessActionOrReusable{
                    {
                        SuccessAction: &arazzo1.SuccessAction{
                            Name:   "proceed",
                            Type:   arazzo1.SuccessActionTypeGoto,
                            StepId: "process-data",
                        },
                    },
                },
                OnFailure: []*arazzo1.FailureActionOrReusable{
                    {
                        FailureAction: &arazzo1.FailureAction{
                            Name:       "retry-on-timeout",
                            Type:       arazzo1.FailureActionTypeRetry,
                            RetryAfter: &retryAfter,
                            RetryLimit: &retryLimit,
                            Criteria: []*arazzo1.Criterion{
                                {Condition: "$statusCode == 504"},
                            },
                        },
                    },
                    {
                        FailureAction: &arazzo1.FailureAction{
                            Name:       "goto-error-handler",
                            Type:       arazzo1.FailureActionTypeGoto,
                            WorkflowId: "error-handling-workflow",
                            Criteria: []*arazzo1.Criterion{
                                {Condition: "$statusCode >= 500"},
                            },
                        },
                    },
                    {
                        FailureAction: &arazzo1.FailureAction{
                            Name: "end-on-client-error",
                            Type: arazzo1.FailureActionTypeEnd,
                        },
                    },
                },
            },
            {
                StepId:      "process-data",
                OperationId: "processData",
            },
        },
    }

    output, _ := json.MarshalIndent(workflow, "", "  ")
    fmt.Println(string(output))
}
```

### Using Reusable Components

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/genelet/arazzo/arazzo1"
)

func main() {
    doc := &arazzo1.Arazzo{
        Arazzo: "1.0.0",
        Info:   &arazzo1.Info{Title: "API", Version: "1.0.0"},
        SourceDescriptions: []*arazzo1.SourceDescription{
            {Name: "api", URL: "./api.json"},
        },
        Components: &arazzo1.Components{
            Parameters: map[string]*arazzo1.Parameter{
                "AuthHeader": {
                    Name:  "Authorization",
                    In:    arazzo1.ParameterInHeader,
                    Value: "$inputs.authToken",
                },
                "ApiVersion": {
                    Name:  "X-API-Version",
                    In:    arazzo1.ParameterInHeader,
                    Value: "2024-01-01",
                },
            },
            SuccessActions: map[string]*arazzo1.SuccessAction{
                "LogAndEnd": {
                    Name: "log-and-end",
                    Type: arazzo1.SuccessActionTypeEnd,
                },
            },
            FailureActions: map[string]*arazzo1.FailureAction{
                "RetryThreeTimes": {
                    Name:       "retry-3x",
                    Type:       arazzo1.FailureActionTypeRetry,
                    RetryLimit: intPtr(3),
                    RetryAfter: floatPtr(1.0),
                },
            },
        },
        Workflows: []*arazzo1.Workflow{
            {
                WorkflowId: "main-workflow",
                // Reference reusable parameters
                Parameters: []*arazzo1.ParameterOrReusable{
                    {Reusable: &arazzo1.ReusableObject{Reference: "$components.parameters.AuthHeader"}},
                    {Reusable: &arazzo1.ReusableObject{Reference: "$components.parameters.ApiVersion"}},
                },
                Steps: []*arazzo1.Step{
                    {
                        StepId:      "fetch",
                        OperationId: "getData",
                        // Reference reusable failure action with override
                        OnFailure: []*arazzo1.FailureActionOrReusable{
                            {
                                Reusable: &arazzo1.ReusableObject{
                                    Reference: "$components.failureActions.RetryThreeTimes",
                                    Value:     map[string]any{"retryLimit": 5}, // Override
                                },
                            },
                        },
                    },
                },
                // Reference reusable success action
                SuccessActions: []*arazzo1.SuccessActionOrReusable{
                    {Reusable: &arazzo1.ReusableObject{Reference: "$components.successActions.LogAndEnd"}},
                },
            },
        },
    }

    output, _ := json.MarshalIndent(doc, "", "  ")
    fmt.Println(string(output))
}

func intPtr(i int) *int       { return &i }
func floatPtr(f float64) *float64 { return &f }
```

### Using Criterion with Expression Types

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/genelet/arazzo/arazzo1"
)

func main() {
    step := &arazzo1.Step{
        StepId:      "validate-response",
        OperationId: "getUser",
        SuccessCriteria: []*arazzo1.Criterion{
            // Simple condition
            {
                Condition: "$statusCode == 200",
                Type:      arazzo1.CriterionTypeSimple,
            },
            // Regex condition
            {
                Context:   "$response.header.Content-Type",
                Condition: "^application/json.*",
                Type:      arazzo1.CriterionTypeRegex,
            },
            // JSONPath with expression type (includes version)
            {
                Context:   "$response.body",
                Condition: "$.user.email",
                Type:      arazzo1.CriterionTypeJSONPath,
                ExpressionType: &arazzo1.CriterionExpressionType{
                    Type:    arazzo1.CriterionTypeJSONPath,
                    Version: "draft-goessner-dispatch-jsonpath-00",
                },
            },
            // XPath with expression type
            {
                Context:   "$response.body",
                Condition: "//user/email/text()",
                Type:      arazzo1.CriterionTypeXPath,
                ExpressionType: &arazzo1.CriterionExpressionType{
                    Type:    arazzo1.CriterionTypeXPath,
                    Version: "xpath-30",
                },
            },
        },
    }

    output, _ := json.MarshalIndent(step, "", "  ")
    fmt.Println(string(output))
}
```

## Type Reference

### Main Types

| Type | Description |
|------|-------------|
| `Arazzo` | Root document object |
| `Info` | Metadata about the Arazzo description |
| `SourceDescription` | Reference to an OpenAPI or Arazzo document |
| `Workflow` | A workflow with steps |
| `Step` | A single step in a workflow |
| `Parameter` | A parameter for operations or workflows |
| `RequestBody` | Request body for an operation |
| `PayloadReplacement` | Dynamic value replacement in payloads |
| `Criterion` | Success/failure assertion |
| `CriterionExpressionType` | Expression type with version |
| `SuccessAction` | Action on step success |
| `FailureAction` | Action on step failure |
| `ReusableObject` | Reference to a reusable component |
| `Components` | Container for reusable objects |

### Union Types

| Type | Description |
|------|-------------|
| `ParameterOrReusable` | Either a Parameter or ReusableObject |
| `SuccessActionOrReusable` | Either a SuccessAction or ReusableObject |
| `FailureActionOrReusable` | Either a FailureAction or ReusableObject |

### Enum Constants

```go
// Source Description Types
arazzo1.SourceDescriptionTypeArazzo  // "arazzo"
arazzo1.SourceDescriptionTypeOpenAPI // "openapi"

// Parameter Locations
arazzo1.ParameterInPath   // "path"
arazzo1.ParameterInQuery  // "query"
arazzo1.ParameterInHeader // "header"
arazzo1.ParameterInCookie // "cookie"

// Criterion Types
arazzo1.CriterionTypeSimple   // "simple"
arazzo1.CriterionTypeRegex    // "regex"
arazzo1.CriterionTypeJSONPath // "jsonpath"
arazzo1.CriterionTypeXPath    // "xpath"

// Success Action Types
arazzo1.SuccessActionTypeEnd  // "end"
arazzo1.SuccessActionTypeGoto // "goto"

// Failure Action Types
arazzo1.FailureActionTypeEnd   // "end"
arazzo1.FailureActionTypeGoto  // "goto"
arazzo1.FailureActionTypeRetry // "retry"
```

## HCL Format Support

The `convert` package provides functions to convert Arazzo documents between JSON and HCL formats using [genelet/horizon](https://github.com/genelet/horizon).

### Converting JSON to HCL

```go
package main

import (
    "fmt"
    "log"

    "github.com/genelet/arazzo/convert"
)

func main() {
    jsonData := []byte(`{
        "arazzo": "1.0.0",
        "info": {"title": "My Workflow", "version": "1.0.0"},
        "sourceDescriptions": [
            {"name": "petstore", "url": "./openapi.json", "type": "openapi"}
        ],
        "workflows": [
            {
                "workflowId": "get-pet",
                "steps": [{"stepId": "fetch", "operationId": "getPetById"}]
            }
        ]
    }`)

    hclData, err := convert.JSONToHCL(jsonData)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(hclData))
}
```

Output:
```hcl
arazzo = "1.0.0"

info {
  title   = "My Workflow"
  version = "1.0.0"
}

sourceDescription "petstore" {
  url  = "./openapi.json"
  type = "openapi"
}

workflow "get-pet" {
  step "fetch" {
    operationId = "getPetById"
  }
}
```

### Converting HCL to JSON

```go
package main

import (
    "fmt"
    "log"

    "github.com/genelet/arazzo/convert"
)

func main() {
    hclData := []byte(`
arazzo = "1.0.0"

info {
  title   = "Pet Store Workflow"
  version = "1.0.0"
}

sourceDescription "petstore" {
  url  = "./openapi.json"
  type = "openapi"
}

workflow "get-pet" {
  step "fetch-pet" {
    operationId = "getPetById"
  }
}
`)

    jsonData, err := convert.HCLToJSONIndent(hclData, "", "  ")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(jsonData))
}
```

### Direct Marshal/Unmarshal with HCL

```go
package main

import (
    "fmt"
    "log"

    "github.com/genelet/arazzo/arazzo1"
    "github.com/genelet/arazzo/convert"
)

func main() {
    // Create a document
    doc := &arazzo1.Arazzo{
        Arazzo: "1.0.0",
        Info: &arazzo1.Info{
            Title:   "My API",
            Version: "1.0.0",
        },
        SourceDescriptions: []*arazzo1.SourceDescription{
            {Name: "api", URL: "./api.json"},
        },
        Workflows: []*arazzo1.Workflow{
            {
                WorkflowId: "test",
                Steps: []*arazzo1.Step{
                    {StepId: "s1", OperationId: "op1"},
                },
            },
        },
    }

    // Marshal to HCL
    hclData, err := convert.MarshalHCL(doc)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(hclData))

    // Unmarshal from HCL
    var doc2 arazzo1.Arazzo
    if err := convert.UnmarshalHCL(hclData, &doc2); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Parsed: %s\n", doc2.Info.Title)
}
```

### Convert Package Functions

| Function | Description |
|----------|-------------|
| `JSONToHCL(jsonData []byte)` | Convert JSON to HCL |
| `HCLToJSON(hclData []byte)` | Convert HCL to JSON |
| `HCLToJSONIndent(hclData []byte, prefix, indent string)` | Convert HCL to indented JSON |
| `MarshalHCL(doc *arazzo1.Arazzo)` | Marshal Arazzo document to HCL |
| `UnmarshalHCL(hclData []byte, doc *arazzo1.Arazzo)` | Unmarshal HCL to Arazzo document |
| `MarshalJSON(doc *arazzo1.Arazzo)` | Marshal Arazzo document to JSON |
| `MarshalJSONIndent(doc *arazzo1.Arazzo, prefix, indent string)` | Marshal to indented JSON |
| `UnmarshalJSON(jsonData []byte, doc *arazzo1.Arazzo)` | Unmarshal JSON to Arazzo document |

### HCL Conversion Notes

**JSON Schema `$ref` Handling**: JSON Schema keys starting with `$` (like `$ref`, `$id`, `$schema`) are automatically transformed to use `_` prefix (e.g., `_ref`) when converting to HCL, since `$` is not valid in HCL identifiers. The transformation is reversed when converting back to JSON.

**Round-Trip Limitations**: The following scenarios may not round-trip perfectly through HCL:

| Limitation | Description |
|------------|-------------|
| Multi-line strings | Long descriptions with embedded newlines in JSON Schema definitions may not parse correctly in HCL |
| Numeric values in `any` fields | Integer/float values in dynamically-typed fields (like `Parameter.Value` in components) may be output without the `=` sign |
| `Workflow.Inputs` with simple `$ref` | When `inputs` contains only a `$ref`, it's rendered as an HCL block but may not parse back into the `any` typed field |

For full fidelity round-trips, use JSON format. HCL is best suited for human-authored configuration where these edge cases are avoided.

## Validation

The `Validate()` method performs comprehensive validation:

- Required fields presence
- Pattern matching (arazzo version, names)
- Enum value validation
- Mutual exclusivity (e.g., step must have exactly one of operationId/operationPath/workflowId)
- Conditional requirements (e.g., goto action requires stepId or workflowId)
- Component name patterns
- Nested object validation

```go
result := doc.Validate()
if !result.Valid() {
    for _, err := range result.Errors {
        fmt.Printf("%s: %s\n", err.Path, err.Message)
    }
}
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [genelet/oas](https://github.com/genelet/oas) - Go parser for OpenAPI 3.0 and 3.1 specifications
- [genelet/horizon](https://github.com/genelet/horizon) - HCL parsing library used for HCL format support
- [Arazzo Specification](https://spec.openapis.org/arazzo/v1.0.0) - Official Arazzo specification
