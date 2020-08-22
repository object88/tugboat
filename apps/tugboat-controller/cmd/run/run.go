package run

import (
	"context"
	"strings"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/controller"
	helmcliflags "github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cliflags"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/watcher"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/object88/tugboat/pkg/k8s/watchers"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	helmFlagMgr *helmcliflags.FlagManager
	httpFlagMgr *httpcliflags.FlagManager
	// k8sFlagMgr  *cliflags.FlagManager

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
		helmFlagMgr: helmcliflags.New(),
		httpFlagMgr: httpcliflags.New(),
		// k8sFlagMgr:  cliflags.New(),
	}

	flags := c.Flags()

	c.helmFlagMgr.ConfigureFlags(flags)
	c.httpFlagMgr.ConfigurePortFlag(flags)
	// c.k8sFlagMgr.ConfigureKubernetesConfig(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	conf, err := c.helmFlagMgr.Client().ToRESTConfig()
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
	f0 := func(ctx context.Context) error {
		m, err := router.New(c.Log).Route(router.Defaults())
		if err != nil {
			return err
		}

		http.New(c.Log, m, c.httpFlagMgr.Port()).Serve(ctx)
		return nil
	}

	namespace := "default"
	f1 := func(ctx context.Context) error {
		// return watchers.New().Run(ctx, c.w)
		// Set default manager options
		options := manager.Options{
			Namespace: namespace,
		}

		// Add support for MultiNamespace set in WATCH_NAMESPACE (e.g ns1,ns2)
		// Note that this is not intended to be used for excluding namespaces, this is better done via a Predicate
		// Also note that you may face performance issues when using this with a high number of namespaces.
		// More Info: https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg/cache#MultiNamespacedCacheBuilder
		if strings.Contains(namespace, ",") {
			options.Namespace = ""
			options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(namespace, ","))
		}

		// Create a new manager to provide shared dependencies and start components
		cfg, err := c.helmFlagMgr.Client().ToRESTConfig()
		if err != nil {
			return err
		}
		mgr, err := manager.New(cfg, options)
		if err != nil {
			return err
		}

		c.Log.Info("Registering Components.")

		// Setup Scheme for all resources
		if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
			return err
		}

		// Setup all Controllers
		if err := controller.AddToManager(mgr); err != nil {
			return err
		}

		// And now, run.  And wait.
		return mgr.Start(ctx.Done())
	}

	err := common.Multiblock(c.Log, f0, f1)

	return err
}
