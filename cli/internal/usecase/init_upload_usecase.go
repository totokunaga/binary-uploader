package usecase

import (
	"fmt"
	"time"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
	"github.com/tomoya.tokunaga/cli/internal/infrastructure"
	"github.com/tomoya.tokunaga/cli/internal/util"
)

// InitUploadUsecase handles initializing file uploads
type InitUploadUsecase struct {
	config               *entity.Config
	fileServerHttpClient infrastructure.FileServerHttpClient
}

// NewInitUploadUsecase creates a new init upload usecase
func NewInitUploadUsecase(config *entity.Config) *InitUploadUsecase {
	client := infrastructure.NewFileServerV1HttpClient(config.ServerURL)
	return &InitUploadUsecase{
		config:               config,
		fileServerHttpClient: client,
	}
}

type UploadPrecheckUsecaseOutput struct {
	Checksum string
	FileSize int64
}

type UploadUsecaseOutput struct {
	UploadID              uint64
	UploadChunkSize       uint64
	MissingChunkNumberMap map[uint64]struct{}
}

// PostPrecheckAction defines the possible actions after the precheck.
type PostPrecheckAction int

const (
	ProceedWithInit PostPrecheckAction = iota
	ProceedWithReUpload
	SuggestExistingEntryDeletion
	Exits
	ReturnError
)

// Execute prechecks the file before initializing a file upload on the server
func (s *InitUploadUsecase) ExecutePrecheck(filePath string, targetFileName string) (action PostPrecheckAction, output *UploadPrecheckUsecaseOutput, err error) {
	_, fileSize, err := checkLocalFileExists(filePath)
	if err != nil {
		return ReturnError, nil, err
	}

	checksum, err := util.CalculateChecksum(filePath)
	if err != nil {
		return ReturnError, nil, fmt.Errorf("failed to calculate file checksum: %w", err)
	}

	output = &UploadPrecheckUsecaseOutput{
		Checksum: checksum,
		FileSize: fileSize,
	}

	// checks the existence of the file on the server using targetFileName
	fileStats, err := s.fileServerHttpClient.GetFileStats(targetFileName)
	if err != nil {
		return ReturnError, nil, fmt.Errorf("failed to get file stats for '%s': %w", targetFileName, err)
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
		// missing chunks if the last updated time is older than 5 minutes.
		isOrphaned := fileStats.UpdatedAt.Before(time.Now().Add(-fileStats.UploadTimeoutSecond))
		if isOrphaned {
			return ProceedWithReUpload, output, nil
		}
		return Exits, nil, fmt.Errorf("remote file server processing %s with the same content currently. try again later", targetFileName)
	}

	// when the file is one of deleted status (DELETE_INITIALIZED, DELETE_IN_PROGRESS, DELETE_FAILED), suggest to
	// delete the existing entry and to retry the upload
	return SuggestExistingEntryDeletion, nil, nil
}

// Execute initializes a file upload on the server
func (s *InitUploadUsecase) Execute(filePath string, targetFileName string, originalChecksum string, chunkSize int64, isReUpload bool) (*UploadUsecaseOutput, error) {
	_, fileSize, err := checkLocalFileExists(filePath)
	if err != nil {
		return nil, err
	}

	// checks the file content is the same as the original checksum in precheck
	checksum, err := util.CalculateChecksum(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate file checksum: %w", err)
	}
	if checksum != originalChecksum {
		return nil, fmt.Errorf("file content has changed for local file '%s' since precheck", filePath)
	}

	// Calculate numChunks as int64
	numChunks := (fileSize + chunkSize - 1) / chunkSize
	reqBody := infrastructure.UploadInitRequest{
		TotalSize:   fileSize,
		TotalChunks: numChunks,
		ChunkSize:   chunkSize,
		Checksum:    checksum,
		IsReUpload:  isReUpload,
	}
	res, err := s.fileServerHttpClient.InitUpload(targetFileName, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize upload for '%s': %w", targetFileName, err)
	}

	// populate values required to re-upload failed chunks from previously upload attempts
	uploadChunkSize := chunkSize
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
