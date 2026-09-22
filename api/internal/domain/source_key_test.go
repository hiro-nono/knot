package domain

import "testing"

func TestPredefinedSourceKeys(t *testing.T) {
	for sourceType, keys := range PredefinedSourceKeys {
		if !sourceType.valid() {
			t.Errorf("PredefinedSourceKeys has invalid SourceType %q", sourceType)
		}
		if len(keys) == 0 {
			t.Errorf("PredefinedSourceKeys[%q] is empty", sourceType)
		}

		seen := make(map[string]bool)
		for _, key := range keys {
			if key == "" {
				t.Errorf("PredefinedSourceKeys[%q] contains an empty key", sourceType)
			}
			if seen[key] {
				t.Errorf("PredefinedSourceKeys[%q] contains duplicate key %q", sourceType, key)
			}
			seen[key] = true
		}
	}
}
