package ipc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
)

type Server struct {
	server *http.Server
}

func NewServer(runner Runner, logger *slog.Logger) (*Server, error) {
	mux := http.NewServeMux()

	mux.Handle(runPath, newRunHandler(runner, logger.With("handler", "run")))

	server := &http.Server{Handler: mux}

	return &Server{server: server}, nil
}

func (s *Server) Serve() error {
	if err := os.MkdirAll(filepath.Dir(config.UnixSocket), 0700); err != nil {
		return fmt.Errorf("create unix socket dir: %w", err)
	}

	_ = os.Remove(config.UnixSocket)

	ln, err := net.Listen("unix", config.UnixSocket)
	if err != nil {
		return fmt.Errorf("listen unix socket: %w", err)
	}

	return s.server.Serve(ln)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
