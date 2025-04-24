package entity

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test default value when env var not set
	os.Unsetenv("TEST_ENV_VAR")
	value := GetEnv("TEST_ENV_VAR", "default")
	if value != "default" {
		t.Errorf("expected default value 'default', got %s", value)
	}

	// Test env var value when set
	os.Setenv("TEST_ENV_VAR", "custom")
	defer os.Unsetenv("TEST_ENV_VAR")

	value = GetEnv("TEST_ENV_VAR", "default")
	if value != "custom" {
		t.Errorf("expected custom value 'custom', got %s", value)
	}
}

func TestGetEnvInt(t *testing.T) {
	// Test default value when env var not set
	os.Unsetenv("TEST_ENV_INT")
	value := GetEnvInt("TEST_ENV_INT", 42)
	if value != 42 {
		t.Errorf("expected default value 42, got %d", value)
	}

	// Test env var value when set
	os.Setenv("TEST_ENV_INT", "100")
	defer os.Unsetenv("TEST_ENV_INT")

	value = GetEnvInt("TEST_ENV_INT", 42)
	if value != 100 {
		t.Errorf("expected custom value 100, got %d", value)
	}

	// Test invalid int value
	os.Setenv("TEST_ENV_INT", "invalid")
	value = GetEnvInt("TEST_ENV_INT", 42)
	if value != 42 {
		t.Errorf("expected default value 42 for invalid input, got %d", value)
	}
}

func TestGetEnvUint64(t *testing.T) {
	// Test default value when env var not set
	os.Unsetenv("TEST_ENV_UINT")
	value := GetEnvUint64("TEST_ENV_UINT", 1024)
	if value != 1024 {
		t.Errorf("expected default value 1024, got %d", value)
	}

	// Test env var value when set
	os.Setenv("TEST_ENV_UINT", "2048")
	defer os.Unsetenv("TEST_ENV_UINT")

	value = GetEnvUint64("TEST_ENV_UINT", 1024)
	if value != 2048 {
		t.Errorf("expected custom value 2048, got %d", value)
	}

	// Test invalid uint value
	os.Setenv("TEST_ENV_UINT", "invalid")
	value = GetEnvUint64("TEST_ENV_UINT", 1024)
	if value != 1024 {
		t.Errorf("expected default value 1024 for invalid input, got %d", value)
	}
}

func TestNewConfig(t *testing.T) {
	// Clear environment variables to test defaults
	envVars := []string{
		"PORT", "BASE_STORAGE_DIR", "UPLOAD_SIZE_LIMIT", "UPLOAD_TIMEOUT",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
		"DB_CONN_TIMEOUT", "WORKER_POOL_SIZE",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}

	config := NewConfig()

	if config.Port != DefaultPort {
		t.Errorf("expected default port %d, got %d", DefaultPort, config.Port)
	}

	if config.BaseStorageDir != DefaultBaseStorageDir {
		t.Errorf("expected default storage dir %s, got %s", DefaultBaseStorageDir, config.BaseStorageDir)
	}

	if config.WorkerPoolSize != DefaultWorkerPoolSize {
		t.Errorf("expected default worker pool size %d, got %d", DefaultWorkerPoolSize, config.WorkerPoolSize)
	}

	// Test with custom values
	os.Setenv("PORT", "9090")
	os.Setenv("BASE_STORAGE_DIR", "/tmp/storage")
	os.Setenv("UPLOAD_SIZE_LIMIT", "200000000")
	os.Setenv("UPLOAD_TIMEOUT", "10")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "3307")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_CONN_TIMEOUT", "20")
	os.Setenv("WORKER_POOL_SIZE", "10")

	// Cleanup after test
	defer func() {
		for _, env := range envVars {
			os.Unsetenv(env)
		}
	}()

	config = NewConfig()

	if config.Port != 9090 {
		t.Errorf("expected custom port 9090, got %d", config.Port)
	}

	if config.BaseStorageDir != "/tmp/storage" {
		t.Errorf("expected custom storage dir /tmp/storage, got %s", config.BaseStorageDir)
	}

	if config.DBHost != "db.example.com" {
		t.Errorf("expected custom DB host db.example.com, got %s", config.DBHost)
	}

	if config.DBPort != 3307 {
		t.Errorf("expected custom DB port 3307, got %d", config.DBPort)
	}

	if config.DBUser != "testuser" {
		t.Errorf("expected custom DB user testuser, got %s", config.DBUser)
	}

	if config.DBPassword != "testpass" {
		t.Errorf("expected custom DB password testpass, got %s", config.DBPassword)
	}

	if config.DBName != "testdb" {
		t.Errorf("expected custom DB name testdb, got %s", config.DBName)
	}

	if config.DBConnTimeout != 20 {
		t.Errorf("expected custom DB connection timeout 20, got %d", config.DBConnTimeout)
	}

	if config.WorkerPoolSize != 10 {
		t.Errorf("expected custom worker pool size 10, got %d", config.WorkerPoolSize)
	}
}
