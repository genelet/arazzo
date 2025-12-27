package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/genelet/arazzo/arazzo1"
	"github.com/genelet/horizon/dethcl"
	"github.com/genelet/oas/openapi31"
	"gopkg.in/yaml.v3"
)

// parseOpenAPI parses an OpenAPI file, handling both JSON and YAML.
// Since openapi31 relies on UnmarshalJSON for custom logic, we convert YAML to JSON first.
func parseOpenAPI(content []byte) (*openapi31.OpenAPI, error) {
	// 1. Try JSON directly
	var doc openapi31.OpenAPI
	if err := json.Unmarshal(content, &doc); err == nil {
		return &doc, nil
	}

	// 2. Try YAML -> Interface -> JSON -> Struct
	var obj interface{}
	if err := yaml.Unmarshal(content, &obj); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("converting yaml to json: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &doc); err != nil {
		return nil, fmt.Errorf("parsing converted json: %w", err)
	}

	return &doc, nil
}

// NewArazzoFromFiles creates an Arazzo document from OpenAPI and Generator files.
func NewArazzoFromFiles(openapiFile, generatorFile string, format ...string) (*arazzo1.Arazzo, error) {
	// Parse Generator
	genBytes, err := os.ReadFile(generatorFile)
	if err != nil {
		return nil, fmt.Errorf("reading generator file: %w", err)
	}
	var gen Generator

	fmtType := "yaml"
	if len(format) > 0 {
		fmtType = format[0]
	}

	switch fmtType {
	case "json":
		if err := json.Unmarshal(genBytes, &gen); err != nil {
			return nil, fmt.Errorf("parsing generator file (json): %w", err)
		}
	case "hcl":
		if err := dethcl.Unmarshal(genBytes, &gen); err != nil {
			return nil, fmt.Errorf("parsing generator file (hcl): %w", err)
		}
	default: // yaml
		if err := yaml.Unmarshal(genBytes, &gen); err != nil {
			return nil, fmt.Errorf("parsing generator file (yaml): %w", err)
		}
	}

	// Parse OpenAPI
	oaBytes, err := os.ReadFile(openapiFile)
	if err != nil {
		return nil, fmt.Errorf("reading openapi file: %w", err)
	}
	doc, err := parseOpenAPI(oaBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing openapi file: %w", err)
	}
	gen.openapiDoc = doc

	return gen.ToArazzo(openapiFile)
}

// NewGeneratorFromArazzo creates a Generator config from Arazzo and OpenAPI files.
func NewGeneratorFromArazzo(arazzoFile, openapiFile string) (*Generator, error) {
	// Parse Arazzo
	azBytes, err := os.ReadFile(arazzoFile)
	if err != nil {
		return nil, fmt.Errorf("reading arazzo file: %w", err)
	}
	var az arazzo1.Arazzo
	// Arazzo can be JSON or YAML
	if err := json.Unmarshal(azBytes, &az); err != nil {
		if err := yaml.Unmarshal(azBytes, &az); err != nil {
			return nil, fmt.Errorf("parsing arazzo file: %w", err)
		}
	}

	// Parse OpenAPI
	oaBytes, err := os.ReadFile(openapiFile)
	if err != nil {
		return nil, fmt.Errorf("reading openapi file: %w", err)
	}
	doc, err := parseOpenAPI(oaBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing openapi file: %w", err)
	}

	gen := &Generator{
		openapiDoc: doc,
		Provider: &Provider{
			Name:      "my-source", // Default
			ServerURL: "",          // Default empty
		},
	}

	if len(doc.Servers) > 0 {
		gen.Provider.ServerURL = doc.Servers[0].URL
	}

	// Try to get source name from Arazzo if possible
	if len(az.SourceDescriptions) > 0 {
		gen.Provider.Name = az.SourceDescriptions[0].Name
	}

	// Reserve Info fields in Appendices
	if az.Info != nil {
		gen.Provider.Appendices = make(map[string]interface{})
		gen.Provider.Appendices["info_title"] = az.Info.Title
		gen.Provider.Appendices["info_version"] = az.Info.Version
		gen.Provider.Appendices["info_summary"] = az.Info.Summary
		gen.Provider.Appendices["info_description"] = az.Info.Description
	}

	// Iterate workflows to build operations
	for _, wf := range az.Workflows {
		// Populate high-fidelity workflow fields
		gen.Name = wf.WorkflowId
		gen.Summary = wf.Summary
		gen.Description = wf.Description
		gen.Inputs = wf.Inputs
		if wf.Outputs != nil {
			gen.Outputs = make(map[string]interface{})
			for k, v := range wf.Outputs {
				gen.Outputs[k] = v
			}
		}

		gen.DependsOn = wf.DependsOn
		// Copy workflow parameters
		for _, pOrRef := range wf.Parameters {
			if pOrRef.Parameter != nil {
				gen.Parameters = append(gen.Parameters, pOrRef.Parameter)
			}
			// Note: Reusable parameters are ignored for now if not inline
		}

		for _, step := range wf.Steps {
			// Expected format: $source.operationId or $sourceDescriptions.name.operationId
			opID := step.OperationId
			if idx := strings.LastIndex(opID, "."); idx != -1 {
				opID = opID[idx+1:]
			}

			// Handle OperationPath if OperationId is empty
			opPath := step.OperationPath

			if opID == "" && opPath == "" {
				continue // Skip if neither ID nor Path is useful
			}

			op := &OperationSpec{
				Name:          step.StepId, // Use StepId for the Name (critical for references)
				Description:   step.Description,
				OperationPath: step.OperationPath,
				OperationId:   step.OperationId, // Original full ID
			}
			// (op.Summary removed from struct above/init logic if not supported by Step, but I added it to struct. Step DOES NOT have Summary in Arazzo struct. So leaving empty.)

			// Copy parameters
			for _, pRaw := range step.Parameters {
				// pRaw is interface{}, likely map[string]interface{}
				// We want to convert it to *arazzo1.Parameter
				// Best way is JSON round-trip for robustness
				b, err := json.Marshal(pRaw)
				if err == nil {
					var p arazzo1.Parameter
					if err := json.Unmarshal(b, &p); err == nil {
						op.Parameters = append(op.Parameters, &p)
					}
				}
			}

			// Copy outputs
			if step.Outputs != nil {
				op.Outputs = make(map[string]interface{})
				for k, v := range step.Outputs {
					op.Outputs[k] = v
				}
			}

			// Copy new fields
			op.RequestBody = step.RequestBody
			op.SuccessCriteria = step.SuccessCriteria
			op.OnSuccess = step.OnSuccess
			op.OnFailure = step.OnFailure

			gen.HTTP = append(gen.HTTP, op)
		}
	}

	return gen, nil
}

