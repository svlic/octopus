package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterBeforeAutoMigration(Migration{
		Version: 7,
		Up:      migrateStatsModelGlobalByName,
	})
}

func migrateStatsModelGlobalByName(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if !db.Migrator().HasTable("stats_models") || !db.Migrator().HasColumn("stats_models", "channel_id") {
		return nil
	}

	statements, err := statsModelMigrationStatements(db.Dialector.Name())
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("migrate stats_models: %w", err)
			}
		}
		return nil
	})
}

func statsModelMigrationStatements(dialect string) ([]string, error) {
	var create string
	var rename string
	switch dialect {
	case "sqlite":
		create = `CREATE TABLE stats_models_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			input_token INTEGER DEFAULT 0,
			output_token INTEGER DEFAULT 0,
			input_cost REAL DEFAULT 0,
			output_cost REAL DEFAULT 0,
			wait_time INTEGER DEFAULT 0,
			request_success INTEGER DEFAULT 0,
			request_failed INTEGER DEFAULT 0
		)`
		rename = `ALTER TABLE stats_models_new RENAME TO stats_models`
	case "mysql":
		create = `CREATE TABLE stats_models_new (
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
			)`
		rename = `RENAME TABLE stats_models TO stats_models_old, stats_models_new TO stats_models`

	case "postgres":
		create = `CREATE TABLE stats_models_new (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			input_token BIGINT DEFAULT 0,
			output_token BIGINT DEFAULT 0,
			input_cost DOUBLE PRECISION DEFAULT 0,
			output_cost DOUBLE PRECISION DEFAULT 0,
			wait_time BIGINT DEFAULT 0,
			request_success BIGINT DEFAULT 0,
			request_failed BIGINT DEFAULT 0
		)`
		rename = `ALTER TABLE stats_models_new RENAME TO stats_models`
	default:
		return nil, fmt.Errorf("unsupported database dialect: %s", dialect)
	}

	statements := []string{
		`DROP TABLE IF EXISTS stats_models_new`,
		create,
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
	}
	if dialect == "mysql" {
		return append(statements,
			`DROP TABLE IF EXISTS stats_models_old`,
			rename,
			`DROP TABLE stats_models_old`,
		), nil
	}
	return append(statements, `DROP TABLE stats_models`, rename), nil
}
