package arazzo1

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a single validation error.
type ValidationError struct {
	Path    string
	Message string
}

// ValidationResult holds the results of validating an Arazzo document.
type ValidationResult struct {
	Errors []ValidationError
}

// Valid returns true if there are no validation errors.
func (r *ValidationResult) Valid() bool {
	return len(r.Errors) == 0
}

// Error returns a string representation of all validation errors.
func (r *ValidationResult) Error() string {
	if r.Valid() {
		return ""
	}
	var msgs []string
	for _, err := range r.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Path, err.Message))
	}
	return strings.Join(msgs, "; ")
}

func (r *ValidationResult) addError(path, message string) {
	r.Errors = append(r.Errors, ValidationError{Path: path, Message: message})
}

var (
	// arazzoVersionPattern matches Arazzo version numbers (1.0.x with optional prerelease)
	arazzoVersionPattern = regexp.MustCompile(`^1\.0\.\d+(-.+)?$`)

	// sourceNamePattern matches source description names
	sourceNamePattern = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)

	// componentNamePattern matches component names
	componentNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`)

	// outputNamePattern matches output names
	outputNamePattern = regexp.MustCompile(`^[a-zA-Z0-9\.\-_]+$`)
)

// Validate validates the Arazzo document and returns a ValidationResult.
func (a *Arazzo) Validate() *ValidationResult {
	result := &ValidationResult{}

	// Required fields
	if a.Arazzo == "" {
		result.addError("arazzo", "required field is missing")
	} else if !arazzoVersionPattern.MatchString(a.Arazzo) {
		result.addError("arazzo", fmt.Sprintf("must match pattern ^1\\.0\\.\\d+(-.+)?$; got %s", a.Arazzo))
	}

	if a.Info == nil {
		result.addError("info", "required field is missing")
	} else {
		a.Info.validate("info", result)
	}

	if len(a.SourceDescriptions) == 0 {
		result.addError("sourceDescriptions", "required field is missing or empty (minItems: 1)")
	} else {
		names := make(map[string]bool)
		for i, sd := range a.SourceDescriptions {
			if sd != nil {
				sd.validate(fmt.Sprintf("sourceDescriptions[%d]", i), result)
				if sd.Name != "" {
					if names[sd.Name] {
						result.addError(fmt.Sprintf("sourceDescriptions[%d].name", i),
							fmt.Sprintf("duplicate source description name: %s", sd.Name))
					}
					names[sd.Name] = true
				}
			}
		}
	}

	if len(a.Workflows) == 0 {
		result.addError("workflows", "required field is missing or empty (minItems: 1)")
	} else {
		workflowIds := make(map[string]bool)
		for i, w := range a.Workflows {
			if w != nil {
				w.validate(fmt.Sprintf("workflows[%d]", i), result)
				if w.WorkflowId != "" {
					if workflowIds[w.WorkflowId] {
						result.addError(fmt.Sprintf("workflows[%d].workflowId", i),
							fmt.Sprintf("duplicate workflowId: %s", w.WorkflowId))
					}
					workflowIds[w.WorkflowId] = true
				}
			}
		}
	}

	if a.Components != nil {
		a.Components.validate("components", result)
	}

	return result
}

func (i *Info) validate(path string, result *ValidationResult) {
	if i.Title == "" {
		result.addError(path+".title", "required field is missing")
	}
	if i.Version == "" {
		result.addError(path+".version", "required field is missing")
	}
}

func (s *SourceDescription) validate(path string, result *ValidationResult) {
	if s.Name == "" {
		result.addError(path+".name", "required field is missing")
	} else if !sourceNamePattern.MatchString(s.Name) {
		result.addError(path+".name", fmt.Sprintf("must match pattern ^[A-Za-z0-9_\\-]+$; got %s", s.Name))
	}

	if s.URL == "" {
		result.addError(path+".url", "required field is missing")
	}

	if s.Type != "" && s.Type != SourceDescriptionTypeArazzo && s.Type != SourceDescriptionTypeOpenAPI {
		result.addError(path+".type", fmt.Sprintf("must be 'arazzo' or 'openapi'; got %s", s.Type))
	}
}

func (w *Workflow) validate(path string, result *ValidationResult) {
	if w.WorkflowId == "" {
		result.addError(path+".workflowId", "required field is missing")
	}

	if len(w.Steps) == 0 {
		result.addError(path+".steps", "required field is missing or empty (minItems: 1)")
	} else {
		stepIds := make(map[string]bool)
		for i, step := range w.Steps {
			if step != nil {
				step.validate(fmt.Sprintf("%s.steps[%d]", path, i), result)
				if step.StepId != "" {
					if stepIds[step.StepId] {
						result.addError(fmt.Sprintf("%s.steps[%d].stepId", path, i),
							fmt.Sprintf("duplicate stepId: %s", step.StepId))
					}
					stepIds[step.StepId] = true
				}
			}
		}
	}

	// Validate outputs keys
	for key := range w.Outputs {
		if !outputNamePattern.MatchString(key) {
			result.addError(fmt.Sprintf("%s.outputs.%s", path, key),
				fmt.Sprintf("output name must match pattern ^[a-zA-Z0-9\\.\\-_]+$; got %s", key))
		}
	}

	// Validate successActions
	for i, action := range w.SuccessActions {
		if action != nil && action.SuccessAction != nil {
			action.SuccessAction.validate(fmt.Sprintf("%s.successActions[%d]", path, i), result)
		}
	}

	// Validate failureActions
	for i, action := range w.FailureActions {
		if action != nil && action.FailureAction != nil {
			action.FailureAction.validate(fmt.Sprintf("%s.failureActions[%d]", path, i), result)
		}
	}

	// Validate parameters
	for i, param := range w.Parameters {
		if param != nil && param.Parameter != nil {
			param.Parameter.validate(fmt.Sprintf("%s.parameters[%d]", path, i), result)
		}
	}
}

func (s *Step) validate(path string, result *ValidationResult) {
	if s.StepId == "" {
		result.addError(path+".stepId", "required field is missing")
	}

	// Must have exactly one of operationId, operationPath, or workflowId
	count := 0
	if s.OperationId != "" {
		count++
	}
	if s.OperationPath != "" {
		count++
	}
	if s.WorkflowId != "" {
		count++
	}

	if count == 0 {
		result.addError(path, "must have one of: operationId, operationPath, or workflowId")
	} else if count > 1 {
		result.addError(path, "must have only one of: operationId, operationPath, or workflowId")
	}

	// Validate outputs keys
	for key := range s.Outputs {
		if !outputNamePattern.MatchString(key) {
			result.addError(fmt.Sprintf("%s.outputs.%s", path, key),
				fmt.Sprintf("output name must match pattern ^[a-zA-Z0-9\\.\\-_]+$; got %s", key))
		}
	}

	// Validate requestBody
	if s.RequestBody != nil {
		s.RequestBody.validate(path+".requestBody", result)
	}

	// Validate successCriteria
	for i, criterion := range s.SuccessCriteria {
		if criterion != nil {
			criterion.validate(fmt.Sprintf("%s.successCriteria[%d]", path, i), result)
		}
	}

	// Validate onSuccess
	for i, action := range s.OnSuccess {
		if action != nil && action.SuccessAction != nil {
			action.SuccessAction.validate(fmt.Sprintf("%s.onSuccess[%d]", path, i), result)
		}
	}

	// Validate onFailure
	for i, action := range s.OnFailure {
		if action != nil && action.FailureAction != nil {
			action.FailureAction.validate(fmt.Sprintf("%s.onFailure[%d]", path, i), result)
		}
	}
}

func (p *Parameter) validate(path string, result *ValidationResult) {
	if p.Name == "" {
		result.addError(path+".name", "required field is missing")
	}
	if p.Value == nil {
		result.addError(path+".value", "required field is missing")
	}

	if p.In != "" {
		validIn := map[ParameterIn]bool{
			ParameterInPath:   true,
			ParameterInQuery:  true,
			ParameterInHeader: true,
			ParameterInCookie: true,
		}
		if !validIn[p.In] {
			result.addError(path+".in",
				fmt.Sprintf("must be one of: path, query, header, cookie; got %s", p.In))
		}
	}
}

func (r *RequestBody) validate(path string, result *ValidationResult) {
	for i, replacement := range r.Replacements {
		if replacement != nil {
			replacement.validate(fmt.Sprintf("%s.replacements[%d]", path, i), result)
		}
	}
}

func (p *PayloadReplacement) validate(path string, result *ValidationResult) {
	if p.Target == "" {
		result.addError(path+".target", "required field is missing")
	}
	if p.Value == "" {
		result.addError(path+".value", "required field is missing")
	}
}

func (c *Criterion) validate(path string, result *ValidationResult) {
	if c.Condition == "" {
		result.addError(path+".condition", "required field is missing")
	}

	// If type is set, context is required
	if c.Type != "" && c.Context == "" {
		result.addError(path+".context", "required when type is specified")
	}

	// Validate type values
	if c.Type != "" {
		validTypes := map[CriterionType]bool{
			CriterionTypeSimple:   true,
			CriterionTypeRegex:    true,
			CriterionTypeJSONPath: true,
			CriterionTypeXPath:    true,
		}
		if !validTypes[c.Type] {
			result.addError(path+".type",
				fmt.Sprintf("must be one of: simple, regex, jsonpath, xpath; got %s", c.Type))
		}
	}

	// Validate expression type if present
	if c.ExpressionType != nil {
		c.ExpressionType.validate(path, result)
	}
}

func (c *CriterionExpressionType) validate(path string, result *ValidationResult) {
	if c.Type == "" {
		result.addError(path+".type", "required field is missing")
	} else if c.Type != CriterionTypeJSONPath && c.Type != CriterionTypeXPath {
		result.addError(path+".type",
			fmt.Sprintf("must be 'jsonpath' or 'xpath' for expression type; got %s", c.Type))
	}

	if c.Version == "" {
		result.addError(path+".version", "required field is missing")
	} else {
		// Validate version based on type
		if c.Type == CriterionTypeJSONPath && c.Version != "draft-goessner-dispatch-jsonpath-00" {
			result.addError(path+".version",
				fmt.Sprintf("for jsonpath type, must be 'draft-goessner-dispatch-jsonpath-00'; got %s", c.Version))
		}
		if c.Type == CriterionTypeXPath {
			validVersions := map[string]bool{
				"xpath-10": true,
				"xpath-20": true,
				"xpath-30": true,
			}
			if !validVersions[c.Version] {
				result.addError(path+".version",
					fmt.Sprintf("for xpath type, must be one of: xpath-10, xpath-20, xpath-30; got %s", c.Version))
			}
		}
	}
}

func (s *SuccessAction) validate(path string, result *ValidationResult) {
	if s.Name == "" {
		result.addError(path+".name", "required field is missing")
	}
	if s.Type == "" {
		result.addError(path+".type", "required field is missing")
	} else if s.Type != SuccessActionTypeEnd && s.Type != SuccessActionTypeGoto {
		result.addError(path+".type",
			fmt.Sprintf("must be 'end' or 'goto'; got %s", s.Type))
	}

	// If type is goto, must have workflowId or stepId
	if s.Type == SuccessActionTypeGoto {
		if s.WorkflowId == "" && s.StepId == "" {
			result.addError(path, "goto action requires either workflowId or stepId")
		}
		if s.WorkflowId != "" && s.StepId != "" {
			result.addError(path, "goto action cannot have both workflowId and stepId")
		}
	}

	// Validate criteria
	for i, criterion := range s.Criteria {
		if criterion != nil {
			criterion.validate(fmt.Sprintf("%s.criteria[%d]", path, i), result)
		}
	}
}

func (f *FailureAction) validate(path string, result *ValidationResult) {
	if f.Name == "" {
		result.addError(path+".name", "required field is missing")
	}
	if f.Type == "" {
		result.addError(path+".type", "required field is missing")
	} else if f.Type != FailureActionTypeEnd && f.Type != FailureActionTypeGoto && f.Type != FailureActionTypeRetry {
		result.addError(path+".type",
			fmt.Sprintf("must be 'end', 'goto', or 'retry'; got %s", f.Type))
	}

	// If type is goto, must have workflowId or stepId
	if f.Type == FailureActionTypeGoto {
		if f.WorkflowId == "" && f.StepId == "" {
			result.addError(path, "goto action requires either workflowId or stepId")
		}
		if f.WorkflowId != "" && f.StepId != "" {
			result.addError(path, "goto action cannot have both workflowId and stepId")
		}
	}

	// Validate retry fields
	if f.RetryAfter != nil && *f.RetryAfter < 0 {
		result.addError(path+".retryAfter", "must be non-negative")
	}
	if f.RetryLimit != nil && *f.RetryLimit < 0 {
		result.addError(path+".retryLimit", "must be non-negative")
	}

	// Validate criteria
	for i, criterion := range f.Criteria {
		if criterion != nil {
			criterion.validate(fmt.Sprintf("%s.criteria[%d]", path, i), result)
		}
	}
}

func (c *Components) validate(path string, result *ValidationResult) {
	// Validate component names
	for name := range c.Inputs {
		if !componentNamePattern.MatchString(name) {
			result.addError(fmt.Sprintf("%s.inputs.%s", path, name),
				fmt.Sprintf("component name must match pattern ^[a-zA-Z0-9\\.\\-_]+$; got %s", name))
		}
	}

	for name, param := range c.Parameters {
		if !componentNamePattern.MatchString(name) {
			result.addError(fmt.Sprintf("%s.parameters.%s", path, name),
				fmt.Sprintf("component name must match pattern ^[a-zA-Z0-9\\.\\-_]+$; got %s", name))
		}
		if param != nil {
			param.validate(fmt.Sprintf("%s.parameters.%s", path, name), result)
		}
	}

	for name, action := range c.SuccessActions {
		if !componentNamePattern.MatchString(name) {
			result.addError(fmt.Sprintf("%s.successActions.%s", path, name),
				fmt.Sprintf("component name must match pattern ^[a-zA-Z0-9\\.\\-_]+$; got %s", name))
		}
		if action != nil {
			action.validate(fmt.Sprintf("%s.successActions.%s", path, name), result)
		}
	}

	for name, action := range c.FailureActions {
		if !componentNamePattern.MatchString(name) {
			result.addError(fmt.Sprintf("%s.failureActions.%s", path, name),
				fmt.Sprintf("component name must match pattern ^[a-zA-Z0-9\\.\\-_]+$; got %s", name))
		}
		if action != nil {
			action.validate(fmt.Sprintf("%s.failureActions.%s", path, name), result)
		}
	}
}
