package domain

import "testing"

func TestPredefinedPreferenceKeys(t *testing.T) {
	if len(PredefinedPreferenceKeys) == 0 {
		t.Fatal("PredefinedPreferenceKeys is empty")
	}

	seen := make(map[string]bool)
	for _, key := range PredefinedPreferenceKeys {
		if key == "" {
			t.Error("PredefinedPreferenceKeys contains an empty key")
		}
		if seen[key] {
			t.Errorf("PredefinedPreferenceKeys contains duplicate key %q", key)
		}
		seen[key] = true
	}
}
