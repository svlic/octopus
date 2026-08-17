package volcengine

import "testing"

func TestThinkingTypeForReasoningEffort_whenUsingHighestPresets(t *testing.T) {
	// Given
	levels := []string{"xhigh", "max", "ultra"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			// When
			thinkingType := thinkingTypeForReasoningEffort(level)

			// Then
			if thinkingType != ThinkingTypeEnabled {
				t.Fatalf("expected thinking enabled for %q, got %q", level, thinkingType)
			}
		})
	}
}
