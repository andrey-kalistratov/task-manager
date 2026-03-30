package cli

import (
	"fmt"

	"task-manager/planner/internal/config"
	"task-manager/planner/unixsocket"

	"github.com/spf13/cobra"
)

// NewAddCmd creates a *cobra.Command representing the CLI subcommand `tm add`.
func NewAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "add a new task",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := unixsocket.NewClient(config.UnixSocket)
			if err != nil {
				return err
			}
			defer client.Close()

			resp, err := client.Do(unixsocket.Request{
				Command: "add",
			})
			if err != nil {
				return fmt.Errorf("failed to connect daemon: %v", err)
			}
			if resp.Error != "" {
				return fmt.Errorf("failed to add task: %s", resp.Error)
			}
			fmt.Println("Added task")
			return nil
		},
	}
}
