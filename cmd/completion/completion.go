package completion

import (
	"github.com/object88/tugboat/cmd/common"
	"github.com/object88/tugboat/cmd/completion/bash"
	"github.com/object88/tugboat/cmd/completion/zsh"
	"github.com/object88/tugboat/cmd/traverse"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs
}

// CreateCommand returns the intermediate 'config' subcommand
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:    "completion",
			Hidden: true,
		},
		CommonArgs: ca,
	}

	c.AddCommand(
		bash.CreateCommand(ca),
		zsh.CreateCommand(ca),
	)

	return traverse.TraverseRunHooks(&c.Command)
}
