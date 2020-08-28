package cmd

import (
	"time"

	// "github.com/object88/tugboat/apps/cmd/common"
	// "github.com/object88/tugboat/cmd/common"
	"github.com/object88/tugboat/apps/tugboat-controller/cmd/run"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/internal/cmd/completion"
	"github.com/object88/tugboat/internal/cmd/version"
	"github.com/spf13/cobra"
	// "github.com/spf13/viper"
)

// const bashCompletionFunc = `
// __tugboat_get_outputs()
// {
// 	COMPREPLY=( "json", "json-compressed", "text" )
// }
// `

// InitializeCommands sets up the cobra commands
func InitializeCommands() *cobra.Command {
	ca, rootCmd := createRootCommand()

	rootCmd.AddCommand(
		completion.CreateCommand(ca),
		run.CreateCommand(ca),
		version.CreateCommand(ca),
	)

	return rootCmd
}

func createRootCommand() (*common.CommonArgs, *cobra.Command) {
	ca := common.NewCommonArgs()

	var start time.Time
	cmd := &cobra.Command{
		Use:   "tugboat-controller",
		Short: "tugboat-controller monitors a helm installation, upgrade, or deletion",
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
