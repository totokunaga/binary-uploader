package usecase

import (
	"context"

	"github.com/tomoya.tokunaga/cli/internal/domain/entity"
)

type ListUsecase interface {
	Execute(ctx context.Context) (*entity.ListFilesResp, error)
}

type InitUploadUsecase interface {
	ExecutePrecheck(ctx context.Context, input *InitUploadPrecheckUsecaseInput) (action PostPrecheckAction, output *InitUploadPrecheckUsecaseOutput, err error)
	Execute(ctx context.Context, input *InitUploadUsecaseInput) (*UploadUsecaseOutput, error)
}

type UploadUsecase interface {
	Execute(ctx context.Context, input *UploadUsecaseInput) error
}

type DeleteUsecase interface {
	Execute(ctx context.Context, targetFileName string) error
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

type InitUploadPrecheckUsecaseInput struct {
	FilePath       string
	TargetFileName string
}

type InitUploadPrecheckUsecaseOutput struct {
	Checksum string
	FileSize int64
}

type InitUploadUsecaseInput struct {
	FilePath         string
	TargetFileName   string
	OriginalChecksum string
	ChunkSize        int64
	IsReUpload       bool
}

type UploadUsecaseOutput struct {
	UploadID              uint64
	UploadChunkSize       uint64
	MissingChunkNumberMap map[uint64]struct{}
}

type UploadUsecaseInput struct {
	UploadID              uint64
	FilePath              string
	ChunkSize             int64
	IsReUpload            bool
	MissingChunkNumberMap map[uint64]struct{}
	ProgressCb            func(size int64)
}
