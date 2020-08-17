package logging

import (
	"context"

	"github.com/sirupsen/logrus"
)

type loggerKey string

const (
	logger = loggerKey("logger")
)

func FromContext(ctx context.Context) *logrus.Logger {
	x, ok := ctx.Value(logger).(*logrus.Logger)
	if !ok {
		return nil
	}
	return x
}

func ToContext(ctx context.Context, log *logrus.Logger) context.Context {
	return context.WithValue(ctx, logger, log)
}
