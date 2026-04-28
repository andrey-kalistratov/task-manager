package unix

import "encoding/json"

type Request struct {
	Command string          `json:"command"`
	Body    json.RawMessage `json:"body,omitempty"`
}

type Response struct {
	Body  json.RawMessage `json:"body,omitempty"`
	Error string          `json:"error,omitempty"`
}
