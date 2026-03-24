package cmd

import (
	"task-manager/planner/internal/app"

	"github.com/spf13/cobra"
)

// NewRootCmd creates a *cobra.Command representing the base CLI command `tm`.
func NewRootCmd(app *app.App) *cobra.Command {
	var root = &cobra.Command{
		Use:   "task",
		Short: "Task Manager CLI",
	}
	root.AddCommand(NewAddCmd(app))
	return root
}
