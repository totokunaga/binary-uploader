package entity

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config represents the application configuration loaded from environment variables
type Config struct {
	// Server config
	Port        int
	MaxBodySize int

	// Storage config
	BaseStorageDir      string
	UploadSizeLimit     uint64
	UploadTimeoutSecond time.Duration

	// Database config
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	DBConnTimeout int

	// Worker pool config
	WorkerPoolSize int
}

// NewConfig loads configuration from environment variables
func NewConfig() *Config {
	return &Config{
		// Server config
		Port:        GetEnvInt("PORT", 8080),
		MaxBodySize: GetEnvInt("MAX_BODY_SIZE", 10*1024*1024), // 10 MiB default

		// Storage config
		BaseStorageDir:      GetEnv("BASE_STORAGE_DIR", "."),
		UploadSizeLimit:     GetEnvUint64("UPLOAD_SIZE_LIMIT", 100*1024*1024), // 100 MiB default
		UploadTimeoutSecond: time.Duration(GetEnvInt("UPLOAD_TIMEOUT", 5)) * time.Second,

		// Database config
		DBHost:        GetEnv("DB_HOST", "localhost"),
		DBPort:        GetEnvInt("DB_PORT", 3306),
		DBUser:        GetEnv("DB_USER", "root"),
		DBPassword:    GetEnv("DB_PASSWORD", ""),
		DBName:        GetEnv("DB_NAME", "fs_store"),
		DBConnTimeout: GetEnvInt("DB_CONN_TIMEOUT", 10), // seconds

		// Worker pool config
		WorkerPoolSize: GetEnvInt("WORKER_POOL_SIZE", 5),
	}
}

// GetEnv gets an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetEnvInt gets an environment variable as an integer or returns a default value
func GetEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("invalid value for %s: %s, using default: %d", key, value, defaultValue)
		return defaultValue
	}
	return intValue
}

// GetEnvUint64 gets an environment variable as an uint64 or returns a default value
func GetEnvUint64(key string, defaultValue uint64) uint64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	uint64Value, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		log.Printf("invalid value for %s: %s, using default: %d", key, value, defaultValue)
		return defaultValue
	}
	return uint64Value
}
