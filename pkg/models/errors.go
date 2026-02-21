package models

import (
	"errors"
	"fmt"
)

// ErrorCode classifies the kind of failure returned by the plugin backends.
type ErrorCode string

const (
	// ErrAuth is returned when credentials are rejected (HTTP 401).
	ErrAuth ErrorCode = "auth"

	// ErrForbidden is returned when the user lacks RBAC permissions (HTTP 403).
	ErrForbidden ErrorCode = "forbidden"

	// ErrUnreachable is returned when the upstream service is not contactable.
	ErrUnreachable ErrorCode = "unreachable"

	// ErrTimeout is returned when a context deadline is exceeded.
	ErrTimeout ErrorCode = "timeout"

	// ErrIndexMissing is returned when the target OpenSearch index does not
	// exist (HTTP 404 on a _search request).
	ErrIndexMissing ErrorCode = "index_missing"

	// ErrBadResponse is returned when the upstream response cannot be parsed.
	ErrBadResponse ErrorCode = "bad_response"
)

// WazuhError is a structured error that carries an ErrorCode so the query
// handler can return an appropriate Grafana HTTP status and user-readable
// message without exposing internal details.
type WazuhError struct {
	Code    ErrorCode
	Message string // user-facing; must not contain credentials
	Cause   error  // internal; logged but never surfaced to users
}

func (e *WazuhError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap allows errors.Is / errors.As to traverse the cause chain.
func (e *WazuhError) Unwrap() error {
	return e.Cause
}

// NewWazuhError constructs a WazuhError.
func NewWazuhError(code ErrorCode, message string, cause error) *WazuhError {
	return &WazuhError{Code: code, Message: message, Cause: cause}
}

// IsWazuhError reports whether any error in the chain is a WazuhError with
// the given code.
func IsWazuhError(err error, code ErrorCode) bool {
	var we *WazuhError
	if errors.As(err, &we) {
		return we.Code == code
	}
	return false
}

// AsWazuhError extracts the first WazuhError from the chain, if any.
func AsWazuhError(err error) (*WazuhError, bool) {
	var we *WazuhError
	return we, errors.As(err, &we)
}

// UserMessage returns the WazuhError.Message if err is (or wraps) a
// WazuhError, otherwise falls back to err.Error().
func UserMessage(err error) string {
	if we, ok := AsWazuhError(err); ok {
		return we.Message
	}
	return err.Error()
}
