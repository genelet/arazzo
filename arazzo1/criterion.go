package arazzo1

import (
	"encoding/json"
)

// CriterionType represents the type of condition to be applied.
type CriterionType string

const (
	// CriterionTypeSimple is a simple string matching condition (default).
	CriterionTypeSimple CriterionType = "simple"

	// CriterionTypeRegex is a regular expression condition.
	CriterionTypeRegex CriterionType = "regex"

	// CriterionTypeJSONPath is a JSONPath expression condition.
	CriterionTypeJSONPath CriterionType = "jsonpath"

	// CriterionTypeXPath is an XPath expression condition.
	CriterionTypeXPath CriterionType = "xpath"
)

// Criterion is an object used to specify the context, conditions, and condition types
// that can be used to prove or satisfy assertions specified in Step Object successCriteria,
// Success Action Object criteria, and Failure Action Object criteria.
type Criterion struct {
	// Context is a runtime expression used to set the context for the condition to be applied on.
	Context string `json:"context,omitempty" yaml:"context,omitempty" hcl:"context,optional"`

	// Condition is the condition to apply (required).
	Condition string `json:"condition" yaml:"condition" hcl:"condition"`

	// Type is the type of condition to be applied.
	// Can be "simple", "regex", "jsonpath", or "xpath".
	// For jsonpath and xpath with version, use ExpressionType instead.
	Type CriterionType `json:"type,omitempty" yaml:"type,omitempty" hcl:"type,optional"`

	// ExpressionType contains the type and version for expression-based criteria (jsonpath/xpath).
	// When set, this takes precedence over Type.
	ExpressionType *CriterionExpressionType `json:"-" yaml:"-" hcl:"expressionType,block"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type criterionAlias struct {
	Context   string        `json:"context,omitempty"`
	Condition string        `json:"condition"`
	Type      CriterionType `json:"type,omitempty"`
}

var criterionKnownFields = []string{
	"context",
	"condition",
	"type",
	"version",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *Criterion) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Check if this has both "type" and "version" (criterion-expression-type-object)
	_, hasType := raw["type"]
	_, hasVersion := raw["version"]

	if hasType && hasVersion {
		// This is a criterion with expression type
		c.ExpressionType = &CriterionExpressionType{}
		if err := json.Unmarshal(data, c.ExpressionType); err != nil {
			return err
		}
	}

	// Unmarshal the common fields
	var alias criterionAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	c.Context = alias.Context
	c.Condition = alias.Condition
	c.Type = alias.Type

	c.Extensions = extractExtensions(raw, criterionKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (c Criterion) MarshalJSON() ([]byte, error) {
	// If we have an expression type with version, include it
	if c.ExpressionType != nil && c.ExpressionType.Version != "" {
		type criterionWithExpression struct {
			Context   string        `json:"context,omitempty"`
			Condition string        `json:"condition"`
			Type      CriterionType `json:"type,omitempty"`
			Version   string        `json:"version,omitempty"`
		}
		alias := criterionWithExpression{
			Context:   c.Context,
			Condition: c.Condition,
			Type:      c.ExpressionType.Type,
			Version:   c.ExpressionType.Version,
		}
		return marshalWithExtensions(&alias, c.Extensions)
	}

	alias := criterionAlias{
		Context:   c.Context,
		Condition: c.Condition,
		Type:      c.Type,
	}
	return marshalWithExtensions(&alias, c.Extensions)
}

// CriterionExpressionType is an object used to describe the type and version
// of an expression used within a Criterion Object.
type CriterionExpressionType struct {
	// Type is the type of condition to be applied (required).
	// Must be "jsonpath" or "xpath".
	Type CriterionType `json:"type" yaml:"type" hcl:"type"`

	// Version is a short hand string representing the version of the expression type (required).
	// For jsonpath: "draft-goessner-dispatch-jsonpath-00"
	// For xpath: "xpath-10", "xpath-20", "xpath-30"
	Version string `json:"version" yaml:"version" hcl:"version"`

	// Extensions contains specification extensions (x-*)
	Extensions map[string]any `json:"-" yaml:"-" hcl:"-"`
}

type criterionExpressionTypeAlias CriterionExpressionType

var criterionExpressionTypeKnownFields = []string{
	"type",
	"version",
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *CriterionExpressionType) UnmarshalJSON(data []byte) error {
	var alias criterionExpressionTypeAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*c = CriterionExpressionType(alias)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	c.Extensions = extractExtensions(raw, criterionExpressionTypeKnownFields)

	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (c CriterionExpressionType) MarshalJSON() ([]byte, error) {
	alias := criterionExpressionTypeAlias(c)
	return marshalWithExtensions(&alias, c.Extensions)
}
