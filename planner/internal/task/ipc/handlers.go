package ipc

import (
	"task-manager/planner/internal/task"
	"task-manager/planner/unixsocket"
)

type AddHandler struct {
	service *task.Service
}

func NewAddHandler(service *task.Service) *AddHandler {
	return &AddHandler{service: service}
}

func (h *AddHandler) ServeIPC(request unixsocket.Request) unixsocket.Response {
	return unixsocket.Response{}
}
