package gemini

import "testing"

func TestReasoningToThinkingBudget_whenUsingHighestPresets(t *testing.T) {
	// Given
	levels := []string{"xhigh", "max", "ultra"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			// When
			budget := reasoningToThinkingBudget(level)

			// Then
			if budget != 24576 {
				t.Fatalf("expected highest budget for %q, got %d", level, budget)
			}
		})
	}
}
