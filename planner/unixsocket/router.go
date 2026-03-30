package unixsocket

type Router struct {
	handlers map[string]Handler
}

func NewRouter() *Router {
	return &Router{handlers: make(map[string]Handler)}
}

func (r *Router) Register(command string, h Handler) {
	r.handlers[command] = h
}

func (r *Router) ServeIPC(req Request) Response {
	h, ok := r.handlers[req.Command]
	if !ok {
		return Response{Error: "unknown command: " + req.Command}
	}
	return h.ServeIPC(req)
}
