package database

import (
	"fmt"

	"golang.org/x/exp/slog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	"github.com/tomoya.tokunaga/server/internal/interface/repository/database"
)

// NewDB creates and initializes the database connection
func NewDB(config *entity.Config, logger *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=%ds",
		config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName, config.DBConnTimeout)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to get database connection", "error", err)
		return nil, err
	}
	sqlDB.SetMaxIdleConns(config.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(config.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.DBConnMaxLifetime)

	// Auto-migrate the database based on the models
	if err := db.AutoMigrate(&database.FileModel{}, &database.FileChunkModel{}); err != nil {
		logger.Error("failed to auto-migrate database", "error", err)
		return nil, err
	}

	logger.Info("database connection established",
		"host", config.DBHost,
		"port", config.DBPort,
		"database", config.DBName)

	return db, nil
}
