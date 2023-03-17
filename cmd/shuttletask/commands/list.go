package commands

import (
	"github.com/kjuulh/shuttletask/pkg/executer"
	"github.com/spf13/cobra"
)

func ListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := executer.List(ctx, "shuttle.yaml"); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().SetInterspersed(false)

	return cmd
}
