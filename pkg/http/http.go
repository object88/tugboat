package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/pkg/http/probes"
)

type Server struct {
	logger logr.Logger
	srv    *http.Server

	tlsSrv *http.Server
}

func New(logger logr.Logger, routes http.Handler, port int) *Server {
	return &Server{
		logger: logger,
		srv: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: routes,
		},
	}
}

func (s *Server) ConfigureTLS(port int, certFile string, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	s.tlsSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.srv.Handler,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
		},
	}
	return nil
}

func (s *Server) Serve(ctx context.Context, r probes.Reporter) {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(err, "listen completed")
		}
	}()

	if s.tlsSrv != nil {
		go func() {
			if err := s.tlsSrv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				s.logger.Error(err, "listen on TLS completed")
			}
		}()
	}

	r.Ready()

	s.logger.Info("Server Started")

	// Wait for the context to wrap up.
	select {
	case <-ctx.Done():
		// Context has been ended; take down the readiness probe
		r.NotReady()

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
		}()

		if err := s.srv.Shutdown(timeoutCtx); err != nil {
			s.logger.Error(err, "Server shutdown failed")
		}
	}
}
