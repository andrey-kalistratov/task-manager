package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates a *cobra.Command representing the base CLI cli `tm`.
func NewRootCmd() *cobra.Command {
	var root = &cobra.Command{
		Use:   "tm",
		Short: "Task Manager CLI",
	}
	root.AddCommand(NewRunCmd())
	return root
}
