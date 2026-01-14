package arazzo1

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Workflow describes the steps to be taken across one or more APIs to achieve an objective.
type Workflow struct {
	// WorkflowId is a unique string to represent the workflow (required)
	WorkflowId string `json:"workflowId" yaml:"workflowId" hcl:"workflowId,label"`

	// Summary is a summary of the purpose or objective of the workflow
	Summary string `json:"summary,omitempty" yaml:"summary,omitempty" hcl:"summary,optional"`

	// Description of the workflow. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty" yaml:"description,omitempty" hcl:"description,optional"`

	// Inputs is a JSON Schema 2020-12 object representing the input parameters used by this workflow
	Inputs any `json:"inputs,omitempty" yaml:"inputs,omitempty" hcl:"inputs,optional"`

	// DependsOn is a list of workflows that MUST be completed before this workflow can be processed
	DependsOn []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty" hcl:"dependsOn,optional"`

	// Steps is an ordered list of steps where each step represents a call to an API operation
	// or to another workflow (required, minItems: 1)
	Steps []*Step `json:"steps" yaml:"steps" hcl:"step,block"`

	// SuccessActions is a list of success actions that are applicable for all steps
	// described under this workflow
	SuccessActions []*SuccessActionOrReusable `json:"successActions,omitempty" yaml:"successActions,omitempty" hcl:"successAction,block"`

	// FailureActions is a list of failure actions that are applicable for all steps
	// described under this workflow
	FailureActions []*FailureActionOrReusable `json:"failureActions,omitempty" yaml:"failureActions,omitempty" hcl:"failureAction,block"`

	// Outputs is a map between a friendly name and a dynamic output value
	// Pattern for keys: ^[a-zA-Z0-9\.\-_]+$
	Outputs map[string]string `json:"outputs,omitempty" yaml:"outputs,omitempty" hcl:"outputs,optional"`

	// Parameters is a list of parameters that are applicable for all steps
	// described under this workflow
	Parameters []*ParameterOrReusable `json:"parameters,omitempty" yaml:"parameters,omitempty" hcl:"parameter,block"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type workflowAlias Workflow

var workflowKnownFields = []string{
	"workflowId",
	"summary",
	"description",
	"inputs",
	"dependsOn",
	"steps",
	"successActions",
	"failureActions",
	"outputs",
	"parameters",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (w *Workflow) UnmarshalJSON(data []byte) error {
	var alias workflowAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return fmt.Errorf("unmarshaling workflow: %w", err)
	}
	*w = Workflow(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshaling workflow extensions: %w", err)
	}
	w.Extensions = extractExtensions(raw, workflowKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (w Workflow) MarshalJSON() ([]byte, error) {
	alias := workflowAlias(w)
	return marshalWithExtensions(&alias, w.Extensions)
}

// UnmarshalHCL implements the dethcl.Unmarshaler interface.
// This custom unmarshaler handles the Inputs field which is typed as `any`
// and needs special handling to parse HCL blocks into map[string]any.
func (w *Workflow) UnmarshalHCL(data []byte, labels ...string) error {
	// Parse HCL
	file, diags := hclsyntax.ParseConfig(data, "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("parsing HCL: %w", diags)
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return fmt.Errorf("unexpected HCL body type: %T", file.Body)
	}

	// Set label (workflowId) if provided
	if len(labels) > 0 {
		w.WorkflowId = labels[0]
	}

	// Accumulate errors for attributes
	var errs []string

	// Process attributes
	for name, attr := range body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}

		switch name {
		case "summary":
			w.Summary = val.AsString()
		case "description":
			w.Description = val.AsString()
		case "dependsOn":
			w.DependsOn = ctyToStringSlice(val)
		case "outputs":
			w.Outputs = ctyToStringMap(val)
		}
	}

	// Process blocks
	for _, block := range body.Blocks {
		switch block.Type {
		case "inputs":
			w.Inputs = hclBlockToMap(block)
		case "step":
			step := &Step{}
			if err := parseStepBlock(block, step); err != nil {
				errs = append(errs, fmt.Sprintf("step block: %s", err.Error()))
			}
			w.Steps = append(w.Steps, step)
		case "successAction":
			action := &SuccessActionOrReusable{SuccessAction: &SuccessAction{}}
			if len(block.Labels) > 0 {
				action.SuccessAction.Name = block.Labels[0]
			}
			// Parse the block content for successAction
			if err := parseSuccessActionBlock(block, action.SuccessAction); err != nil {
				errs = append(errs, fmt.Sprintf("successAction block: %s", err.Error()))
			}
			w.SuccessActions = append(w.SuccessActions, action)
		case "failureAction":
			action := &FailureActionOrReusable{FailureAction: &FailureAction{}}
			if len(block.Labels) > 0 {
				action.FailureAction.Name = block.Labels[0]
			}
			if err := parseFailureActionBlock(block, action.FailureAction); err != nil {
				errs = append(errs, fmt.Sprintf("failureAction block: %s", err.Error()))
			}
			w.FailureActions = append(w.FailureActions, action)
		case "parameter":
			param := &ParameterOrReusable{Parameter: &Parameter{}}
			if len(block.Labels) > 0 {
				param.Parameter.Name = block.Labels[0]
			}
			if err := parseParameterBlock(block, param.Parameter); err != nil {
				errs = append(errs, fmt.Sprintf("parameter block: %s", err.Error()))
			}
			w.Parameters = append(w.Parameters, param)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("HCL unmarshal errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// hclBlockToMap converts an HCL block to a map[string]any.
// Attributes with evaluation errors are skipped and logged as warnings via the returned error slice.
func hclBlockToMap(block *hclsyntax.Block) map[string]any {
	result := make(map[string]any)

	// Process attributes - skip those with errors but the calling code
	// should be aware these attributes were not processed
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			// Attribute requires context evaluation; store placeholder
			result[name] = nil
			continue
		}
		result[name] = ctyToGo(val)
	}

	// Process nested blocks
	for _, nestedBlock := range block.Body.Blocks {
		blockName := nestedBlock.Type
		if len(nestedBlock.Labels) > 0 {
			blockName = nestedBlock.Labels[0]
		}
		result[blockName] = hclBlockToMap(nestedBlock)
	}

	return result
}

// hclBlockToBytes is deprecated and not used.
// This function was intended to convert an HCL block body to bytes for nested unmarshaling,
// but we handle nested blocks directly through the specific parsing functions
// (parseStepBlock, parseSuccessActionBlock, etc.) instead of re-serializing to bytes.
// This approach is more efficient and avoids double parsing.
//
// Deprecated: Use the specific parse*Block functions directly instead.
func hclBlockToBytes(_ *hclsyntax.Block) []byte {
	return nil
}

// ctyToGo converts a cty.Value to a Go value
func ctyToGo(val cty.Value) any {
	if val.IsNull() {
		return nil
	}

	switch {
	case val.Type() == cty.String:
		return val.AsString()
	case val.Type() == cty.Number:
		bf := val.AsBigFloat()
		f, acc := bf.Float64()
		if acc != 0 {
			// Precision was lost in conversion; check for infinity
			if math.IsInf(f, 0) {
				// Return as string representation if too large for float64
				return bf.Text('f', -1)
			}
		}
		// Check if it's an integer
		if f == float64(int64(f)) {
			return int64(f)
		}
		return f
	case val.Type() == cty.Bool:
		return val.True()
	case val.Type().IsListType() || val.Type().IsTupleType():
		var result []any
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			result = append(result, ctyToGo(v))
		}
		return result
	case val.Type().IsMapType() || val.Type().IsObjectType():
		result := make(map[string]any)
		for it := val.ElementIterator(); it.Next(); {
			k, v := it.Element()
			result[k.AsString()] = ctyToGo(v)
		}
		return result
	default:
		return val.GoString()
	}
}

// ctyToStringSlice converts a cty.Value list to []string
func ctyToStringSlice(val cty.Value) []string {
	if val.IsNull() || !val.CanIterateElements() {
		return nil
	}
	var result []string
	for it := val.ElementIterator(); it.Next(); {
		_, v := it.Element()
		if v.Type() == cty.String {
			result = append(result, v.AsString())
		}
	}
	return result
}

// ctyToStringMap converts a cty.Value map to map[string]string
func ctyToStringMap(val cty.Value) map[string]string {
	if val.IsNull() || !val.CanIterateElements() {
		return nil
	}
	result := make(map[string]string)
	for it := val.ElementIterator(); it.Next(); {
		k, v := it.Element()
		if v.Type() == cty.String {
			result[k.AsString()] = v.AsString()
		}
	}
	return result
}

// parseSuccessActionBlock parses HCL block attributes into a SuccessAction
func parseSuccessActionBlock(block *hclsyntax.Block, action *SuccessAction) error {
	var errs []string
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}
		switch name {
		case "type":
			action.Type = SuccessActionType(val.AsString())
		case "workflowId":
			action.WorkflowId = val.AsString()
		case "stepId":
			action.StepId = val.AsString()
		}
	}
	for _, nestedBlock := range block.Body.Blocks {
		switch nestedBlock.Type {
		case "criterion":
			criterion := &Criterion{}
			if err := parseCriterionBlock(nestedBlock, criterion); err != nil {
				errs = append(errs, fmt.Sprintf("criterion: %s", err.Error()))
				continue
			}
			action.Criteria = append(action.Criteria, criterion)
		case "successAction":
			if len(nestedBlock.Labels) > 0 {
				action.Name = nestedBlock.Labels[0]
			}
			if err := parseSuccessActionBlock(nestedBlock, action); err != nil {
				errs = append(errs, fmt.Sprintf("successAction: %s", err.Error()))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("success action errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// parseFailureActionBlock parses HCL block attributes into a FailureAction
func parseFailureActionBlock(block *hclsyntax.Block, action *FailureAction) error {
	var errs []string
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}
		switch name {
		case "type":
			action.Type = FailureActionType(val.AsString())
		case "workflowId":
			action.WorkflowId = val.AsString()
		case "stepId":
			action.StepId = val.AsString()
		case "retryAfter":
			f, acc := val.AsBigFloat().Float64()
			if acc != 0 && math.IsInf(f, 0) {
				errs = append(errs, fmt.Sprintf("retryAfter value too large for float64"))
				continue
			}
			action.RetryAfter = &f
		case "retryLimit":
			f, acc := val.AsBigFloat().Float64()
			if acc != 0 && math.IsInf(f, 0) {
				errs = append(errs, fmt.Sprintf("retryLimit value too large for int"))
				continue
			}
			i := int(f)
			action.RetryLimit = &i
		}
	}
	for _, nestedBlock := range block.Body.Blocks {
		switch nestedBlock.Type {
		case "criterion":
			criterion := &Criterion{}
			if err := parseCriterionBlock(nestedBlock, criterion); err != nil {
				errs = append(errs, fmt.Sprintf("criterion: %s", err.Error()))
				continue
			}
			action.Criteria = append(action.Criteria, criterion)
		case "failureAction":
			if len(nestedBlock.Labels) > 0 {
				action.Name = nestedBlock.Labels[0]
			}
			if err := parseFailureActionBlock(nestedBlock, action); err != nil {
				errs = append(errs, fmt.Sprintf("failureAction: %s", err.Error()))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failure action errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// parseParameterBlock parses HCL block attributes into a Parameter
func parseParameterBlock(block *hclsyntax.Block, param *Parameter) error {
	var errs []string
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}
		switch name {
		case "in":
			param.In = ParameterIn(val.AsString())
		case "value":
			param.Value = ctyToGo(val)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("parameter errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// parseStepBlock parses an HCL step block into a Step struct
func parseStepBlock(block *hclsyntax.Block, s *Step) error {
	var errs []string

	// Set label (stepId) if provided
	if len(block.Labels) > 0 {
		s.StepId = block.Labels[0]
	}

	// Process attributes
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}

		switch name {
		case "description":
			s.Description = val.AsString()
		case "operationId":
			s.OperationId = val.AsString()
		case "operationPath":
			s.OperationPath = val.AsString()
		case "workflowId":
			s.WorkflowId = val.AsString()
		case "outputs":
			s.Outputs = ctyToStringMap(val)
		case "parameters":
			s.Parameters = ctyToParameters(val)
		}
	}

	// Process nested blocks
	for _, nestedBlock := range block.Body.Blocks {
		switch nestedBlock.Type {
		case "requestBody":
			s.RequestBody = &RequestBody{}
			if err := parseRequestBodyBlock(nestedBlock, s.RequestBody); err != nil {
				errs = append(errs, fmt.Sprintf("requestBody: %s", err.Error()))
			}
		case "successCriterion":
			criterion := &Criterion{}
			if err := parseCriterionBlock(nestedBlock, criterion); err != nil {
				errs = append(errs, fmt.Sprintf("successCriterion: %s", err.Error()))
			}
			s.SuccessCriteria = append(s.SuccessCriteria, criterion)
		case "onSuccess":
			action := &SuccessActionOrReusable{}
			if reusable, ok, err := parseReusableFromBlock(nestedBlock); ok {
				if err != nil {
					errs = append(errs, fmt.Sprintf("onSuccess reusable: %s", err.Error()))
					break
				}
				action.Reusable = reusable
			} else {
				action.SuccessAction = &SuccessAction{}
				if len(nestedBlock.Labels) > 0 {
					action.SuccessAction.Name = nestedBlock.Labels[0]
				}
				if err := parseSuccessActionBlock(nestedBlock, action.SuccessAction); err != nil {
					errs = append(errs, fmt.Sprintf("onSuccess: %s", err.Error()))
				}
			}
			s.OnSuccess = append(s.OnSuccess, action)
		case "onFailure":
			action := &FailureActionOrReusable{}
			if reusable, ok, err := parseReusableFromBlock(nestedBlock); ok {
				if err != nil {
					errs = append(errs, fmt.Sprintf("onFailure reusable: %s", err.Error()))
					break
				}
				action.Reusable = reusable
			} else {
				action.FailureAction = &FailureAction{}
				if len(nestedBlock.Labels) > 0 {
					action.FailureAction.Name = nestedBlock.Labels[0]
				}
				if err := parseFailureActionBlock(nestedBlock, action.FailureAction); err != nil {
					errs = append(errs, fmt.Sprintf("onFailure: %s", err.Error()))
				}
			}
			s.OnFailure = append(s.OnFailure, action)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("step %q errors: %s", s.StepId, strings.Join(errs, "; "))
	}
	return nil
}

func parseReusableFromBlock(block *hclsyntax.Block) (*ReusableObject, bool, error) {
	for _, nestedBlock := range block.Body.Blocks {
		if nestedBlock.Type != "reusable" {
			continue
		}
		reusable := &ReusableObject{}
		if err := parseReusableBody(nestedBlock.Body, reusable); err != nil {
			return reusable, true, err
		}
		return reusable, true, nil
	}

	if _, ok := block.Body.Attributes["reference"]; ok {
		reusable := &ReusableObject{}
		if err := parseReusableBody(block.Body, reusable); err != nil {
			return reusable, true, err
		}
		return reusable, true, nil
	}

	return nil, false, nil
}

func parseReusableBody(body *hclsyntax.Body, reusable *ReusableObject) error {
	var errs []string
	for name, attr := range body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}
		switch name {
		case "reference":
			reusable.Reference = val.AsString()
		case "value":
			reusable.Value = ctyToGo(val)
		}
	}
	for _, nestedBlock := range body.Blocks {
		if nestedBlock.Type != "value" {
			continue
		}
		reusable.Value = hclBlockToMap(nestedBlock)
	}
	if len(errs) > 0 {
		return fmt.Errorf("reusable errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// parseRequestBodyBlock parses HCL block into RequestBody
func parseRequestBodyBlock(block *hclsyntax.Block, rb *RequestBody) error {
	var errs []string
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}
		switch name {
		case "contentType":
			rb.ContentType = val.AsString()
		case "payload":
			rb.Payload = ctyToGo(val)
		}
	}
	// Handle payload block
	for _, nestedBlock := range block.Body.Blocks {
		switch nestedBlock.Type {
		case "payload":
			rb.Payload = hclBlockToMap(nestedBlock)
		case "replacement":
			replacement := &PayloadReplacement{}
			if err := parsePayloadReplacementBlock(nestedBlock, replacement); err != nil {
				errs = append(errs, fmt.Sprintf("replacement: %s", err.Error()))
				continue
			}
			rb.Replacements = append(rb.Replacements, replacement)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("request body errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func parsePayloadReplacementBlock(block *hclsyntax.Block, replacement *PayloadReplacement) error {
	var errs []string
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}
		switch name {
		case "target":
			replacement.Target = val.AsString()
		case "value":
			replacement.Value = val.AsString()
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("replacement errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// parseCriterionBlock parses HCL block into Criterion
func parseCriterionBlock(block *hclsyntax.Block, c *Criterion) error {
	var errs []string
	var version string
	var exprType *CriterionExpressionType
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}
		switch name {
		case "context":
			c.Context = val.AsString()
		case "condition":
			c.Condition = val.AsString()
		case "type":
			c.Type = CriterionType(val.AsString())
		case "version":
			version = val.AsString()
		}
	}
	for _, nestedBlock := range block.Body.Blocks {
		if nestedBlock.Type != "expressionType" {
			continue
		}
		exprType = &CriterionExpressionType{}
		if err := parseCriterionExpressionTypeBlock(nestedBlock, exprType); err != nil {
			errs = append(errs, fmt.Sprintf("expressionType: %s", err.Error()))
		}
	}
	if version != "" {
		if exprType != nil {
			errs = append(errs, "expressionType block and version attribute both set")
		} else if c.Type == "" {
			errs = append(errs, "version requires type")
		} else {
			c.ExpressionType = &CriterionExpressionType{
				Type:    c.Type,
				Version: version,
			}
			c.Type = ""
		}
	}
	if exprType != nil {
		c.ExpressionType = exprType
		c.Type = ""
	}
	if len(errs) > 0 {
		return fmt.Errorf("criterion errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func parseCriterionExpressionTypeBlock(block *hclsyntax.Block, expr *CriterionExpressionType) error {
	var errs []string
	for name, attr := range block.Body.Attributes {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			errs = append(errs, fmt.Sprintf("attribute %q: %s", name, diags.Error()))
			continue
		}
		switch name {
		case "type":
			expr.Type = CriterionType(val.AsString())
		case "version":
			expr.Version = val.AsString()
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("expressionType errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// ctyToParameters converts a cty.Value to []any for parameters
func ctyToParameters(val cty.Value) []any {
	if val.IsNull() || !val.CanIterateElements() {
		return nil
	}
	var result []any
	for it := val.ElementIterator(); it.Next(); {
		_, v := it.Element()
		result = append(result, ctyToGo(v))
	}
	return result
}

// SuccessActionOrReusable represents either a SuccessAction or a ReusableObject.
type SuccessActionOrReusable struct {
	SuccessAction *SuccessAction  `hcl:"successAction,block"`
	Reusable      *ReusableObject `hcl:"reusable,block"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *SuccessActionOrReusable) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as ReusableObject first (check for "reference" field)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, hasReference := raw["reference"]; hasReference {
		s.Reusable = &ReusableObject{}
		return json.Unmarshal(data, s.Reusable)
	}

	s.SuccessAction = &SuccessAction{}
	return json.Unmarshal(data, s.SuccessAction)
}

