package main

import (
	"github.com/spf13/cobra"
)

func initRootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "what",
		Short: "A task manager inside your CLI!",
	}

	cmd.AddCommand(
		nextCmd,
		nowCmd,
	)
	return cmd, nil
}
