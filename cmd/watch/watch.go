package watch

import (
	"github.com/object88/tugboat/cmd/flags"
	"github.com/object88/tugboat/cmd/traverse"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type command struct {
	cobra.Command

	cflags *genericclioptions.ConfigFlags

	output flags.Output
}

// CreateCommand returns the watch command
func CreateCommand() *cobra.Command {
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:   "watch",
			Short: "report the version of the tool",
			Args:  cobra.NoArgs,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				return c.Preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.Execute(cmd, args)
			},
		},
	}

	flgs := c.Flags()

	flags.CreateOutputFlag(flgs)

	c.cflags = genericclioptions.NewConfigFlags(false)
	c.cflags.AddFlags(flgs)

	return traverse.TraverseRunHooks(&c.Command)
}

func (c *command) Preexecute(cmd *cobra.Command, args []string) error {
	var err error
	c.output, err = flags.ReadOutputFlag()
	if err != nil {
		return err
	}

	return nil
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {

	return nil
}
