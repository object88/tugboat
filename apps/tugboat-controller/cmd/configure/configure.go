package configure

import (
	"github.com/object88/tugboat/apps/tugboat-controller/cmd/configure/repositories"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs
}

func CreateCommand(cmmn *common.CommonArgs) *cobra.Command {
	var c command
	c = command{
		Command: cobra.Command{
			Use:   "configure",
			Short: "Configure various aspects of the tugboat controller",
			Args:  cobra.NoArgs,
		},
		CommonArgs: cmmn,
	}

	c.AddCommand(repositories.CreateCommand(cmmn))

	return common.TraverseRunHooks(&c.Command)
}
