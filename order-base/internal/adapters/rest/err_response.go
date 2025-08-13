package rest

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, log *slog.Logger, statusCode int, message string, err error) {
	log.Error(message, slog.String("error", err.Error()))
	c.AbortWithStatusJSON(statusCode, errorResponse{message})
}
