package metav1

// TODO: proof of concept

type StringFilter struct {
	// Is apply if value is exact equal.
	Is string `json:"is,omitempty"`

	// Match apply with regular expression.
	Match string `json:"match,omitempty"`
}

type UintFilter struct {
	// Is apply if value is exact equal.
	Is uint `json:"is,omitempty"`
}
