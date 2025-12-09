// Package arazzo1 provides Go types for parsing and generating Arazzo 1.0.x documents.
// Created with the help of Claude Code.
package arazzo1

import (
	"encoding/json"
	"strings"
)

// extractExtensions extracts x-* extension fields from a raw JSON map.
// It filters out known fields and returns only extension fields.
func extractExtensions(raw map[string]json.RawMessage, knownFields []string) map[string]any {
	known := make(map[string]bool)
	for _, f := range knownFields {
		known[f] = true
	}

	extensions := make(map[string]any)
	for key, value := range raw {
		if strings.HasPrefix(key, "x-") && !known[key] {
			var v any
			if err := json.Unmarshal(value, &v); err == nil {
				extensions[key] = v
			}
		}
	}

	if len(extensions) == 0 {
		return nil
	}
	return extensions
}

// marshalWithExtensions marshals an object along with its x-* extensions.
func marshalWithExtensions(v any, extensions map[string]any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if len(extensions) == 0 {
		return data, nil
	}

	// Merge extensions into JSON object
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	for key, value := range extensions {
		extData, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		m[key] = extData
	}

	return json.Marshal(m)
}
