package commands

import (
	"log"

	"github.com/kjuulh/shuttletask/pkg/compile"
	"github.com/kjuulh/shuttletask/pkg/discover"
	"github.com/spf13/cobra"
)

func CompileCommand() *cobra.Command {
	return &cobra.Command{
		Use: "compile",
		RunE: func(cmd *cobra.Command, args []string) error {
			println("running: compile")

			ctx := cmd.Context()

			disc, err := discover.Discover(ctx, "shuttle.yaml")
			if err != nil {
				return err
			}

			path, err := compile.Compile(ctx, disc)
			if err != nil {
				return err
			}

			log.Printf("compiled: %s", path)

			return nil
		},
	}
}
