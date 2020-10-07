package testlogger

import (
	"testing"

	"github.com/go-logr/logr"
)

// TestLogger is a logr.Logger that prints through a testing.T object.
// Only error logs will have any effect.
type TestLogger struct {
	T *testing.T
}

var _ logr.Logger = TestLogger{}

func (tl TestLogger) Info(msg string, args ...interface{}) {
	tl.T.Logf("%s: %v", msg, args)
}

func (TestLogger) Enabled() bool {
	return true
}

func (tl TestLogger) Error(err error, msg string, args ...interface{}) {
	tl.T.Errorf("%s: %v -- %v", msg, err, args)
}

func (tl TestLogger) V(v int) logr.Logger {
	return tl
}

func (tl TestLogger) WithName(_ string) logr.Logger {
	return tl
}

func (tl TestLogger) WithValues(_ ...interface{}) logr.Logger {
	return tl
}
