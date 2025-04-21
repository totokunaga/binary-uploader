package error

import (
	e "github.com/tomoya.tokunaga/server/internal/domain/entity/error"
	"golang.org/x/exp/slog"
)

type ErrorResponse struct {
	Error      string `json:"error"`
	ErrorCode  string `json:"code"`
	StatusCode int    `json:"status_code"`
}

func GetErrorResponse(err e.CustomError, logger *slog.Logger) ErrorResponse {
	logger.Error("error",
		"error", err.Error(),
		"code", err.ErrorCode(),
		"status_code", err.StatusCode(),
	)

	return ErrorResponse{
		Error:      err.Error(),
		ErrorCode:  err.ErrorCode(),
		StatusCode: err.StatusCode(),
	}
}
