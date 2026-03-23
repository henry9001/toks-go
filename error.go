package toks

import "fmt"

// APIError represents an error response from the TOKS API.
type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"error"`
	Reason     string `json:"reason,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("toks: %d %s", e.StatusCode, e.Code)
	if e.Reason != "" {
		msg += ": " + e.Reason
	}
	if e.Detail != "" {
		msg += " (" + e.Detail + ")"
	}
	return msg
}

// IsNotFound reports whether the error is a 404 Not Found.
func IsNotFound(err error) bool {
	e, ok := err.(*APIError)
	return ok && e.StatusCode == 404
}

// IsUnauthorized reports whether the error is a 401 Unauthorized.
func IsUnauthorized(err error) bool {
	e, ok := err.(*APIError)
	return ok && e.StatusCode == 401
}

// IsForbidden reports whether the error is a 403 Forbidden.
func IsForbidden(err error) bool {
	e, ok := err.(*APIError)
	return ok && e.StatusCode == 403
}

// IsKilled reports whether the error indicates an active kill switch.
func IsKilled(err error) bool {
	e, ok := err.(*APIError)
	return ok && e.Code == "killed"
}
