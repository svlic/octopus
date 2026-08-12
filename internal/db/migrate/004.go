package migrate

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/model"
	"gorm.io/gorm"
)

func init() {
	RegisterBeforeAutoMigration(Migration{
		Version: 4,
		Up:      migrateStatsModelIdentity,
	})
}

// 004: 合并重复的模型渠道统计，供 AutoMigrate 建立唯一索引。
func migrateStatsModelIdentity(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if !db.Migrator().HasTable(&model.StatsModel{}) {
		return nil
	}

	var rows []model.StatsModel
	if err := db.Find(&rows).Error; err != nil {
		return fmt.Errorf("failed to load model stats: %w", err)
	}

	type modelKey struct {
		Name      string
		ChannelID int
	}
	aggregated := make(map[modelKey]model.StatsModel, len(rows))
	for _, row := range rows {
		key := modelKey{Name: row.Name, ChannelID: row.ChannelID}
		current := aggregated[key]
		if current.ID == 0 || row.ID < current.ID {
			current.ID = row.ID
		}
		current.Name = row.Name
		current.ChannelID = row.ChannelID
		current.StatsMetrics.Add(row.StatsMetrics)
		aggregated[key] = current
	}
	if len(aggregated) == len(rows) {
		return nil
	}

	models := make([]model.StatsModel, 0, len(aggregated))
	for _, stats := range aggregated {
		models = append(models, stats)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.StatsModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete duplicate model stats: %w", err)
		}
		if len(models) == 0 {
			return nil
		}
		if err := tx.Create(&models).Error; err != nil {
			return fmt.Errorf("failed to restore model stats: %w", err)
		}
		return nil
	})
}
