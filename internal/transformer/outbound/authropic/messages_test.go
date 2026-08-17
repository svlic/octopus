package authropic

import "testing"

func TestGetThinkingBudget_whenUsingHighestPresets(t *testing.T) {
	// Given
	levels := []string{"xhigh", "max", "ultra"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			// When
			budget := getThinkingBudget(level, nil)

			// Then
			if budget == nil || *budget != 32768 {
				t.Fatalf("expected highest budget for %q, got %v", level, budget)
			}
		})
	}
}
