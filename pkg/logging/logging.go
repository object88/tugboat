package logging

import (
	"context"

	"github.com/go-logr/logr"
)

type loggerKey string

const (
	logger = loggerKey("logger")
)

func FromContext(ctx context.Context) logr.Logger {
	x, ok := ctx.Value(logger).(logr.Logger)
	if !ok {
		return nil
	}
	return x
}

func ToContext(ctx context.Context, log logr.Logger) context.Context {
	return context.WithValue(ctx, logger, log)
}
