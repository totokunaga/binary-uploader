package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"

	"github.com/tomoya.tokunaga/server/internal/usecase"
)

// FileHandler handles HTTP requests for file operations
type FileHandler struct {
	fileUseCase usecase.FileUseCase
	logger      *slog.Logger
}

// NewFileHandler creates a new FileHandler instance
func NewFileHandler(fileUseCase usecase.FileUseCase, logger *slog.Logger) *FileHandler {
	return &FileHandler{
		fileUseCase: fileUseCase,
		logger:      logger,
	}
}

// InitUpload handles file upload initialization
func (h *FileHandler) InitUpload(c *gin.Context) {
	// Get the file name from the URL
	fileName := c.Param("file_name")
	if fileName == "" {
		h.logger.Error("file name is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "file name is required"})
		return
	}

	// Parse the request body
	var req struct {
		TotalSize   uint64 `json:"total_size" binding:"required"`
		TotalChunks uint   `json:"total_chunks" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Initialize the upload
	uploadID, err := h.fileUseCase.InitUpload(c.Request.Context(), fileName, req.TotalSize, req.TotalChunks)
	if err != nil {
		h.logger.Error("failed to initialize upload", "error", err)
		switch err {
		case usecase.ErrFileAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "file already exists"})
		case usecase.ErrInvalidFileName:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file name"})
		case usecase.ErrFileSizeTooLarge:
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file size too large"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize upload"})
		}
		return
	}

	// Return the upload ID
	c.JSON(http.StatusOK, gin.H{"upload_id": uploadID})
}

// UploadChunk handles file chunk upload
func (h *FileHandler) UploadChunk(c *gin.Context) {
	// Get the upload ID and chunk ID from the URL
	uploadIDStr := c.Param("upload_id")
	chunkIDStr := c.Param("chunk_id")

	uploadID, err := strconv.ParseUint(uploadIDStr, 10, 64)
	if err != nil {
		h.logger.Error("invalid upload ID", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload ID"})
		return
	}

	chunkID, err := strconv.ParseUint(chunkIDStr, 10, 64)
	if err != nil {
		h.logger.Error("invalid chunk ID", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chunk ID"})
		return
	}

	// Get the file content from the request
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Error("failed to get file from request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		h.logger.Error("failed to open file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	// Upload the chunk
	err = h.fileUseCase.UploadChunk(c.Request.Context(), uploadID, uint(chunkID), src)
	if err != nil {
		h.logger.Error("failed to upload chunk", "error", err)
		switch err {
		case usecase.ErrInvalidUploadID:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid upload ID"})
		case usecase.ErrInvalidChunkID:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chunk ID"})
		case usecase.ErrChunkWriteFailed:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write chunk"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to upload chunk: %v", err)})
		}
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// DeleteFile handles file deletion
func (h *FileHandler) DeleteFile(c *gin.Context) {
	// Get the file name from the URL
	fileName := c.Param("file_name")
	if fileName == "" {
		h.logger.Error("file name is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "file name is required"})
		return
	}

	// Delete the file
	err := h.fileUseCase.DeleteFile(c.Request.Context(), fileName)
	if err != nil {
		h.logger.Error("failed to delete file", "error", err)
		switch err {
		case usecase.ErrFileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		}
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListFiles handles file listing
func (h *FileHandler) ListFiles(c *gin.Context) {
	// List all files
	files, err := h.fileUseCase.ListFiles(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list files", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}

	// Return the file list
	c.JSON(http.StatusOK, gin.H{"files": files})
}
