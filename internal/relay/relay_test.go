package relay

import (
	"testing"

	"github.com/bestruirui/octopus/internal/transformer/model"
)

func TestResolveReasoningEffort_whenGroupItemOverridesClientValue(t *testing.T) {
	// Given
	const original = "medium"

	// When
	got := resolveReasoningEffort(original, "high")

	// Then
	if got != "high" {
		t.Fatalf("resolveReasoningEffort() = %q, want %q", got, "high")
	}
}

func TestApplyThinkingLevelOverride_whenClientProvidesReasoningBudget(t *testing.T) {
	// Given
	budget := int64(30000)
	request := &model.InternalLLMRequest{
		ReasoningEffort:  "high",
		ReasoningBudget:  &budget,
		AdaptiveThinking: true,
	}

	// When
	applyThinkingLevelOverride(request, reasoningState{
		effort:           request.ReasoningEffort,
		budget:           request.ReasoningBudget,
		adaptiveThinking: request.AdaptiveThinking,
	}, "low")

	// Then
	if request.ReasoningEffort != "low" || request.ReasoningBudget != nil || request.AdaptiveThinking {
		t.Fatalf("request reasoning state = effort:%q budget:%v adaptive:%t", request.ReasoningEffort, request.ReasoningBudget, request.AdaptiveThinking)
	}
}

func TestResolveReasoningEffort_whenRetryInheritsClientValue(t *testing.T) {
	// Given
	const original = "medium"
	firstAttempt := resolveReasoningEffort(original, "high")
	if firstAttempt != "high" {
		t.Fatalf("first attempt reasoning effort = %q, want %q", firstAttempt, "high")
	}

	// When
	got := resolveReasoningEffort(original, "")

	// Then
	if got != original {
		t.Fatalf("retry reasoning effort = %q, want original %q", got, original)
	}
}

func TestApplyThinkingLevelOverride_whenLevelIsDefault(t *testing.T) {
	// Given
	budget := int64(30000)
	request := &model.InternalLLMRequest{}
	original := reasoningState{
		effort:           "medium",
		budget:           &budget,
		adaptiveThinking: true,
	}

	// When
	applyThinkingLevelOverride(request, original, "default")

	// Then
	if request.ReasoningEffort != original.effort || request.ReasoningBudget != original.budget || request.AdaptiveThinking != original.adaptiveThinking {
		t.Fatalf("request reasoning state = effort:%q budget:%v adaptive:%t", request.ReasoningEffort, request.ReasoningBudget, request.AdaptiveThinking)
	}
}
