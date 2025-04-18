package database

import (
	"fmt"

	"golang.org/x/exp/slog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/tomoya.tokunaga/server/internal/core/entity"
	mysqldriver "github.com/tomoya.tokunaga/server/internal/interface/database/mysql"
)

// NewDB creates and initializes the database connection
func NewDB(config *entity.Config, logger *slog.Logger) (*gorm.DB, error) {
	// Build connection string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%ds",
		config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName, config.DBConnTimeout)

	// Connect to database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return nil, err
	}

	// Auto-migrate the database models
	if err := db.AutoMigrate(&mysqldriver.FileModel{}, &mysqldriver.FileChunkModel{}); err != nil {
		logger.Error("failed to auto-migrate database", "error", err)
		return nil, err
	}

	logger.Info("database connection established",
		"host", config.DBHost,
		"port", config.DBPort,
		"database", config.DBName)

	return db, nil
}
