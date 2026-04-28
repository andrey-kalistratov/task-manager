package ipc

import (
	"context"
	"log/slog"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task"
	"github.com/andrey-kalistratov/task-manager/planner/unix"
)

var _ task.Transport = (*Server)(nil)

type Server struct {
	server *unix.Server
}

func NewServer(service task.Service, logger *slog.Logger) *Server {
	router := unix.NewRouter()

	router.Register("run", NewRunHandler(service, logger.With("handler", "run")))

	return &Server{server: unix.NewServer(config.UnixSocket, router, logger)}
}

func (s *Server) Serve(ctx context.Context) error {
	return s.server.ListenAndServe(ctx)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
