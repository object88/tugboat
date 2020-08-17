package version

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/object88/tugboat/internal/cmd/cliflags"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/pkg/version"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	output cliflags.Output
}

// CreateCommand returns the version command
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:   "version",
			Short: "report the version of the tool",
			Args:  cobra.NoArgs,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				return c.Preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.Execute(cmd, args)
			},
		},
		CommonArgs: ca,
	}

	flags := c.Flags()

	c.FlagMgr.ConfigureOutputFlag(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) Preexecute(cmd *cobra.Command, args []string) error {
	c.output = c.FlagMgr.Output()

	return nil
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {
	var v version.Version

	switch c.output {
	case cliflags.Text:
		os.Stdout.WriteString(v.String())
	case cliflags.JSON:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		err := enc.Encode(v)
		if err != nil {
			return fmt.Errorf("internal error: failed to encode version: %w", err)
		}
	case cliflags.JSONCompact:
		enc := json.NewEncoder(os.Stdout)
		err := enc.Encode(v)
		if err != nil {
			return fmt.Errorf("internal error: failed to encode version: %w", err)
		}
	}

	return nil
}
