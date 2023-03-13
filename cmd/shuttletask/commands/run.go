package commands

import (
	"github.com/kjuulh/shuttletask/pkg/executer"
	"github.com/spf13/cobra"
)

func RunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{"r"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := executer.Run(ctx, "shuttle.yaml", args...); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().SetInterspersed(false)

	return cmd
}
