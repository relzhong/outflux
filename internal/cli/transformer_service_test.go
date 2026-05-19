package cli

import "testing"

func TestFilterSkippedNames(t *testing.T) {
	filtered := filterSkippedNames([]string{"a", "b", "c"}, []string{"b", "missing"})
	if len(filtered) != 2 || filtered[0] != "a" || filtered[1] != "c" {
		t.Fatalf("unexpected filtered names: %v", filtered)
	}
}
