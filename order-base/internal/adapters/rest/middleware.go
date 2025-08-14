package rest

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/Util787/order-base/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const expectedDurationMs = 2000

func NewBasicMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.NewString()
		op := strings.TrimSuffix(c.HandlerName(), "-fm")

		c.Set("op", op)
		// gin only allows to use this way or clone request
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), common.ContextKey("request_id"), requestId))

		log := log.With(slog.String("op", op), slog.String("request_id", requestId))

		log.Info("Request received", slog.String("ip", c.ClientIP()), slog.String("user_agent", c.GetHeader("User-Agent")), slog.String("path", c.FullPath()))

		start := time.Now()

		c.Next()

		durationMs := time.Since(start).Milliseconds()
		log.Debug("Request finished", slog.Int64("duration_ms", durationMs), slog.Int("status", c.Writer.Status()))
		if durationMs > expectedDurationMs {
			log.Warn("Operation is taking more time than expected", slog.Int("expected_duration(ms)", expectedDurationMs), slog.Int64("actual_duration(ms)", durationMs))
		}

	}
}
