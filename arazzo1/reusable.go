package arazzo1

// ReusableObject is a simple object to allow referencing of objects
// contained within the Components Object.
type ReusableObject struct {
	// Reference is a runtime expression used to reference the desired object (required).
	Reference string `json:"reference" yaml:"reference" hcl:"reference"`

	// Value sets a value of the referenced parameter.
	// Can be string, boolean, object, array, number, or null.
	Value any `json:"value,omitempty" yaml:"value,omitempty" hcl:"value,optional"`
}
