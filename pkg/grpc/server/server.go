package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/pkg/http/probes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Handler interface {
	Register(s *grpc.Server, logger logr.Logger) error
}

type Server struct {
	S             *grpc.Server
	logger        logr.Logger
	port          uint
	registerFuncs []Handler
}

func New(logger logr.Logger, port uint, registers ...Handler) (*Server, error) {
	if len(registers) == 0 {
		return nil, fmt.Errorf("grpc.New requires at least one RegisterHandler")
	}
	s := &Server{
		logger:        logger,
		port:          port,
		registerFuncs: registers,
		S: grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
		})),
	}

	reflection.Register(s.S)
	return s, nil
}

// Serve initiates the plugin.
func (s *Server) Serve(ctx context.Context, r probes.Reporter) error {
	for _, h := range s.registerFuncs {
		if err := h.Register(s.S, s.logger); err != nil {
			return fmt.Errorf("failed to register gRPC handler: %w", err)
		}
	}

	errCh := make(chan error)

	s.logger.Info("Starting tcp listener", "port", s.port)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		s.logger.Info("Starting gRPC server")
		err := s.S.Serve(lis)
		s.logger.Info("gRPC server stopped", "err", err)

		r.NotReady()

		errCh <- err
	}()

	r.Ready()

	select {
	case <-ctx.Done():
		s.S.GracefulStop()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
