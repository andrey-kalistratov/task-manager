package unixsocket

import (
	"encoding/json"
	"log/slog"
	"net"
)

type Handler interface {
	ServeIPC(req Request) Response
}

type Server struct {
	socket  string
	handler Handler
	log     *slog.Logger
}

func NewServer(socket string, handler Handler, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{
		socket:  socket,
		handler: handler,
		log:     log,
	}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("unix", s.socket)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			s.log.Error("accept", "err", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	var (
		req  Request
		resp Response
	)
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		s.log.Error("decode request", "err", err)
		resp = Response{Error: err.Error()}
	} else {
		resp = s.handler.ServeIPC(req)
	}
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		s.log.Error("encode response", "err", err)
	}
}
