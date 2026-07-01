package models

import (
	"errors"
	"fmt"
	"testing"
)

// --- WazuhError ---

func TestWazuhError_Error_withCause(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	we := NewWazuhError(ErrUnreachable, "cannot reach indexer", cause)

	if got := we.Error(); got != "cannot reach indexer" {
		t.Errorf("expected %q, got %q", "cannot reach indexer", got)
	}
}

func TestWazuhError_Error_doesNotExposeCause(t *testing.T) {
	cause := fmt.Errorf("dial tcp 192.168.1.10:9200: connection refused")
	we := NewWazuhError(ErrUnreachable, "cannot reach indexer", cause)

	if got := we.Error(); contains(got, "192.168") {
		t.Errorf("Error() must not expose cause details, got %q", got)
	}
}

func TestWazuhError_Error_withoutCause(t *testing.T) {
	we := NewWazuhError(ErrAuth, "authentication failed", nil)
	if we.Error() != "authentication failed" {
		t.Errorf("expected %q, got %q", "authentication failed", we.Error())
	}
}

func TestWazuhError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	we := NewWazuhError(ErrTimeout, "timed out", cause)

	if !errors.Is(we, cause) {
		t.Error("errors.Is should find the wrapped cause")
	}
}

func TestIsWazuhError(t *testing.T) {
	we := NewWazuhError(ErrForbidden, "permission denied", nil)

	if !IsWazuhError(we, ErrForbidden) {
		t.Error("expected IsWazuhError(ErrForbidden) = true")
	}
	if IsWazuhError(we, ErrAuth) {
		t.Error("expected IsWazuhError(ErrAuth) = false for ErrForbidden error")
	}
}

func TestIsWazuhError_wrapped(t *testing.T) {
	we := NewWazuhError(ErrIndexMissing, "index not found", nil)
	wrapped := fmt.Errorf("outer: %w", we)

	if !IsWazuhError(wrapped, ErrIndexMissing) {
		t.Error("expected IsWazuhError to unwrap through fmt.Errorf")
	}
}

func TestIsWazuhError_nil(t *testing.T) {
	if IsWazuhError(nil, ErrAuth) {
		t.Error("expected false for nil error")
	}
}

func TestAsWazuhError_found(t *testing.T) {
	we := NewWazuhError(ErrBadResponse, "bad response", nil)
	got, ok := AsWazuhError(we)
	if !ok {
		t.Fatal("expected AsWazuhError to find the error")
	}
	if got.Code != ErrBadResponse {
		t.Errorf("expected code ErrBadResponse, got %q", got.Code)
	}
}

func TestAsWazuhError_notFound(t *testing.T) {
	_, ok := AsWazuhError(fmt.Errorf("plain error"))
	if ok {
		t.Error("expected AsWazuhError to return false for plain error")
	}
}

func TestUserMessage_wazuhError(t *testing.T) {
	we := NewWazuhError(ErrAuth, "user-friendly message", fmt.Errorf("raw internal cause"))
	if got := UserMessage(we); got != "user-friendly message" {
		t.Errorf("expected %q, got %q", "user-friendly message", got)
	}
}

func TestUserMessage_plainError(t *testing.T) {
	err := fmt.Errorf("dial tcp 10.0.0.5:443")
	if got := UserMessage(err); got != "An unexpected error occurred" {
		t.Errorf("expected %q, got %q", "An unexpected error occurred", got)
	}
}

// --- SanitizeErrorMessage ---

func TestSanitizeErrorMessage_replacesPassword(t *testing.T) {
	secrets := &SecretPluginSettings{Password: "s3cr3t"}
	msg := SanitizeErrorMessage("connection error: auth failed for s3cr3t", secrets)
	if contains(msg, "s3cr3t") {
		t.Errorf("expected password to be redacted, got: %q", msg)
	}
	if !contains(msg, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in message, got: %q", msg)
	}
}

func TestSanitizeErrorMessage_replacesIndexerPassword(t *testing.T) {
	secrets := &SecretPluginSettings{IndexerPassword: "idx-pass"}
	msg := SanitizeErrorMessage("indexer said: bad pass idx-pass rejected", secrets)
	if contains(msg, "idx-pass") {
		t.Errorf("expected indexer password to be redacted, got: %q", msg)
	}
}

func TestSanitizeErrorMessage_nilSecrets(t *testing.T) {
	msg := SanitizeErrorMessage("no secrets here", nil)
	if msg != "no secrets here" {
		t.Errorf("expected message unchanged, got: %q", msg)
	}
}

func TestSanitizeErrorMessage_emptyPassword(t *testing.T) {
	secrets := &SecretPluginSettings{Password: ""}
	msg := SanitizeErrorMessage("message", secrets)
	if msg != "message" {
		t.Errorf("expected message unchanged with empty password, got: %q", msg)
	}
}

// --- helpers ---

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// Ensure the testing.T is used (suppress unused import warning).
var _ *testing.T