// ToArazzo converts the generator configuration and OpenAPI document to an Arazzo object.
func (g *Generator) ToArazzo(openapiFilename string) (*arazzo1.Arazzo, error) {
	if g.openapiDoc == nil {
		return nil, fmt.Errorf("openapi document not set")
	}

	// Create Arazzo root
	arazzo := &arazzo1.Arazzo{
		Arazzo: "1.0.0",
		Info: &arazzo1.Info{
			Title:   "Generated Arazzo from " + g.openapiDoc.Info.Title,
			Version: "1.0.0",
			Summary: "Generated from " + openapiFilename,
		},
		SourceDescriptions: []*arazzo1.SourceDescription{
			{
				Name: g.Provider.Name,
				URL:  openapiFilename,
				Type: arazzo1.SourceDescriptionTypeOpenAPI,
			},
		},
		Workflows: []*arazzo1.Workflow{
			{
				WorkflowId: "main-workflow",
				Summary:    "Main workflow containing all operations",
				Steps:      []*arazzo1.Step{},
			},
		},
	}

	// Restore Info from Appendices if available
	if g.Provider.Appendices != nil {
		if v, ok := g.Provider.Appendices["info_title"].(string); ok && v != "" {
			arazzo.Info.Title = v
		}
		if v, ok := g.Provider.Appendices["info_version"].(string); ok && v != "" {
			arazzo.Info.Version = v
		}
		if v, ok := g.Provider.Appendices["info_summary"].(string); ok && v != "" {
			arazzo.Info.Summary = v
		}
		if v, ok := g.Provider.Appendices["info_description"].(string); ok && v != "" {
			arazzo.Info.Description = v
		}
	}

	// Collect all operations
	// Collect all operations
	var operations []*OperationSpec
	operations = append(operations, g.HTTP...)

	if len(operations) == 0 {
		return nil, fmt.Errorf("no operations found in generator config")
	}

	// Create steps
	for _, op := range operations {
		step := &arazzo1.Step{
			StepId:      op.Name,
			Description: op.Description,
			// parameters and outputs below
		}

		if len(op.SuccessCriteria) > 0 {
			step.SuccessCriteria = op.SuccessCriteria
		} else {
			step.SuccessCriteria = []*arazzo1.Criterion{
				{
					Condition: "$statusCode == 200",
				},
			}
		}

		step.RequestBody = op.RequestBody
		step.OnSuccess = op.OnSuccess
		step.OnFailure = op.OnFailure

		// Handle Operation Reference
		if op.OperationId != "" {
			step.OperationId = op.OperationId
		} else if op.OperationPath != "" {
			step.OperationPath = op.OperationPath
		} else {
			// Default fallback
			step.OperationId = "$source." + op.Name
		}

		// Handle Parameters
		if len(op.Parameters) > 0 {
			step.Parameters = make([]any, len(op.Parameters))
			for i, p := range op.Parameters {
				step.Parameters[i] = p
			}
		}

		// Handle Outputs
		if len(op.Outputs) > 0 {
			step.Outputs = make(map[string]string)
			for k, v := range op.Outputs {
				step.Outputs[k] = fmt.Sprint(v)
			}
		}

		arazzo.Workflows[0].Steps = append(arazzo.Workflows[0].Steps, step)
	}

	// Update Workflow details from Generator if present
	if g.Name != "" {
		arazzo.Workflows[0].WorkflowId = g.Name
	}
	if g.Summary != "" {
		arazzo.Workflows[0].Summary = g.Summary
	}
	if g.Description != "" {
		arazzo.Workflows[0].Description = g.Description
	}
	if g.Inputs != nil {
		arazzo.Workflows[0].Inputs = g.Inputs
	}
	if len(g.Outputs) > 0 {
		arazzo.Workflows[0].Outputs = make(map[string]string)
		for k, v := range g.Outputs {
			arazzo.Workflows[0].Outputs[k] = fmt.Sprint(v)
		}
	}
	if len(g.DependsOn) > 0 {
		arazzo.Workflows[0].DependsOn = g.DependsOn
	}
	if len(g.Parameters) > 0 {
		for _, p := range g.Parameters {
			arazzo.Workflows[0].Parameters = append(arazzo.Workflows[0].Parameters, &arazzo1.ParameterOrReusable{Parameter: p})
		}
	}

	return arazzo, nil
}
