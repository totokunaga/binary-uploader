package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// CalculateChecksum computes the SHA256 checksum of a file in a streaming fashion
func CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum calculation: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Failed to close file")
		}
	}()

	hash := sha256.New()

	// Reads the file content and writes it to the hash in streaming fashion
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("error calculating checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
