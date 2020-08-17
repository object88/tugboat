package zsh

import (
	"os"

	// "github.com/object88/tugboat/cmd/common"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs
}

// CreateCommand returns the 'current' subcommand
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:   "zsh",
			Short: "installs zsh shell completion",
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.Execute(cmd, args)
			},
		},
		CommonArgs: ca,
	}

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {
	err := c.Root().GenZshCompletion(os.Stdout)
	if err != nil {
		return errors.Wrapf(err, "Internal error: failed to generate zsh command completions")
	}
	return nil
}
