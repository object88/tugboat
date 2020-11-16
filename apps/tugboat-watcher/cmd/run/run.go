package run

import (
	"context"
	"fmt"

	"github.com/object88/tugboat/apps/tugboat-watcher/pkg/watcher"
	"github.com/object88/tugboat/internal/cmd/common"
	notificationsclient "github.com/object88/tugboat/internal/notifications/client"
	notificationscliflags "github.com/object88/tugboat/internal/notifications/cliflags"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/probes"

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

	httpFlagMgr          *httpcliflags.FlagManager
	k8sFlagMgr           *k8scliflags.FlagManager
	notificationsFlagMgr *notificationscliflags.FlagManager

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
		CommonArgs:           ca,
		httpFlagMgr:          httpcliflags.New(),
		k8sFlagMgr:           cliflags.New(),
		notificationsFlagMgr: notificationscliflags.New(),
	}

	flags := c.Flags()

	c.httpFlagMgr.ConfigureHttpFlag(flags)
	c.k8sFlagMgr.ConfigureKubernetesConfig(flags)
	c.notificationsFlagMgr.ConfigureListenersFlag(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	targets, err := c.notificationsFlagMgr.Listeners()
	if err != nil {
		return fmt.Errorf("failed to get notification listeners: %w", err)
	}
	c.Log.Info("Listeners", "listeners", targets)
	notifier := notificationsclient.New(c.Log)
	if err := notifier.Connect(targets); err != nil {
		return fmt.Errorf("failed to establish clients for notification listeners: %w", err)
	}

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
	p := probes.New()

	f0 := func(ctx context.Context, r probes.Reporter) error {
		m, err := router.New(c.Log).Route(router.LoggingDefaultRoute, router.Defaults(p))
		if err != nil {
			return err
		}

		http.New(c.Log, m, c.httpFlagMgr.HttpPort()).Serve(ctx, r)
		return nil
	}

	f1 := func(ctx context.Context, r probes.Reporter) error {
		wm := watchers.New(c.Log)
		return wm.Run(ctx, r, c.w)
	}

	return common.Multiblock(c.Log, p, f0, f1)
}
