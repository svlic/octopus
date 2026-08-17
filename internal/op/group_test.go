package op

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bestruirui/octopus/internal/db"
	"github.com/bestruirui/octopus/internal/model"
)

func TestGroupUpdate_whenThinkingLevelIsOmitted(t *testing.T) {
	// Given
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	group := &model.Group{
		Name: "legacy-update-group",
		Mode: model.GroupModeWeighted,
		Items: []model.GroupItem{{
			ChannelID:     1,
			ModelName:     "reasoning-model",
			Priority:      1,
			Weight:        1,
			ThinkingLevel: "high",
		}},
	}
	if err := GroupCreate(group, context.Background()); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}

	// When
	priority := 2
	weight := 3
	updated, err := GroupUpdate(&model.GroupUpdateRequest{
		ID: group.ID,
		ItemsToUpdate: []model.GroupItemUpdateRequest{{
			ID:       group.Items[0].ID,
			Priority: &priority,
			Weight:   &weight,
		}},
	}, context.Background())

	// Then
	if err != nil {
		t.Fatalf("GroupUpdate() error = %v", err)
	}
	if got := updated.Items[0].ThinkingLevel; got != "high" {
		t.Fatalf("thinking level = %q, want %q", got, "high")
	}
}

func TestGroupUpdate_whenOnlyThinkingLevelChanges(t *testing.T) {
	// Given
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	group := &model.Group{
		Name: "thinking-only-update-group",
		Mode: model.GroupModeWeighted,
		Items: []model.GroupItem{{
			ChannelID:     1,
			ModelName:     "reasoning-model",
			Priority:      4,
			Weight:        7,
			ThinkingLevel: "low",
		}},
	}
	if err := GroupCreate(group, context.Background()); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}

	// When
	thinkingLevel := "high"
	updated, err := GroupUpdate(&model.GroupUpdateRequest{
		ID: group.ID,
		ItemsToUpdate: []model.GroupItemUpdateRequest{{
			ID:            group.Items[0].ID,
			ThinkingLevel: &thinkingLevel,
		}},
	}, context.Background())

	// Then
	if err != nil {
		t.Fatalf("GroupUpdate() error = %v", err)
	}
	item := updated.Items[0]
	if item.Priority != 4 || item.Weight != 7 || item.ThinkingLevel != "high" {
		t.Fatalf("updated item = %+v, want priority=4 weight=7 thinking_level=high", item)
	}
}

func TestGroupUpdate_whenThinkingLevelsChangeAndItemsAreAdded(t *testing.T) {
	// Given
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	group := &model.Group{
		Name: "reasoning-group",
		Mode: model.GroupModeFailover,
		Items: []model.GroupItem{{
			ChannelID:     1,
			ModelName:     "first-model",
			Priority:      1,
			Weight:        1,
			ThinkingLevel: "low",
		}},
	}
	if err := GroupCreate(group, context.Background()); err != nil {
		t.Fatalf("GroupCreate() error = %v", err)
	}

	// When
	priority := 1
	weight := 1
	thinkingLevel := "high"
	updated, err := GroupUpdate(&model.GroupUpdateRequest{
		ID: group.ID,
		ItemsToUpdate: []model.GroupItemUpdateRequest{{
			ID:            group.Items[0].ID,
			Priority:      &priority,
			Weight:        &weight,
			ThinkingLevel: &thinkingLevel,
		}},
		ItemsToAdd: []model.GroupItemAddRequest{{
			ChannelID:     2,
			ModelName:     "second-model",
			Priority:      2,
			Weight:        1,
			ThinkingLevel: "max",
		}},
	}, context.Background())

	// Then
	if err != nil {
		t.Fatalf("GroupUpdate() error = %v", err)
	}
	levels := make(map[string]string, len(updated.Items))
	for _, item := range updated.Items {
		levels[item.ModelName] = item.ThinkingLevel
	}
	if got := levels["first-model"]; got != "high" {
		t.Fatalf("updated thinking level = %q, want %q", got, "high")
	}
	if got := levels["second-model"]; got != "max" {
		t.Fatalf("added thinking level = %q, want %q", got, "max")
	}
}
