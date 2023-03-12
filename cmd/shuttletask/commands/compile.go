package commands

import "github.com/spf13/cobra"

func CompileCommand() *cobra.Command {
	return &cobra.Command{
		Use: "compile",
		RunE: func(cmd *cobra.Command, args []string) error {
			println("running: compile")

			return nil
		},
	}
}
