package service

import "strings"

// ValidationError carries field-level validation messages.
type ValidationError struct {
	Fields map[string]string // field name -> error message
	Msg    string            // general message shown above the form
}

func (e *ValidationError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	msgs := make([]string, 0, len(e.Fields))
	for _, v := range e.Fields {
		msgs = append(msgs, v)
	}
	return strings.Join(msgs, "; ")
}

// Field returns the error message for a specific field, or an empty string.
func (e *ValidationError) Field(name string) string {
	if e == nil {
		return ""
	}
	return e.Fields[name]
}
