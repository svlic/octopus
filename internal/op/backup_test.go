package op

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/bestruirui/octopus/internal/db"
	"github.com/bestruirui/octopus/internal/model"
)

// largeImportRowCount exceeds modernc SQLite SQLITE_MAX_VARIABLE_NUMBER (~32766)
// for relay_logs (~17 binds/row → unbatched Create fails above ~1927 rows).
const largeImportRowCount = 2500

func Test_DBImportIncremental_large_stats_and_logs(t *testing.T) {
	// Given
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	dump := &model.DBDump{
		Version:      dbDumpVersion,
		IncludeStats: true,
		IncludeLogs:  true,
		StatsModel:   make([]model.StatsModel, largeImportRowCount),
		StatsDaily:   make([]model.StatsDaily, largeImportRowCount),
		RelayLogs:    make([]model.RelayLog, largeImportRowCount),
	}
	for i := 0; i < largeImportRowCount; i++ {
		dump.StatsModel[i] = model.StatsModel{
			ID:   i + 1,
			Name: fmt.Sprintf("model-%d", i),
			StatsMetrics: model.StatsMetrics{
				RequestSuccess: int64(i + 1),
			},
		}
		dump.StatsDaily[i] = model.StatsDaily{
			Date: fmt.Sprintf("%08d", i+1),
			StatsMetrics: model.StatsMetrics{
				RequestSuccess: int64(i + 1),
			},
		}
		dump.RelayLogs[i] = model.RelayLog{
			ID:               int64(i + 1),
			Time:             int64(1_700_000_000 + i),
			RequestModelName: fmt.Sprintf("model-%d", i),
			ChannelId:        1,
		}
	}

	// When
	res, err := DBImportIncremental(context.Background(), dump)

	// Then
	if err != nil {
		t.Fatalf("DBImportIncremental() error = %v", err)
	}
	if got := res.RowsAffected["stats_model"]; got != int64(largeImportRowCount) {
		t.Fatalf("stats_model rows = %d, want %d", got, largeImportRowCount)
	}
	if got := res.RowsAffected["stats_daily"]; got != int64(largeImportRowCount) {
		t.Fatalf("stats_daily rows = %d, want %d", got, largeImportRowCount)
	}
	if got := res.RowsAffected["relay_logs"]; got != int64(largeImportRowCount) {
		t.Fatalf("relay_logs rows = %d, want %d", got, largeImportRowCount)
	}

	var modelCount, dailyCount, logCount int64
	if err := db.GetDB().Model(&model.StatsModel{}).Count(&modelCount).Error; err != nil {
		t.Fatalf("count stats_model: %v", err)
	}
	if err := db.GetDB().Model(&model.StatsDaily{}).Count(&dailyCount).Error; err != nil {
		t.Fatalf("count stats_daily: %v", err)
	}
	if err := db.GetDB().Model(&model.RelayLog{}).Count(&logCount).Error; err != nil {
		t.Fatalf("count relay_logs: %v", err)
	}
	if modelCount != largeImportRowCount || dailyCount != largeImportRowCount || logCount != largeImportRowCount {
		t.Fatalf("db counts = model:%d daily:%d logs:%d, want all %d",
			modelCount, dailyCount, logCount, largeImportRowCount)
	}
}

func Test_DBImportIncremental_unbatched_create_hits_sqlite_variable_limit(t *testing.T) {
	// Given — documents the failure mode fixed by createBatched.
	if err := db.InitDB("sqlite", filepath.Join(t.TempDir(), "octopus.db"), false); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	rows := make([]model.RelayLog, largeImportRowCount)
	for i := range rows {
		rows[i] = model.RelayLog{
			ID:   int64(i + 1),
			Time: int64(1_700_000_000 + i),
		}
	}

	// When — single multi-row Create without batching
	err := db.GetDB().Create(&rows).Error

	// Then — must fail under SQLite bind limit when batching is absent
	if err == nil {
		t.Fatal("unbatched Create() error = nil, want SQLite variable-limit failure")
	}
}
