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
}
