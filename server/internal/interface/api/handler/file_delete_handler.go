package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"github.com/tomoya.tokunaga/server/internal/usecase"
	"golang.org/x/exp/slog"
)

type FileDeleteHandler struct {
	fileDeleteUseCase usecase.FileDeleteUseCase
	logger            *slog.Logger
}

type FileDeleteResponse struct {
	Status string `json:"status"`
}

func NewFileDeleteHandler(fileDeleteUseCase usecase.FileDeleteUseCase, logger *slog.Logger) *FileDeleteHandler {
	return &FileDeleteHandler{
		fileDeleteUseCase: fileDeleteUseCase,
		logger:            logger,
	}
}

// Execute handles file deletion
func (h *FileDeleteHandler) Execute(ctx *gin.Context) {
	// Get the file name from the URL
	fileName := ctx.Param("file_name")
	if fileName == "" {
		ctx.JSON(http.StatusBadRequest, getErrorResponse(
			e.NewInvalidInputError(errors.New("file_name parameter is required"), ""),
			h.logger,
		))
		return
	}

	// Delete the file
	err := h.fileDeleteUseCase.Execute(ctx.Request.Context(), fileName)
	if err != nil {
		ctx.JSON(err.StatusCode(), getErrorResponse(err, h.logger))
		return
	}

	ctx.JSON(http.StatusOK, FileDeleteResponse{Status: "OK"})
}
