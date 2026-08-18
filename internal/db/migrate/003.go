package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterBeforeAutoMigration(Migration{
		Version: 8,
		Up:      migrateChannelTypeToProviderDirect,
	})
}

// migrateChannelTypeToProviderDirect 补偿已占用迁移版本 3 的 dyna 数据库，将数字渠道类型直接迁移为提供方类型。
func migrateChannelTypeToProviderDirect(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if !db.Migrator().HasTable("channels") || !db.Migrator().HasColumn("channels", "type") {
		return nil
	}

	switch db.Dialector.Name() {
	case "mysql":
		if err := db.Exec("ALTER TABLE `channels` MODIFY COLUMN `type` varchar(191)").Error; err != nil {
			return fmt.Errorf("failed to alter channels.type: %w", err)
		}
	case "postgres":
		if err := db.Exec(`ALTER TABLE "channels" ALTER COLUMN "type" TYPE text USING "type"::text`).Error; err != nil {
			return fmt.Errorf("failed to alter channels.type: %w", err)
		}
	}

	typeExpr := `CAST("type" AS TEXT)`
	typeColumn := `"type"`
	switch db.Dialector.Name() {
	case "mysql":
		typeExpr = "CAST(`type` AS CHAR)"
		typeColumn = "`type`"
	case "postgres":
		typeExpr = `"type"::text`
	}

	if err := db.Exec(fmt.Sprintf(`
UPDATE channels
SET %s = CASE %s
	WHEN '0' THEN ?
	WHEN '1' THEN ?
	WHEN '2' THEN ?
	WHEN '3' THEN ?
	WHEN '4' THEN ?
	WHEN '5' THEN ?
	ELSE %s
END
	`, typeColumn, typeExpr, typeColumn), "openai",
		"openai_responses",
		"anthropic",
		"gemini",
		"volcengine",
		"openai").Error; err != nil {

		return fmt.Errorf("failed to migrate channels.type: %w", err)
	}
	return nil
}
