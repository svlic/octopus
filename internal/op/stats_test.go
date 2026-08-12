package op

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bestruirui/octopus/internal/db"
	"github.com/bestruirui/octopus/internal/model"
)

func TestStatsModelListAggregatesTokensAcrossChannels(t *testing.T) {
	// Given
	resetStatsModelCache()
	t.Cleanup(resetStatsModelCache)
	updates := []model.StatsModel{
		{Name: "gpt-4o", ChannelID: 1, StatsMetrics: model.StatsMetrics{InputToken: 100, OutputToken: 50}},
		{Name: "gpt-4o", ChannelID: 2, StatsMetrics: model.StatsMetrics{InputToken: 25, OutputToken: 75}},
		{Name: "claude-3-7", ChannelID: 1, StatsMetrics: model.StatsMetrics{InputToken: 40, OutputToken: 10}},
	}
	for _, update := range updates {
		if err := StatsModelUpdate(update); err != nil {
			t.Fatalf("StatsModelUpdate() error = %v", err)
		}
	}

	// When
	stats := StatsModelList()

	// Then
	byName := make(map[string]model.StatsModelRanking, len(stats))
	for _, item := range stats {
		byName[item.Name] = item
	}
	if got := byName["gpt-4o"].InputToken + byName["gpt-4o"].OutputToken; got != 250 {
		t.Fatalf("gpt-4o total tokens = %d, want 250", got)
	}
	if got := byName["claude-3-7"].InputToken + byName["claude-3-7"].OutputToken; got != 50 {
		t.Fatalf("claude-3-7 total tokens = %d, want 50", got)
	}
}

func TestStatsModelPersistsAcrossCacheRefresh(t *testing.T) {
	// Given
	resetStatsModelCache()
	t.Cleanup(resetStatsModelCache)
	path := filepath.Join(t.TempDir(), "stats.db")
	if err := db.InitDB("sqlite", path, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	if err := StatsModelUpdate(model.StatsModel{
		Name: "gpt-4o", ChannelID: 1,
		StatsMetrics: model.StatsMetrics{InputToken: 100, OutputToken: 50},
	}); err != nil {
		t.Fatalf("StatsModelUpdate() error = %v", err)
	}

	// When
	if err := StatsSaveDB(context.Background()); err != nil {
		t.Fatalf("StatsSaveDB() error = %v", err)
	}
	resetStatsModelCache()
	if err := statsRefreshCache(context.Background()); err != nil {
		t.Fatalf("statsRefreshCache() error = %v", err)
	}

	// Then
	stats := StatsModelList()
	if len(stats) != 1 || stats[0].Name != "gpt-4o" || stats[0].InputToken+stats[0].OutputToken != 150 {
		t.Fatalf("StatsModelList() = %+v, want persisted gpt-4o with 150 tokens", stats)
	}
}

func TestStatsModelRepeatedSavesUpdateSameRow(t *testing.T) {
	// Given
	resetStatsModelCache()
	t.Cleanup(resetStatsModelCache)
	path := filepath.Join(t.TempDir(), "stats.db")
	if err := db.InitDB("sqlite", path, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	update := model.StatsModel{
		Name: "gpt-4o", ChannelID: 1,
		StatsMetrics: model.StatsMetrics{InputToken: 100},
	}
	if err := StatsModelUpdate(update); err != nil {
		t.Fatalf("StatsModelUpdate() error = %v", err)
	}
	if err := StatsSaveDB(context.Background()); err != nil {
		t.Fatalf("first StatsSaveDB() error = %v", err)
	}

	// When
	update.StatsMetrics = model.StatsMetrics{OutputToken: 50}
	if err := StatsModelUpdate(update); err != nil {
		t.Fatalf("StatsModelUpdate() error = %v", err)
	}
	if err := StatsSaveDB(context.Background()); err != nil {
		t.Fatalf("second StatsSaveDB() error = %v", err)
	}

	// Then
	var rows []model.StatsModel
	if err := db.GetDB().Find(&rows).Error; err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(rows) != 1 || rows[0].InputToken != 100 || rows[0].OutputToken != 50 {
		t.Fatalf("model rows = %+v, want one row with 150 total tokens", rows)
	}
}

func TestStatsModelBackupIncludesChannelID(t *testing.T) {
	// Given
	stats := model.StatsModel{ID: 1, Name: "gpt-4o", ChannelID: 7}

	// When
	encoded, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Then
	if !strings.Contains(string(encoded), `"channel_id":7`) {
		t.Fatalf("StatsModel JSON = %s, want channel_id", encoded)
	}
}

func resetStatsModelCache() {
	statsModelCache.Clear()
	statsModelCacheNeedUpdateLock.Lock()
	statsModelCacheNeedUpdate = make(map[statsModelKey]struct{})
	statsModelCacheNeedUpdateLock.Unlock()
}
