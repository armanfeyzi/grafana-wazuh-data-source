package models

import "strings"

// SanitizeErrorMessage replaces any occurrence of known secret values in msg
// with [REDACTED] to prevent credentials from appearing in user-visible panel
// errors. It is safe to call with a nil secrets pointer.
func SanitizeErrorMessage(msg string, secrets *SecretPluginSettings) string {
	if secrets == nil {
		return msg
	}
	if secrets.Password != "" {
		msg = strings.ReplaceAll(msg, secrets.Password, "[REDACTED]")
	}
	if secrets.IndexerPassword != "" {
		msg = strings.ReplaceAll(msg, secrets.IndexerPassword, "[REDACTED]")
	}
	return msg
}
