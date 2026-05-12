package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
)

type Client struct {
	cli *http.Client
}

func NewClient() *Client {
	return &Client{
		cli: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					var d net.Dialer
					return d.DialContext(ctx, "unix", config.UnixSocket)
				},
			},
		},
	}
}

func (c *Client) RunTask(ctx context.Context, req *RunRequest) (*RunResponse, error) {
	var resp RunResponse
	if err := c.do(ctx, runPath, req, &resp); err != nil {
		if apiErr, ok := errors.AsType[*APIError](err); ok {
			return nil, apiErr
		}
		return nil, errors.New("internal error")
	}
	return &resp, nil
}

func (c *Client) do(ctx context.Context, path string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix"+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.cli.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseError(resp)
	}

	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func parseError(resp *http.Response) error {
	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read error from response: %w", err)
	}

	return &APIError{Message: string(bytes.TrimSpace(msg))}
}
