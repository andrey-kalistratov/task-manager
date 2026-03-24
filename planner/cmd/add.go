package cmd

import (
	"task-manager/planner/internal/app"

	"github.com/spf13/cobra"
)

// NewAddCmd creates a *cobra.Command representing the CLI subcommand `tm add`.
func NewAddCmd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "add a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.AddUC.Execute()
		},
	}
}
