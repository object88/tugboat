package bash

import (
	"fmt"
	"os"

	"github.com/object88/tugboat/internal/cmd/common"
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
			Use:   "bash",
			Short: "installs bash shell completion",
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.Execute(cmd, args)
			},
		},
		CommonArgs: ca,
	}

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {
	err := c.Root().GenBashCompletion(os.Stdout)
	if err != nil {
		return fmt.Errorf("internal error: failed to generate bash command completions: %w", err)
	}
	return nil
}
