package v1

import "regexp"

// TODO: proof of concept

type StringFilter struct {
	// Is apply if value is exact equal.
	Is string `json:"is,omitempty"`

	// Match apply with regular expression.
	Match *regexp.Regexp `json:"match,omitempty"`
}
