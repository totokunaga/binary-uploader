package entity

import (
	"log"
	"os"
	"strconv"
	"time"
)

const (
	DefaultPort                = 8080
	DefaultBaseStorageDir      = "./storage"
	DefaultUploadTimeoutSecond = 5
	DefaultWorkerPoolSize      = 5
	DefaultStreamBufferSize    = 1024 * 1024 // 1MB

	DefaultDBHost               = "localhost"
	DefaultDBPort               = 3306
	DefaultDBUser               = "root"
	DefaultDBPassword           = ""
	DefaultDBName               = "fs_store"
	DefaultDBConnTimeout        = 10
	DefaultDBMaxIdleConns       = 10
	DefaultDBMaxOpenConns       = 100
	DefaultDBConnMaxLifetimeMin = 60
)

// Config represents the application configuration loaded from environment variables
type Config struct {
	// Server config
	Port int

	// Storage config
	BaseStorageDir      string
	StreamBufferSize    int
	UploadTimeoutSecond time.Duration

	// Database config
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBConnTimeout     int
	DBMaxIdleConns    int
	DBMaxOpenConns    int
	DBConnMaxLifetime time.Duration

	// Worker pool config
	WorkerPoolSize int
}

// NewConfig loads configuration from environment variables
func NewConfig() *Config {
	return &Config{
		// Server config
		Port: GetEnvInt("PORT", DefaultPort),

		// Storage config
		BaseStorageDir:      GetEnv("BASE_STORAGE_DIR", DefaultBaseStorageDir),
		StreamBufferSize:    GetEnvInt("STREAM_BUFFER_SIZE", DefaultStreamBufferSize),
		UploadTimeoutSecond: time.Duration(GetEnvInt("UPLOAD_TIMEOUT_SECOND", DefaultUploadTimeoutSecond)) * time.Second,

		// Database config
		DBHost:            GetEnv("DB_HOST", DefaultDBHost),
		DBPort:            GetEnvInt("DB_PORT", DefaultDBPort),
		DBUser:            GetEnv("DB_USER", DefaultDBUser),
		DBPassword:        GetEnv("DB_PASSWORD", DefaultDBPassword),
		DBName:            GetEnv("DB_NAME", DefaultDBName),
		DBConnTimeout:     GetEnvInt("DB_CONN_TIMEOUT", DefaultDBConnTimeout),
		DBMaxIdleConns:    GetEnvInt("DB_MAX_IDLE_CONNS", DefaultDBMaxIdleConns),
		DBMaxOpenConns:    GetEnvInt("DB_MAX_OPEN_CONNS", DefaultDBMaxOpenConns),
		DBConnMaxLifetime: time.Duration(GetEnvInt("DB_CONN_MAX_LIFETIME_MIN", DefaultDBConnMaxLifetimeMin)) * time.Minute,

		// Worker pool config
		WorkerPoolSize: GetEnvInt("WORKER_POOL_SIZE", DefaultWorkerPoolSize),
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
