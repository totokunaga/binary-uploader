package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"golang.org/x/exp/slog"
)

type FileListHandler struct {
	fileListUseCase usecase.FileListUseCase
	logger          *slog.Logger
}

type FileListResponse struct {
	Files []string `json:"files"`
}

func NewFileListHandler(fileListUseCase usecase.FileListUseCase, logger *slog.Logger) *FileListHandler {
	return &FileListHandler{
		fileListUseCase: fileListUseCase,
		logger:          logger,
	}
}

func (h *FileListHandler) Execute(ctx *gin.Context) {
	files, err := h.fileListUseCase.Execute(ctx.Request.Context())
	if err != nil {
		ctx.JSON(err.StatusCode(), getErrorResponse(err, h.logger))
		return
	}

	ctx.JSON(http.StatusOK, FileListResponse{Files: files})
}
