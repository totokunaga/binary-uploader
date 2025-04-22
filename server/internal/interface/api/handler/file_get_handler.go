package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tomoya.tokunaga/server/internal/domain/entity"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"golang.org/x/exp/slog"
)

type FileGetHandler struct {
	fileGetUseCase usecase.FileGetUseCase
	config         *entity.Config
	logger         *slog.Logger
}

type FileGetResponse struct {
	Files []string `json:"files"`
}

type FileGetStatsResponse struct {
	*entity.File
	UploadTimeoutSecond time.Duration `json:"upload_timeout_second"`
}

func NewFileGetHandler(fileGetUseCase usecase.FileGetUseCase, config *entity.Config, logger *slog.Logger) *FileGetHandler {
	return &FileGetHandler{
		fileGetUseCase: fileGetUseCase,
		config:         config,
		logger:         logger,
	}
}

func (h *FileGetHandler) Execute(ctx *gin.Context) {
	files, err := h.fileGetUseCase.Execute(ctx.Request.Context())
	if err != nil {
		sendErrorResponse(ctx, h.logger, err)
		return
	}

	ctx.JSON(http.StatusOK, FileGetResponse{Files: files})
}

func (h *FileGetHandler) ExecuteGetStats(ctx *gin.Context) {
	fileName := ctx.Param("file_name")
	if fileName == "" {
		sendErrorResponse(ctx, h.logger, e.NewInvalidInputError(errors.New("file_name parameter is required"), ""))
		return
	}

	fileStats, err := h.fileGetUseCase.ExecuteGetStats(ctx.Request.Context(), fileName)
	if err != nil {
		sendErrorResponse(ctx, h.logger, err)
		return
	}
	if fileStats == nil {
		sendErrorResponse(ctx, h.logger, e.NewNotFoundError(errors.New("file not found"), ""))
		return
	}

	ctx.JSON(http.StatusOK, FileGetStatsResponse{
		File:                fileStats,
		UploadTimeoutSecond: h.config.UploadTimeoutSecond,
	})
}
