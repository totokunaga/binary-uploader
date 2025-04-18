package entity

import (
	"github.com/tomoya.tokunaga/server/internal/util"
)

// Config represents the application configuration loaded from environment variables
type Config struct {
	// Server config
	Port        int
	MaxBodySize int

	// Storage config
	BaseStorageDir  string
	UploadSizeLimit uint64

	// Database config
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	DBConnTimeout int
}

// NewConfig loads configuration from environment variables
func NewConfig() *Config {
	return &Config{
		// Server config
		Port:        util.GetEnvInt("PORT", 8080),
		MaxBodySize: util.GetEnvInt("MAX_BODY_SIZE", 10*1024*1024), // 10 MiB default

		// Storage config
		BaseStorageDir:  util.GetEnv("BASE_STORAGE_DIR", "."),
		UploadSizeLimit: util.GetEnvUint64("UPLOAD_SIZE_LIMIT", 100*1024*1024), // 100 MiB default

		// Database config
		DBHost:        util.GetEnv("DB_HOST", "localhost"),
		DBPort:        util.GetEnvInt("DB_PORT", 3306),
		DBUser:        util.GetEnv("DB_USER", "root"),
		DBPassword:    util.GetEnv("DB_PASSWORD", ""),
		DBName:        util.GetEnv("DB_NAME", "fs_store"),
		DBConnTimeout: util.GetEnvInt("DB_CONN_TIMEOUT", 10), // seconds
	}
}
