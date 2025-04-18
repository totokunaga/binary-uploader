package util

import (
	"log"
	"os"
	"strconv"
)

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
