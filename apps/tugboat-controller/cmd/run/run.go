package run

import (
	"context"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/watcher"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/pkg/k8s/cliflags"
	"github.com/object88/tugboat/pkg/k8s/watchers"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	k8sFlagMgr *cliflags.FlagManager

	w watchers.Watcher
}

func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c command
	c = command{
		Command: cobra.Command{
			Use:   "run",
			Short: "run observes the state of tugboat.lauches",
			Args:  cobra.NoArgs,
			PreRunE: func(cmd *cobra.Command, args []string) error {
				return c.preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.execute(cmd, args)
			},
		},
		CommonArgs: ca,
		k8sFlagMgr: cliflags.New(),
	}

	flags := c.Flags()

	c.k8sFlagMgr.ConfigureKubernetesConfig(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	conf, err := c.k8sFlagMgr.KubernetesConfig().ToRESTConfig()
	if err != nil {
		return err
	}
	clientset, err := versioned.NewForConfig(conf)
	if err != nil {
		return err
	}

	c.w = watcher.New(clientset)

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	wm := watchers.New()

	return common.Block(func(ctx context.Context) error {
		return wm.Run(ctx, c.w)
	})
}
