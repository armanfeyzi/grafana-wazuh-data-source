package models

import "testing"

func TestSanitizeStringList(t *testing.T) {
	t.Parallel()

	if got := SanitizeStringList([]string{"$__all", "", "fedora"}); len(got) != 1 || got[0] != "fedora" {
		t.Fatalf("unexpected sanitize result: %#v", got)
	}

	if SanitizeStringList([]string{"$__all"}) != nil {
		t.Fatal("expected nil for all-only values")
	}

	if got := SanitizeStringList([]string{"$agent"}); got != nil {
		t.Fatalf("expected nil for unresolved template variable, got %#v", got)
	}

	if got := SanitizeStringList([]string{"fedora,wazuh.manager"}); len(got) != 2 || got[0] != "fedora" || got[1] != "wazuh.manager" {
		t.Fatalf("unexpected comma-separated result: %#v", got)
	}
}
