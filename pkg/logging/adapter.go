package logging

import (
	"bufio"
	"fmt"
	"io"

	"github.com/go-logr/logr"
)

type Adapter struct {
	Log logr.Logger
}

func (a *Adapter) Logf(msg string, v ...interface{}) {
	a.Log.Info(fmt.Sprintf(msg, v...))
}

type Writer struct {
	Log logr.Logger
}

func (w *Writer) Out() io.Writer {
	pr, pw := io.Pipe()
	go func() {
		s := bufio.NewScanner(pr)
		for s.Scan() {
			w.Log.Info(s.Text())
		}
	}()

	return pw
}
