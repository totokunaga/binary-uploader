package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateChecksum(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "checksum-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	testFiles := map[string]string{
		"empty.txt":     "",
		"simple.txt":    "Hello, world!",
		"multiline.txt": "Line 1\nLine 2\nLine 3",
	}

	filePaths := make(map[string]string)
	for name, content := range testFiles {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
		filePaths[name] = path
	}

	// Calculate expected checksums manually
	// empty.txt: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	// simple.txt: 315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3
	// multiline.txt: 391ba54caa9e9da3dd31dca1eff275e706979e76c1f60c91401f0624734f52b0

	tests := []struct {
		name        string
		filePath    string
		expected    string
		expectError bool
	}{
		{
			name:        "Empty file",
			filePath:    filePaths["empty.txt"],
			expected:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError: false,
		},
		{
			name:        "Simple file",
			filePath:    filePaths["simple.txt"],
			expected:    "315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3",
			expectError: false,
		},
		{
			name:        "Multiline file",
			filePath:    filePaths["multiline.txt"],
			expected:    "391ba54caa9e9da3dd31dca1eff275e706979e76c1f60c91401f0624734f52b0",
			expectError: false,
		},
		{
			name:        "Non-existent file",
			filePath:    filepath.Join(tempDir, "does-not-exist.txt"),
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			checksum, err := CalculateChecksum(tt.filePath)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError && checksum != tt.expected {
				t.Errorf("Expected checksum %s but got %s", tt.expected, checksum)
			}
		})
	}
}
