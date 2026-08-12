package migrate

import (
	"testing"

	"github.com/bestruirui/octopus/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestMigrateStatsModelIdentityConsolidatesDuplicates(t *testing.T) {
	// Given
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := conn.AutoMigrate(&model.StatsModel{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if err := conn.Migrator().DropIndex(&model.StatsModel{}, "idx_stats_models_name_channel"); err != nil {
		t.Fatalf("DropIndex() error = %v", err)
	}
	rows := []model.StatsModel{
		{Name: "gpt-4o", ChannelID: 1, StatsMetrics: model.StatsMetrics{InputToken: 100}},
		{Name: "gpt-4o", ChannelID: 1, StatsMetrics: model.StatsMetrics{OutputToken: 50}},
	}
	if err := conn.Create(&rows).Error; err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// When
	if err := migrateStatsModelIdentity(conn); err != nil {
		t.Fatalf("migrateStatsModelIdentity() error = %v", err)
	}
	if err := conn.Migrator().CreateIndex(&model.StatsModel{}, "idx_stats_models_name_channel"); err != nil {
		t.Fatalf("CreateIndex() error = %v", err)
	}

	// Then
	var stats []model.StatsModel
	if err := conn.Find(&stats).Error; err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if len(stats) != 1 || stats[0].InputToken != 100 || stats[0].OutputToken != 50 {
		t.Fatalf("model stats = %+v, want one row with aggregated metrics", stats)
	}
	duplicate := model.StatsModel{Name: "gpt-4o", ChannelID: 1}
	if err := conn.Create(&duplicate).Error; err == nil {
		t.Fatal("duplicate model/channel insert succeeded, want unique index violation")
	}
}
