package unixsocket

import (
	"encoding/json"
	"net"
)

type Client struct {
	conn net.Conn
}

func NewClient(socket string) (*Client, error) {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Do(req Request) (Response, error) {
	if err := json.NewEncoder(c.conn).Encode(req); err != nil {
		return Response{}, err
	}

	var resp Response
	err := json.NewDecoder(c.conn).Decode(&resp)
	return resp, err
}

func (c *Client) Close() error {
	return c.conn.Close()
}
