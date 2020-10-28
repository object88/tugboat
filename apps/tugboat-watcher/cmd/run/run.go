package run

import (
	"context"

	"github.com/object88/tugboat/apps/tugboat-watcher/pkg/watcher"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/object88/tugboat/pkg/k8s/cliflags"
	k8scliflags "github.com/object88/tugboat/pkg/k8s/cliflags"
	"github.com/object88/tugboat/pkg/k8s/watchers"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	httpFlagMgr *httpcliflags.FlagManager
	k8sFlagMgr  *k8scliflags.FlagManager

	w watchers.Watcher
}

// CreateCommand returns the `run` Command
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
		CommonArgs:  ca,
		httpFlagMgr: httpcliflags.New(),
		k8sFlagMgr:  cliflags.New(),
	}

	flags := c.Flags()

	c.httpFlagMgr.ConfigureHttpFlag(flags)
	c.k8sFlagMgr.ConfigureKubernetesConfig(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	conf, err := c.k8sFlagMgr.KubernetesConfig().ToRESTConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return err
	}

	c.w = watcher.New(c.Log, clientset)

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	f0 := func(ctx context.Context) error {
		m, err := router.New(c.Log).Route(router.LoggingDefaultRoute, router.Defaults())
		if err != nil {
			return err
		}

		http.New(c.Log, m, c.httpFlagMgr.HttpPort()).Serve(ctx)
		return nil
	}

	f1 := func(ctx context.Context) error {
		wm := watchers.New(c.Log)
		return wm.Run(ctx, c.w)
	}

	return common.Multiblock(c.Log, f0, f1)
}
