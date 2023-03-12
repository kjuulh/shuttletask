package cmder

import (
	"log"
	"reflect"

	"github.com/spf13/cobra"
)

type RootCmd struct {
	Cmds []*Cmd
}

func NewRoot() *RootCmd {
	return &RootCmd{}
}

func (rc *RootCmd) AddCmds(cmd ...*Cmd) *RootCmd {
	rc.Cmds = append(rc.Cmds, cmd...)

	return rc
}

func (rc *RootCmd) Execute() {
	rootcmd := &cobra.Command{Use: "shuttletask"}

	for _, cmd := range rc.Cmds {
		parameters := make([]string, len(cmd.Args))

		cobracmd := &cobra.Command{
			Use: cmd.Name,
			RunE: func(cobracmd *cobra.Command, args []string) error {
				if err := cobracmd.ParseFlags(args); err != nil {
					return err
				}

				inputs := make([]reflect.Value, len(cmd.Args))
				for i, arg := range parameters {
					inputs[i] = reflect.ValueOf(arg)
				}

				reflect.
					ValueOf(cmd.Func).
					Call(inputs)
				return nil
			},
		}
		for i, arg := range cmd.Args {
			cobracmd.Flags().StringVar(&parameters[i], arg.Name, "", "")
			_ = cobracmd.MarkFlagRequired(arg.Name)
		}

		rootcmd.AddCommand(cobracmd)
	}

	if err := rootcmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
}

type Arg struct {
	Name string
}

type Cmd struct {
	Name string
	Func any
	Args []Arg
}

func NewCmd(name string, f any) *Cmd {
	return &Cmd{
		Name: name,
		Func: f,
		Args: []Arg{},
	}
}

func WithArgs(cmd *Cmd, argName string) *Cmd {
	cmd.Args = append(cmd.Args, Arg{Name: argName})
	return cmd
}
