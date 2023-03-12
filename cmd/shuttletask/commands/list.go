package commands

import (
	"log"
	"os/exec"

	"github.com/kjuulh/shuttletask/pkg/compile"
	"github.com/kjuulh/shuttletask/pkg/discover"
	"github.com/spf13/cobra"
)

func ListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			disc, err := discover.Discover(ctx, "shuttle.yaml")
			if err != nil {
				return err
			}

			binary, err := compile.Compile(ctx, disc)
			if err != nil {
				return err
			}

			execmd := exec.Command(binary, "ls")
			output, err := execmd.CombinedOutput()
			log.Printf("%s\n", string(output))
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	cmd.Flags().SetInterspersed(false)

	return cmd
}
