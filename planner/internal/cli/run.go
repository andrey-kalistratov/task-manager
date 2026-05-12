package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/andrey-kalistratov/task-manager/planner/internal/task/ipc"
)

// NewRunCmd creates a *cobra.Command representing the CLI subcommand `tm run`.
func NewRunCmd() *cobra.Command {
	var req ipc.RunRequest

	cmd := &cobra.Command{
		Use:   "run <cmd>",
		Short: "run a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req.Command = args[0]

			cli := ipc.NewClient()

			resp, err := cli.RunTask(context.Background(), &req)
			if err != nil {
				return err
			}

			cmd.Printf("Run task: %s\n", resp.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Name, "name", "", "task name")
	cmd.Flags().StringVar(&req.Image, "image", "", "docker image")
	cmd.Flags().StringToStringVar(
		&req.Inputs, "in", nil, "input params (key=value,...)",
	)
	cmd.Flags().StringToStringVar(
		&req.Outputs, "out", nil, "output params (key=value,...)",
	)

	return cmd
}
