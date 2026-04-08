package unixsocket

import "encoding/json"

type (
	Request struct {
		Command string          `json:"command"`
		Body    json.RawMessage `json:"body,omitempty"`
	}

	Response struct {
		Body  json.RawMessage `json:"body,omitempty"`
		Error string          `json:"error,omitempty"`
	}
)
