package cli

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/andrey-kalistratov/task-manager/planner/internal/config"
	"github.com/andrey-kalistratov/task-manager/planner/internal/task/ipc"
	"github.com/andrey-kalistratov/task-manager/planner/unix"

	"github.com/spf13/cobra"
)

var (
	ErrConnection      = errors.New("cannot connect to daemon")
	ErrInternal        = errors.New("internal CLI error")
	ErrRequestFailed   = errors.New("request to daemon failed")
	ErrInvalidResponse = errors.New("invalid response from daemon")
)

// NewRunCmd creates a *cobra.Command representing the CLI subcommand `tm run`.
func NewRunCmd() *cobra.Command {
	var opts ipc.RunOptions

	cmd := &cobra.Command{
		Use:   "run <cmd>",
		Short: "run a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Command = args[0]

			result, err := runTask(opts)
			if err != nil {
				return err
			}

			cmd.Printf("Task running: %s\n", result.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "task name")
	cmd.Flags().StringVar(&opts.Image, "image", "", "docker image")
	cmd.Flags().StringToStringVar(
		&opts.Inputs, "in", nil, "input params (key=value,...)",
	)
	cmd.Flags().StringToStringVar(
		&opts.Outputs, "out", nil, "output params (key=value,...)",
	)

	return cmd
}

func runTask(opts ipc.RunOptions) (*ipc.RunResult, error) {
	client, err := unix.NewClient(config.UnixSocket)
	if err != nil {
		return nil, ErrConnection
	}
	defer func() { _ = client.Close() }()

	for _, files := range []map[string]string{opts.Inputs, opts.Outputs} {
		for name, path := range files {
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, ErrInternal
			}
			files[name] = abspath
		}
	}

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, ErrInternal
	}

	resp, err := client.Do(unix.Request{
		Command: "run",
		Body:    body,
	})
	if err != nil {
		return nil, ErrRequestFailed
	}

	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}

	var result ipc.RunResult
	if err = json.Unmarshal(resp.Body, &result); err != nil {
		return nil, ErrInvalidResponse
	}
	return &result, nil
}
