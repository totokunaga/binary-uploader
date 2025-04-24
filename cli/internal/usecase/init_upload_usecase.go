package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
	"github.com/tomoya.tokunaga/cli/internal/util"
)

// InitUploadUsecase handles initializing file uploads
type initUploadUsecase struct {
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewInitUploadUsecase creates a new init upload usecase
func NewInitUploadUsecase(fileClient infrastructure.FileServerHttpClient) *initUploadUsecase {
	return &initUploadUsecase{
		fileServerHttpClient: fileClient,
	}
}

// Execute prechecks the file before initializing a file upload on the server
func (s *initUploadUsecase) ExecutePrecheck(ctx context.Context, input *InitUploadPrecheckUsecaseInput) (action PostPrecheckAction, output *InitUploadPrecheckUsecaseOutput, err error) {
	_, fileSize, err := checkLocalFileExists(input.FilePath)
	if err != nil {
		return ReturnError, nil, err
	}

	checksum, err := util.CalculateChecksum(input.FilePath)
	if err != nil {
		return ReturnError, nil, fmt.Errorf("failed to calculate file checksum: %w", err)
	}

	output = &InitUploadPrecheckUsecaseOutput{
		Checksum: checksum,
		FileSize: fileSize,
	}

	// checks the existence of the file on the server using targetFileName
	fileStats, err := s.fileServerHttpClient.GetFileStats(ctx, input.TargetFileName)
	if err != nil {
		return ReturnError, nil, fmt.Errorf("failed to get file stats for '%s': %w", input.TargetFileName, err)
	}
	if fileStats == nil {
		return ProceedWithInit, output, nil
	}

	isSameFile := fileStats.Checksum == checksum && fileStats.Size == uint64(fileSize)
	if !isSameFile {
		// the file on the server has different content, so suggest to delete the existing entry and to retry the upload
		return SuggestExistingEntryDeletion, output, nil
	}
	switch fileStats.Status {
	case entity.FileStatusUploaded:
		// the file on the server has same content, so do nothing and exit
		return Exits, nil, nil
	case entity.FileStatusInitialized:
		// the file upload was initialized, but not completed, so proceed with the upload
		return ProceedWithInit, output, nil
	case entity.FileStatusFailed:
		// the file upload failed previously (e.g. database/storage error, network disconnection, etc), so proceed with
		// re-uploading missing chunks
		return ProceedWithReUpload, output, nil
	case entity.FileStatusInProgress:
		// the file upload is being processed by other client or something went wrong on the server side (e.g. server crash,
		// database down, etc) and the file data is orphaned. determine if the file is orphaned and proceed with re-uploading
		// missing chunks if the last updated time is older than the server's upload timeout.
		isOrphaned := fileStats.UpdatedAt.Before(time.Now().Add(-fileStats.UploadTimeoutSecond))
		if isOrphaned {
			return ProceedWithReUpload, output, nil
		}
		return Exits, nil, fmt.Errorf("remote file server processing '%s' with the same content currently or faces some problems (database crash, server error, etc). try again later, or delete the existing entry and retry the upload", input.TargetFileName)
	}

	// when the file is one of deleted status (DELETE_INITIALIZED, DELETE_IN_PROGRESS, DELETE_FAILED), suggest to
	// delete the existing entry and to retry the upload
	return SuggestExistingEntryDeletion, nil, nil
}

// Execute initializes a file upload on the server
func (s *initUploadUsecase) Execute(ctx context.Context, input *InitUploadUsecaseInput) (*UploadUsecaseOutput, error) {
	_, fileSize, err := checkLocalFileExists(input.FilePath)
	if err != nil {
		return nil, err
	}

	// checks the file content is the same as the original checksum in precheck
	checksum, err := util.CalculateChecksum(input.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file checksum: %w", err)
	}
	if checksum != input.OriginalChecksum {
		return nil, fmt.Errorf("file content has changed for local file '%s' since precheck", input.FilePath)
	}

	// Calculate numChunks as int64
	numChunks := (fileSize + input.ChunkSize - 1) / input.ChunkSize
	reqBody := infrastructure.UploadInitRequest{
		TotalSize:   fileSize,
		TotalChunks: numChunks,
		ChunkSize:   input.ChunkSize,
		Checksum:    checksum,
		IsReUpload:  input.IsReUpload,
	}
	res, err := s.fileServerHttpClient.InitUpload(ctx, input.TargetFileName, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize upload for '%s': %w", input.TargetFileName, err)
	}

	// populate values required to re-upload failed chunks from previously upload attempts
	uploadChunkSize := input.ChunkSize
	missingChunkNumberMap := make(map[uint64]struct{})
	if res.MissingChunkInfo != nil {
		uploadChunkSize = int64(res.MissingChunkInfo.MaxChunkSize)
		for _, chunkNumber := range res.MissingChunkInfo.ChunkNumbers {
			missingChunkNumberMap[chunkNumber] = struct{}{}
		}
	}

	return &UploadUsecaseOutput{
		UploadID:              res.UploadID,
		UploadChunkSize:       uint64(uploadChunkSize),
		MissingChunkNumberMap: missingChunkNumberMap,
	}, nil
}
