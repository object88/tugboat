package cmd

import (
	"strings"
	"time"

	"github.com/object88/tugboat/cmd/common"
	"github.com/object88/tugboat/cmd/completion"
	"github.com/object88/tugboat/cmd/traverse"
	"github.com/object88/tugboat/cmd/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const bashCompletionFunc = `
__tugboat_get_outputs()
{
	COMPREPLY=( "json", "json-compressed", "text" )
}
`

// InitializeCommands sets up the cobra commands
func InitializeCommands() *cobra.Command {
	ca, rootCmd := createRootCommand()

	rootCmd.AddCommand(
		completion.CreateCommand(ca),
		version.CreateCommand(),
	)

	return rootCmd
}

func createRootCommand() (*common.CommonArgs, *cobra.Command) {
	ca := common.NewCommonArgs()

	var start time.Time
	cmd := &cobra.Command{
		Use:                    "tugboat",
		Short:                  "tugboat monitors a helm installation, upgrade, or deletion",
		BashCompletionFunction: bashCompletionFunc,
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

			ca.Logger.Infof("Executed command \"%s\" in %s", strings.Join(segments, " "), duration)
			return nil
		},
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("TUGBOAT")

	flags := cmd.PersistentFlags()
	ca.Setup(flags)

	return ca, traverse.TraverseRunHooks(cmd)
}
