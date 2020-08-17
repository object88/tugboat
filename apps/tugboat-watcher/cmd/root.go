package cmd

import (
	"strings"
	"time"

	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/internal/cmd/completion"
	"github.com/object88/tugboat/internal/cmd/version"
	"github.com/spf13/cobra"
)

// InitializeCommands sets up the cobra commands
func InitializeCommands() *cobra.Command {
	ca, rootCmd := createRootCommand()

	rootCmd.AddCommand(
		completion.CreateCommand(ca),
		// run.CreateCommand(ca),
		version.CreateCommand(ca),
	)

	return rootCmd
}

func createRootCommand() (*common.CommonArgs, *cobra.Command) {
	ca := common.NewCommonArgs()

	var start time.Time
	cmd := &cobra.Command{
		Use:   "tugboat-watcher",
		Short: "tugboat-watcher reports on activities of a deployment",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			start = time.Now()
			ca.Evaluate()

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
		PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
			duration := time.Since(start)

			segments := []string{}
			var f func(c1 *cobra.Command)
			f = func(c1 *cobra.Command) {
				parent := c1.Parent()
				if parent != nil {
					f(parent)
				}
				segments = append(segments, c1.Name())
			}
			f(cmd)

			ca.Log.Infof("Executed command \"%s\" in %s", strings.Join(segments, " "), duration)
			return nil
		},
	}

	flags := cmd.PersistentFlags()
	ca.Setup(flags)

	return ca, common.TraverseRunHooks(cmd)
}
