package main

import (
	"fmt"
	"os"

	"task-manager/planner/cmd"
	"task-manager/planner/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		fmt.Printf("failed to start tool: %v\n", err)
		os.Exit(1)
	}

	err = cmd.NewRootCmd(application).Execute()

	if closeErr := application.Close(); closeErr != nil {
		fmt.Printf("failed to cleanup tool: %v\n", closeErr)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
