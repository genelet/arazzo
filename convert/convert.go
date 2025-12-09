// Package convert provides functions to convert Arazzo documents between JSON and HCL formats.
package convert

import (
	"encoding/json"
	"strings"

	"github.com/genelet/arazzo/arazzo1"
	"github.com/genelet/horizon/dethcl"
)

// transformKeys recursively transforms map keys in a value.
// When toHCL is true, it converts $ref to _ref (for HCL compatibility).
// When toHCL is false, it converts _ref back to $ref (for JSON compatibility).
func transformKeys(v any, toHCL bool) any {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any)
		for k, v := range val {
			newKey := k
			if toHCL && strings.HasPrefix(k, "$") {
				newKey = "_" + k[1:]
			} else if !toHCL && strings.HasPrefix(k, "_") {
				// Check if this looks like a transformed $ref key
				// Common JSON Schema keys that start with $: $ref, $id, $schema, $defs, $comment, $vocabulary
				switch k {
				case "_ref", "_id", "_schema", "_defs", "_comment", "_vocabulary", "_anchor", "_dynamicRef", "_dynamicAnchor":
					newKey = "$" + k[1:]
				}
			}
			result[newKey] = transformKeys(v, toHCL)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = transformKeys(item, toHCL)
		}
		return result
	default:
		return v
	}
}

// transformArazzoForHCL transforms an Arazzo document's dynamic fields ($ref -> _ref) for HCL compatibility.
func transformArazzoForHCL(doc *arazzo1.Arazzo) {
	// Transform workflow inputs
	for _, wf := range doc.Workflows {
		if wf.Inputs != nil {
			wf.Inputs = transformKeys(wf.Inputs, true)
		}
	}
	// Transform component inputs
	if doc.Components != nil && doc.Components.Inputs != nil {
		for k, v := range doc.Components.Inputs {
			doc.Components.Inputs[k] = transformKeys(v, true)
		}
	}
}

// transformArazzoFromHCL transforms an Arazzo document's dynamic fields (_ref -> $ref) back from HCL.
func transformArazzoFromHCL(doc *arazzo1.Arazzo) {
	// Transform workflow inputs
	for _, wf := range doc.Workflows {
		if wf.Inputs != nil {
			wf.Inputs = transformKeys(wf.Inputs, false)
		}
	}
	// Transform component inputs
	if doc.Components != nil && doc.Components.Inputs != nil {
		for k, v := range doc.Components.Inputs {
			doc.Components.Inputs[k] = transformKeys(v, false)
		}
	}
}

// JSONToHCL converts an Arazzo document from JSON format to HCL format.
// It first unmarshals the JSON into an Arazzo struct, then marshals it to HCL.
// JSON Schema keys like $ref are transformed to _ref for HCL compatibility.
func JSONToHCL(jsonData []byte) ([]byte, error) {
	var doc arazzo1.Arazzo
	if err := json.Unmarshal(jsonData, &doc); err != nil {
		return nil, err
	}
	transformArazzoForHCL(&doc)
	return dethcl.Marshal(&doc)
}

// HCLToJSON converts an Arazzo document from HCL format to JSON format.
// It first unmarshals the HCL into an Arazzo struct, then marshals it to JSON.
// HCL keys like _ref are transformed back to $ref for JSON compatibility.
func HCLToJSON(hclData []byte) ([]byte, error) {
	var doc arazzo1.Arazzo
	if err := dethcl.Unmarshal(hclData, &doc); err != nil {
		return nil, err
	}
	transformArazzoFromHCL(&doc)
	return json.Marshal(&doc)
}

// HCLToJSONIndent converts an Arazzo document from HCL format to indented JSON format.
// HCL keys like _ref are transformed back to $ref for JSON compatibility.
func HCLToJSONIndent(hclData []byte, prefix, indent string) ([]byte, error) {
	var doc arazzo1.Arazzo
	if err := dethcl.Unmarshal(hclData, &doc); err != nil {
		return nil, err
	}
	transformArazzoFromHCL(&doc)
	return json.MarshalIndent(&doc, prefix, indent)
}

// MarshalHCL marshals an Arazzo document to HCL format.
// JSON Schema keys like $ref are transformed to _ref for HCL compatibility.
// Note: This function modifies the document in place. If you need to preserve
// the original, make a copy before calling this function.
func MarshalHCL(doc *arazzo1.Arazzo) ([]byte, error) {
	transformArazzoForHCL(doc)
	return dethcl.Marshal(doc)
}

// UnmarshalHCL unmarshals HCL data into an Arazzo document.
// HCL keys like _ref are transformed back to $ref for JSON compatibility.
func UnmarshalHCL(hclData []byte, doc *arazzo1.Arazzo) error {
	if err := dethcl.Unmarshal(hclData, doc); err != nil {
		return err
	}
	transformArazzoFromHCL(doc)
	return nil
}

// MarshalJSON marshals an Arazzo document to JSON format.
func MarshalJSON(doc *arazzo1.Arazzo) ([]byte, error) {
	return json.Marshal(doc)
}

// MarshalJSONIndent marshals an Arazzo document to indented JSON format.
func MarshalJSONIndent(doc *arazzo1.Arazzo, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(doc, prefix, indent)
}

// UnmarshalJSON unmarshals JSON data into an Arazzo document.
func UnmarshalJSON(jsonData []byte, doc *arazzo1.Arazzo) error {
	return json.Unmarshal(jsonData, doc)
}
