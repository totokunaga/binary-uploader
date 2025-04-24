package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"compress/gzip"

	"github.com/gin-gonic/gin"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"golang.org/x/exp/slog"
)

type FileUploadHandler struct {
	fileUploadUseCase usecase.FileUploadUseCase
	config            *entity.Config
	logger            *slog.Logger
}

func NewFileUploadHandler(fileUploadUseCase usecase.FileUploadUseCase, config *entity.Config, logger *slog.Logger) *FileUploadHandler {
	return &FileUploadHandler{
		fileUploadUseCase: fileUploadUseCase,
		config:            config,
		logger:            logger,
	}
}

type InitUploadRequest struct {
	Checksum    string `json:"checksum" binding:"required"`
	TotalSize   uint64 `json:"total_size" binding:"required"`
	TotalChunks uint   `json:"total_chunks" binding:"required"`
	ChunkSize   uint64 `json:"chunk_size" binding:"required"`
	IsReUpload  bool   `json:"is_reupload"`
}

type MissingChunkInfo struct {
	MaxChunkSize uint64   `json:"max_size"`
	ChunkNumbers []uint64 `json:"chunk_numbers"`
}

type InitUploadResponse struct {
	UploadID         uint64           `json:"upload_id"`
	MissingChunkInfo MissingChunkInfo `json:"missing_chunk_info"`
}

type UploadResponse struct {
	Status string `json:"status"`
}

// InitUpload handles file upload initialization
func (h *FileUploadHandler) ExecuteInit(ctx *gin.Context) {
	// Get the file name from the URL
	fileName := ctx.Param("file_name")
	if fileName == "" {
		sendErrorResponse(ctx, h.logger, e.NewInvalidInputError(errors.New("file_name parameter is required"), ""))
		return
	}

	// Validate file name
	if fileName == "." || fileName == ".." {
		sendErrorResponse(ctx, h.logger, e.NewInvalidInputError(errors.New("invalid file name"), ""))
		return
	}

	// Parse the request body
	var req InitUploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sendErrorResponse(ctx, h.logger, e.NewInvalidInputError(err, "invalid request body"))
		return
	}

	// Check if the file size is too large
	if req.TotalSize > h.config.UploadSizeLimit {
		sendErrorResponse(ctx, h.logger, e.NewInvalidInputError(
			errors.New("total_size is too large"),
			fmt.Sprintf("must be less than %d bytes", h.config.UploadSizeLimit),
		))
		return
	}

	// Initialize the upload
	fileRecord, missingChunks, err := h.fileUploadUseCase.ExecuteInit(ctx.Request.Context(), usecase.FileUploadUseCaseExecuteInitInput{
		FileName:    fileName,
		Checksum:    req.Checksum,
		TotalSize:   req.TotalSize,
		TotalChunks: req.TotalChunks,
		ChunkSize:   req.ChunkSize,
		IsReUpload:  req.IsReUpload,
	})
	if err != nil {
		sendErrorResponse(ctx, h.logger, err)
		return
	}

	// Form the response and send it back
	res := InitUploadResponse{
		UploadID: fileRecord.ID,
	}

	if len(missingChunks) > 0 {
		missingChunkInfo := MissingChunkInfo{
			MaxChunkSize: fileRecord.ChunkSize,
			ChunkNumbers: make([]uint64, len(missingChunks)),
		}
		for i, chunk := range missingChunks {
			missingChunkInfo.ChunkNumbers[i] = uint64(chunk.ChunkNumber)
		}
		res.MissingChunkInfo = missingChunkInfo
	}

	ctx.JSON(http.StatusOK, res)
}

// UploadChunk handles file chunk upload
func (h *FileUploadHandler) Execute(ctx *gin.Context) {
	// Get the upload ID and chunk ID from the URL
	fileIDStr := ctx.Param("file_id")
	chunkNumberStr := ctx.Param("chunk_number")

	fileID, err := strconv.ParseUint(fileIDStr, 10, 64)
	if err != nil {
		sendErrorResponse(ctx, h.logger, e.NewInvalidInputError(err, fmt.Sprintf("invalid file ID: %s", fileIDStr)))
		return
	}

	chunkNumber, err := strconv.ParseUint(chunkNumberStr, 10, 64)
	if err != nil {
		sendErrorResponse(ctx, h.logger, e.NewInvalidInputError(err, fmt.Sprintf("invalid chunk ID: %s", chunkNumberStr)))
		return
	}

	// Setup the reader from request body
	var reader io.Reader = ctx.Request.Body

	// Check if content is gzip encoded
	if ctx.GetHeader("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			sendErrorResponse(ctx, h.logger, e.NewInvalidInputError(err, "failed to create gzip reader"))
			return
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	ucErr := h.fileUploadUseCase.Execute(ctx, usecase.FileUploadUseCaseExecuteInput{
		FileID:      fileID,
		ChunkNumber: chunkNumber,
		Reader:      reader,
	})

	if ucErr == nil {
		ctx.JSON(http.StatusOK, UploadResponse{Status: "OK"})
		return
	}

	// Updates file and chunk status to the failed status
	if failRecovErr := h.fileUploadUseCase.ExecuteFailRecovery(ctx.Request.Context(), fileID, chunkNumber); failRecovErr != nil {
		sendErrorResponse(ctx, h.logger, failRecovErr)
		return
	}
	sendErrorResponse(ctx, h.logger, ucErr)
}
