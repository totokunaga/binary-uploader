package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	handler_err "github.com/tomoya.tokunaga/server/internal/interface/api/handler/error"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"golang.org/x/exp/slog"
)

type FileUploadHandler struct {
	fileUploadUseCase usecase.FileUploadUseCase
	logger            *slog.Logger
}

func NewFileUploadHandler(fileUploadUseCase usecase.FileUploadUseCase, logger *slog.Logger) *FileUploadHandler {
	return &FileUploadHandler{
		fileUploadUseCase: fileUploadUseCase,
		logger:            logger,
	}
}

type InitUploadRequest struct {
	TotalSize   uint64 `json:"total_size" binding:"required"`
	TotalChunks uint   `json:"total_chunks" binding:"required"`
}

type InitUploadResponse struct {
	UploadID uint64 `json:"upload_id"`
}

type UploadResponse struct {
	Status string `json:"status"`
}

// InitUpload handles file upload initialization
func (h *FileUploadHandler) ExecuteInit(ctx *gin.Context) {
	// Get the file name from the URL
	fileName := ctx.Param("file_name")
	if fileName == "" {
		ctx.JSON(http.StatusBadRequest, handler_err.GetErrorResponse(
			e.NewInvalidInputError(errors.New("file_name parameter is required"), ""),
			h.logger,
		))
		return
	}

	// Parse the request body
	var req InitUploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, handler_err.GetErrorResponse(
			e.NewInvalidInputError(err, "invalid request body"),
			h.logger,
		))
		return
	}

	// Initialize the upload
	uploadID, err := h.fileUploadUseCase.ExecuteInit(ctx.Request.Context(), fileName, req.TotalSize, req.TotalChunks)
	if err != nil {
		ctx.JSON(err.StatusCode(), handler_err.GetErrorResponse(err, h.logger))
		return
	}

	res := InitUploadResponse{UploadID: uploadID}
	ctx.JSON(http.StatusOK, res)
}

// UploadChunk handles file chunk upload
func (h *FileUploadHandler) Execute(ctx *gin.Context) {
	// Get the upload ID and chunk ID from the URL
	uploadIDStr := ctx.Param("upload_id")
	chunkIDStr := ctx.Param("chunk_id")

	uploadID, err := strconv.ParseUint(uploadIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, handler_err.GetErrorResponse(
			e.NewInvalidInputError(err, fmt.Sprintf("invalid upload ID: %s", uploadIDStr)),
			h.logger,
		))
		return
	}

	chunkID, err := strconv.ParseUint(chunkIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, handler_err.GetErrorResponse(
			e.NewInvalidInputError(err, fmt.Sprintf("invalid chunk ID: %s", chunkIDStr)),
			h.logger,
		))
		return
	}

	// Get the file content from the request
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, handler_err.GetErrorResponse(
			e.NewInvalidInputError(err, "file is required"),
			h.logger,
		))
		return
	}

	// Open the file // TODO: can be at the usecase layer?
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, handler_err.GetErrorResponse(
			e.NewFileStorageError(err, "failed to open file"),
			h.logger,
		))
		return
	}
	defer src.Close()

	// Upload the chunk
	ucErr := h.fileUploadUseCase.Execute(ctx.Request.Context(), uploadID, uint(chunkID), src)
	if ucErr != nil {
		ctx.JSON(ucErr.StatusCode(), handler_err.GetErrorResponse(ucErr, h.logger))
		return
	}

	ctx.JSON(http.StatusOK, UploadResponse{Status: "OK"})
}
