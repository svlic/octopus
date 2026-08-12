package op

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bestruirui/octopus/internal/db"
	"github.com/bestruirui/octopus/internal/model"
)

func Test_DBImportIncremental_imports_large_statistics_only_dump(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "backup.db")
	if err := db.InitDB("sqlite", path, false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	const modelCount = 5000
	models := make([]model.StatsModel, modelCount)
	for i := range models {
		models[i] = model.StatsModel{
			Name:      fmt.Sprintf("model-%d", i),
			ChannelID: 1,
			StatsMetrics: model.StatsMetrics{
				RequestSuccess: 1_000_000_000,
			},
		}
	}
	dump := &model.DBDump{
		Version:      dbDumpVersion,
		IncludeStats: true,
		StatsDaily: []model.StatsDaily{{
			Date: "20260812",
			StatsMetrics: model.StatsMetrics{
				RequestSuccess: 1_000_000_000,
			},
		}},
		StatsModel: models,
	}

	// When
	result, err := DBImportIncremental(context.Background(), dump)

	// Then
	if err != nil {
		t.Fatalf("DBImportIncremental() error = %v", err)
	}
	if got := result.RowsAffected["stats_model"]; got != modelCount {
		t.Fatalf("stats_model rows affected = %d, want %d", got, modelCount)
	}
}
