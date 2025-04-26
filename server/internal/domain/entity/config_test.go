package entity

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Test default value when env var not set
	err := os.Unsetenv("TEST_ENV_VAR")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	value := GetEnv("TEST_ENV_VAR", "default")
	if value != "default" {
		t.Errorf("expected default value 'default', got %s", value)
	}

	// Test env var value when set
	err = os.Setenv("TEST_ENV_VAR", "custom")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	defer func() {
		err := os.Unsetenv("TEST_ENV_VAR")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}()

	value = GetEnv("TEST_ENV_VAR", "default")
	if value != "custom" {
		t.Errorf("expected custom value 'custom', got %s", value)
	}
}

func TestGetEnvInt(t *testing.T) {
	// Test default value when env var not set
	err := os.Unsetenv("TEST_ENV_INT")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	value := GetEnvInt("TEST_ENV_INT", 42)
	if value != 42 {
		t.Errorf("expected default value 42, got %d", value)
	}

	// Test env var value when set
	err = os.Setenv("TEST_ENV_INT", "100")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	defer func() {
		err := os.Unsetenv("TEST_ENV_INT")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}()

	value = GetEnvInt("TEST_ENV_INT", 42)
	if value != 100 {
		t.Errorf("expected custom value 100, got %d", value)
	}

	// Test invalid int value
	err = os.Setenv("TEST_ENV_INT", "invalid")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	value = GetEnvInt("TEST_ENV_INT", 42)
	if value != 42 {
		t.Errorf("expected default value 42 for invalid input, got %d", value)
	}
}

func TestGetEnvUint64(t *testing.T) {
	// Test default value when env var not set
	err := os.Unsetenv("TEST_ENV_UINT")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	value := GetEnvUint64("TEST_ENV_UINT", 1024)
	if value != 1024 {
		t.Errorf("expected default value 1024, got %d", value)
	}

	// Test env var value when set
	err = os.Setenv("TEST_ENV_UINT", "2048")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	defer func() {
		err := os.Unsetenv("TEST_ENV_UINT")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}()

	value = GetEnvUint64("TEST_ENV_UINT", 1024)
	if value != 2048 {
		t.Errorf("expected custom value 2048, got %d", value)
	}

	// Test invalid uint value
	err = os.Setenv("TEST_ENV_UINT", "invalid")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
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
		err := os.Unsetenv(env)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
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
	err := os.Setenv("PORT", "9090")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("BASE_STORAGE_DIR", "/tmp/storage")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("UPLOAD_SIZE_LIMIT", "200000000")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("UPLOAD_TIMEOUT", "10")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("DB_HOST", "db.example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("DB_PORT", "3307")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("DB_USER", "testuser")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("DB_PASSWORD", "testpass")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("DB_NAME", "testdb")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("DB_CONN_TIMEOUT", "20")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	err = os.Setenv("WORKER_POOL_SIZE", "10")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Cleanup after test
	defer func() {
		for _, env := range envVars {
			err := os.Unsetenv(env)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
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
