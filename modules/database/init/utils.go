package init

import (
	"gorm.io/gorm/logger"

	"userstyles.world/modules/config"
	"userstyles.world/modules/database"
)

func logLevel() logger.LogLevel {
	switch config.DBDebug {
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Silent
	}
}

func migrate(t ...interface{}) error {
	return database.Conn.AutoMigrate(t...)
}

func drop(t ...interface{}) error {
	return database.Conn.Migrator().DropTable(t...)
}
