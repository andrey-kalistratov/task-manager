package unix

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type Handler interface {
	ServeIPC(ctx context.Context, req *Request) Response
}

type Server struct {
	socket   string
	handler  Handler
	logger   *slog.Logger
	listener net.Listener
	wg       sync.WaitGroup
}

func NewServer(socket string, handler Handler, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		socket:  socket,
		handler: handler,
		logger:  logger,
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	// Create socket directory if not present.
	if err := os.MkdirAll(filepath.Dir(s.socket), 0755); err != nil {
		return fmt.Errorf("create unix socket dir: %w", err)
	}

	// Remove old socket if present.
	_ = os.Remove(s.socket)

	ln, err := net.Listen("unix", s.socket)
	if err != nil {
		return err
	}
	s.listener = ln

	for {
		conn, err := ln.Accept()
		switch {
		case ctx.Err() != nil:
			return nil
		case err != nil:
			return err
		}
		s.wg.Go(func() {
			s.handle(ctx, conn)
		})
	}
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Error("failed to close connection", "error", err)
		}
	}()

	var (
		req  Request
		resp Response
	)
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		s.logger.Error("failed to decode request", "error", err)
		resp = Response{Error: err.Error()}
	} else {
		resp = s.handler.ServeIPC(ctx, &req)
	}
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		s.logger.Error("failed to encode response", "error", err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.listener.Close(); err != nil {
		s.logger.Error("failed to close socket", "error", err)
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
