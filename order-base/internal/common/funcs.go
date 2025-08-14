package common

import (
	"context"
	"log/slog"
	"runtime"
)

// GetOperationName returns PackageName.FunctionName of the func it was called in
//
// it should be used for logging or error wrapping
func GetOperationName() string {

	function, _, _, ok := runtime.Caller(1)
	if !ok {
		return "couldnt get op name"
	}

	return runtime.FuncForPC(function).Name()
}

// Should be used in the start of every handler and usecase
func LogOpAndReqId(ctx context.Context, op string, log *slog.Logger) *slog.Logger {
	requestID := ctx.Value(ContextKey("request_id")).(string)
	return log.With(slog.String("op", op), slog.String("request_id", requestID))
}