// MarshalJSON implements the json.Marshaler interface.
func (s SuccessActionOrReusable) MarshalJSON() ([]byte, error) {
	if s.Reusable != nil {
		return json.Marshal(s.Reusable)
	}
	return json.Marshal(s.SuccessAction)
}

// FailureActionOrReusable represents either a FailureAction or a ReusableObject.
type FailureActionOrReusable struct {
	FailureAction *FailureAction  `hcl:"failureAction,block"`
	Reusable      *ReusableObject `hcl:"reusable,block"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *FailureActionOrReusable) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as ReusableObject first (check for "reference" field)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, hasReference := raw["reference"]; hasReference {
		f.Reusable = &ReusableObject{}
		return json.Unmarshal(data, f.Reusable)
	}

	f.FailureAction = &FailureAction{}
	return json.Unmarshal(data, f.FailureAction)
}

// MarshalJSON implements the json.Marshaler interface.
func (f FailureActionOrReusable) MarshalJSON() ([]byte, error) {
	if f.Reusable != nil {
		return json.Marshal(f.Reusable)
	}
	return json.Marshal(f.FailureAction)
}

// ParameterOrReusable represents either a Parameter or a ReusableObject.
type ParameterOrReusable struct {
	Parameter *Parameter      `hcl:"parameter,block"`
	Reusable  *ReusableObject `hcl:"reusable,block"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *ParameterOrReusable) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as ReusableObject first (check for "reference" field)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, hasReference := raw["reference"]; hasReference {
		p.Reusable = &ReusableObject{}
		return json.Unmarshal(data, p.Reusable)
	}

	p.Parameter = &Parameter{}
	return json.Unmarshal(data, p.Parameter)
}

// MarshalJSON implements the json.Marshaler interface.
func (p ParameterOrReusable) MarshalJSON() ([]byte, error) {
	if p.Reusable != nil {
		return json.Marshal(p.Reusable)
	}
	return json.Marshal(p.Parameter)
}
