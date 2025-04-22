package error

import (
	"fmt"
	"net/http"
)

type FileStorageError struct {
	err    error
	errMsg string
}

func NewFileStorageError(err error, errMsg string) *FileStorageError {
	return &FileStorageError{
		err:    err,
		errMsg: errMsg,
	}
}

func (e *FileStorageError) Error() string {
	return fmt.Sprintf("%s: %v", e.errMsg, e.err)
}

func (e *FileStorageError) ErrorObject() error {
	return fmt.Errorf("%s: %w", e.errMsg, e.err)
}

func (e *FileStorageError) StatusCode() int {
	return http.StatusInternalServerError
}

func (e *FileStorageError) ErrorCode() string {
	return "FILE_STORAGE"
}
