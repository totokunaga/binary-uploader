package handler

import (
	"github.com/gin-gonic/gin"
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"golang.org/x/exp/slog"
)

type ErrorResponse struct {
	Error      string `json:"error"`
	ErrorCode  string `json:"code"`
	StatusCode int    `json:"status_code"`
}

func sendErrorResponse(ctx *gin.Context, logger *slog.Logger, err e.CustomError) {
	logger.Error("error",
		"error", err.Error(),
		"code", err.ErrorCode(),
		"status_code", err.StatusCode(),
	)

	errResp := ErrorResponse{
		Error:      err.Error(),
		ErrorCode:  err.ErrorCode(),
		StatusCode: err.StatusCode(),
	}

	ctx.JSON(errResp.StatusCode, errResp)
}
