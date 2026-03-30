package ipc

import (
	"task-manager/planner/internal/task"
	"task-manager/planner/unixsocket"
)

type AddHandler struct {
	storage task.Storage
}

func NewAddHandler(storage task.Storage) *AddHandler {
	return &AddHandler{storage: storage}
}

func (h *AddHandler) ServeIPC(request unixsocket.Request) unixsocket.Response {
	return unixsocket.Response{}
}
