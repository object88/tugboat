package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

type Server struct {
	logger logr.Logger
	srv    *http.Server
}

func New(logger logr.Logger, routes http.Handler, port int) *Server {
	addr := fmt.Sprintf(":%d", port)
	return &Server{
		logger: logger,
		srv: &http.Server{
			Addr:    addr,
			Handler: routes,
		},
	}
}

func (s *Server) Serve(ctx context.Context) {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(err, "listen completed")
		}
	}()

	s.logger.Info("Server Started")

	// Wait for the context to wrap up.
	select {
	case <-ctx.Done():
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
		}()

		if err := s.srv.Shutdown(timeoutCtx); err != nil {
			s.logger.Error(err, "Server shutdown failed")
		}
	}
}
