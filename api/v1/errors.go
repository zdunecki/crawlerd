package v1

import "errors"

var (
	ErrNoStorage   = errors.New("no store")
	ErrNoScheduler = errors.New("no scheduler")
)

const ErrorTypeValidation = "Validation"
const ErrorTypeInternal = "Internal"

type APIError struct {
	TraceID string                 `json:"trace_id,omitempty"` // TraceID is unique through the entire path in distributed system
	SpanID  string                 `json:"span_id"`            // SpanID is unique per distributed component
	Type    string                 `json:"type"`               // Type is e.g 'Validation'
	Code    string                 `json:"code"`               // Code e.g 'InvalidEmail'
	Message string                 `json:"message"`            // Message e.g 'Invalid address email'
	Params  map[string]interface{} `json:"params"`             // Params e.g. {"email": "john.doe@example.com"}
}

func (e *APIError) Error() string {
	return e.Message
}
