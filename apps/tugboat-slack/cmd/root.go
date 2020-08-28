package cmd

import (
	"time"

	"github.com/object88/tugboat/apps/tugboat-slack/cmd/serve"
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
		serve.CreateCommand(ca),
		version.CreateCommand(ca),
	)

	return rootCmd
}

func createRootCommand() (*common.CommonArgs, *cobra.Command) {
	ca := common.NewCommonArgs()

	var start time.Time
	cmd := &cobra.Command{
		Use:   "tugboat-slack",
		Short: "tugboat-slack connects tugboat to slack",
		// BashCompletionFunction: bashCompletionFunc,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			start = time.Now()
			ca.Evaluate()

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
		PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
			ca.ReportDuration(cmd, start)
			return nil
		},
	}

	flags := cmd.PersistentFlags()
	ca.Setup(flags)

	return ca, common.TraverseRunHooks(cmd)
}
