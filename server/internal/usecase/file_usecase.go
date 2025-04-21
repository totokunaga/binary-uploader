package usecase

import (
	"errors"
)

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrFileAlreadyExists = errors.New("file already exists")
	ErrInvalidFileName   = errors.New("invalid file name")
	ErrInvalidUploadID   = errors.New("invalid upload ID")
	ErrInvalidChunkID    = errors.New("invalid chunk ID")
	ErrChunkWriteFailed  = errors.New("chunk write failed")
	ErrFileSizeTooLarge  = errors.New("file size too large")
)
