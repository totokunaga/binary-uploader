package usecase

import (
	"fmt"
	"os"
	"path/filepath"
)

func checkLocalFileExists(filePath string) (string, int64, error) {
	// checks the existence of the file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.IsDir() {
		return "", 0, fmt.Errorf("cannot upload a directory, please provide a file")
	}

	// checks the file isn't empty
	fileSize := fileInfo.Size()
	if fileSize == 0 {
		return "", 0, fmt.Errorf("file is empty")
	}

	// checks the file name is valid
	fileName := filepath.Base(filePath)
	if fileName == "" {
		return "", 0, fmt.Errorf("invalid file name")
	}

	return fileName, fileSize, nil
}
