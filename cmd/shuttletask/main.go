package main

import (
	"log"

	"github.com/kjuulh/shuttletask/cmd/shuttletask/commands"
	"github.com/spf13/cobra"
)

func rootCmd() *cobra.Command {
	return &cobra.Command{Use: "shuttletask"}
}

func main() {

	rootcmd := rootCmd()

	rootcmd.AddCommand(
		commands.CompileCommand(),
		commands.RunCommand(),
		commands.ListCommand(),
	)

	if err := rootcmd.Execute(); err != nil {
		log.Fatalf("command failed: %v", err)
	}
}
