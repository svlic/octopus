package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterAfterAutoMigration(Migration{
		Version: 3,
		Up:      migrateStatsModelGlobalByName,
	})
}

// 003: stats_models becomes global-by-name (drop channel_id; unique name).
// Existing per-channel rows are collapsed by summing metrics for the same name.
func migrateStatsModelGlobalByName(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if !db.Migrator().HasTable("stats_models") {
		return nil
	}

	dialect := db.Dialector.Name()
	hasColumn := func(table, column string) bool {
		switch dialect {
		case "sqlite":
			var name string
			db.Raw("SELECT name FROM pragma_table_info(?) WHERE name = ? LIMIT 1", table, column).Scan(&name)
			return name == column
		case "mysql":
			var count int64
			db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?", table, column).Scan(&count)
			return count > 0
		case "postgres":
			var count int64
			db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = ?", table, column).Scan(&count)
			return count > 0
		default:
			return db.Migrator().HasColumn(table, column)
		}
	}

	if !hasColumn("stats_models", "channel_id") {
		return nil
	}

	// Aggregate duplicate names into a temp table, then rebuild stats_models.
	switch dialect {
	case "sqlite":
		stmts := []string{
			`CREATE TABLE stats_models_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				input_token INTEGER DEFAULT 0,
				output_token INTEGER DEFAULT 0,
				input_cost REAL DEFAULT 0,
				output_cost REAL DEFAULT 0,
				wait_time INTEGER DEFAULT 0,
				request_success INTEGER DEFAULT 0,
				request_failed INTEGER DEFAULT 0
			)`,
			`INSERT INTO stats_models_new (name, input_token, output_token, input_cost, output_cost, wait_time, request_success, request_failed)
			 SELECT name,
			        COALESCE(SUM(input_token), 0),
			        COALESCE(SUM(output_token), 0),
			        COALESCE(SUM(input_cost), 0),
			        COALESCE(SUM(output_cost), 0),
			        COALESCE(SUM(wait_time), 0),
			        COALESCE(SUM(request_success), 0),
			        COALESCE(SUM(request_failed), 0)
			 FROM stats_models
			 WHERE name IS NOT NULL AND TRIM(name) != ''
			 GROUP BY name`,
			`DROP TABLE stats_models`,
			`ALTER TABLE stats_models_new RENAME TO stats_models`,
		}
		for _, sql := range stmts {
			if err := db.Exec(sql).Error; err != nil {
				return fmt.Errorf("sqlite migrate stats_models: %w", err)
			}
		}
	case "mysql":
		stmts := []string{
			`CREATE TABLE stats_models_new (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				input_token BIGINT DEFAULT 0,
				output_token BIGINT DEFAULT 0,
				input_cost DOUBLE DEFAULT 0,
				output_cost DOUBLE DEFAULT 0,
				wait_time BIGINT DEFAULT 0,
				request_success BIGINT DEFAULT 0,
				request_failed BIGINT DEFAULT 0,
				UNIQUE KEY idx_stats_models_name (name)
			)`,
			`INSERT INTO stats_models_new (name, input_token, output_token, input_cost, output_cost, wait_time, request_success, request_failed)
			 SELECT name,
			        COALESCE(SUM(input_token), 0),
			        COALESCE(SUM(output_token), 0),
			        COALESCE(SUM(input_cost), 0),
			        COALESCE(SUM(output_cost), 0),
			        COALESCE(SUM(wait_time), 0),
			        COALESCE(SUM(request_success), 0),
			        COALESCE(SUM(request_failed), 0)
			 FROM stats_models
			 WHERE name IS NOT NULL AND TRIM(name) != ''
			 GROUP BY name`,
			`DROP TABLE stats_models`,
			`RENAME TABLE stats_models_new TO stats_models`,
		}
		for _, sql := range stmts {
			if err := db.Exec(sql).Error; err != nil {
				return fmt.Errorf("mysql migrate stats_models: %w", err)
			}
		}
	case "postgres":
		stmts := []string{
			`CREATE TABLE stats_models_new (
				id BIGSERIAL PRIMARY KEY,
				name TEXT NOT NULL UNIQUE,
				input_token BIGINT DEFAULT 0,
				output_token BIGINT DEFAULT 0,
				input_cost DOUBLE PRECISION DEFAULT 0,
				output_cost DOUBLE PRECISION DEFAULT 0,
				wait_time BIGINT DEFAULT 0,
				request_success BIGINT DEFAULT 0,
				request_failed BIGINT DEFAULT 0
			)`,
			`INSERT INTO stats_models_new (name, input_token, output_token, input_cost, output_cost, wait_time, request_success, request_failed)
			 SELECT name,
			        COALESCE(SUM(input_token), 0),
			        COALESCE(SUM(output_token), 0),
			        COALESCE(SUM(input_cost), 0),
			        COALESCE(SUM(output_cost), 0),
			        COALESCE(SUM(wait_time), 0),
			        COALESCE(SUM(request_success), 0),
			        COALESCE(SUM(request_failed), 0)
			 FROM stats_models
			 WHERE name IS NOT NULL AND TRIM(name) != ''
			 GROUP BY name`,
			`DROP TABLE stats_models`,
			`ALTER TABLE stats_models_new RENAME TO stats_models`,
		}
		for _, sql := range stmts {
			if err := db.Exec(sql).Error; err != nil {
				return fmt.Errorf("postgres migrate stats_models: %w", err)
			}
		}
	default:
		// Best-effort: drop column if present; AutoMigrate unique index covers new DBs.
		if err := db.Migrator().DropColumn("stats_models", "channel_id"); err != nil {
			return fmt.Errorf("drop stats_models.channel_id: %w", err)
		}
	}

	return nil
}
