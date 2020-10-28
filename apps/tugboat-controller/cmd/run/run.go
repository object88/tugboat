package run

import (
	"context"
	"fmt"
	"strings"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	v1 "github.com/object88/tugboat/apps/tugboat-controller/pkg/http/router/v1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/validator"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/watcher"
	"github.com/object88/tugboat/internal/cmd/common"
	notificationsclient "github.com/object88/tugboat/internal/notifications/client"
	notificationscliflags "github.com/object88/tugboat/internal/notifications/cliflags"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/router"
	k8scliflags "github.com/object88/tugboat/pkg/k8s/cliflags"
	"github.com/object88/tugboat/pkg/k8s/watchers"
	"github.com/spf13/cobra"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	mgr    manager.Manager
	scheme *runtime.Scheme
	w      watchers.Watcher

	httpFlagMgr          *httpcliflags.FlagManager
	notificationsFlagMgr *notificationscliflags.FlagManager
	k8sFlagMgr           *k8scliflags.FlagManager
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
		notificationsFlagMgr: notificationscliflags.New(),
		k8sFlagMgr:           k8scliflags.New(),
	}

	flags := c.Flags()

	c.httpFlagMgr.ConfigureHttpFlag(flags)
	c.httpFlagMgr.ConfigureHttpsFlags(flags)
	c.notificationsFlagMgr.ConfigureListenersFlag(flags)
	c.k8sFlagMgr.ConfigureKubernetesConfig(flags)

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

	namespace := "default"

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
		c.Log.Info(fmt.Sprintf("Using multi-namespace: %s", namespace))
		options.Namespace = ""
		options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(namespace, ","))
	}

	c.scheme = runtime.NewScheme()
	if err = apis.AddToScheme(c.scheme); err != nil {
		return err
	}
	if err = clientgoscheme.AddToScheme(c.scheme); err != nil {
		return err
	}
	if err = apiextv1.AddToScheme(c.scheme); err != nil {
		return err
	}
	if err = apiextv1beta1.AddToScheme(c.scheme); err != nil {
		return err
	}

	c.Log.Info("Registering custom resource components.")

	cfg, err := c.k8sFlagMgr.KubernetesConfig().ToRESTConfig()
	if err != nil {
		return err
	}
	c.mgr, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: c.scheme,
		// MetricsBindAddress: metricsAddr,
		Port: 9443,
		// LeaderElection:     enableLeaderElection,
		// LeaderElectionID:   "e486e3e8.my.domain",
	})
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	versionedclientset, err := versioned.NewForConfig(cfg)
	if err != nil {
		return err
	}
	c.w = watcher.New(c.Log, clientset, versionedclientset)

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	return common.Multiblock(c.Log, c.startHTTPServer, c.startManager, c.startWatcher)
}

func (c *command) startHTTPServer(ctx context.Context) error {
	m := validator.NewMutator(c.Log, c.scheme)
	v := validator.New(c.Log, c.scheme)
	rts, err := router.New(c.Log).Route(router.LoggingDefaultRoute, router.Defaults(v1.Defaults(c.Log, m, v)))
	if err != nil {
		return err
	}

	cf, err := c.httpFlagMgr.HttpsCertFile()
	if err != nil {
		return err
	}
	kf, err := c.httpFlagMgr.HttpsKeyFile()
	if err != nil {
		return err
	}

	h := http.New(c.Log, rts, c.httpFlagMgr.HttpPort())
	if p := c.httpFlagMgr.HttpsPort(); p != 0 {
		if err = h.ConfigureTLS(p, cf, kf); err != nil {
			return err
		}
	}

	c.Log.Info("starting http")
	defer c.Log.Info("http complete")

	h.Serve(ctx)
	return nil
}

func (c *command) startManager(ctx context.Context) error {
	// And now, run.  And wait.
	c.Log.Info("starting controller manager")
	defer c.Log.Info("controller manager complete")

	return c.mgr.Start(ctx.Done())
}

func (c *command) startWatcher(ctx context.Context) error {
	c.Log.Info("starting watcher")
	defer c.Log.Info("watcher complete")

	wm := watchers.New(c.Log)
	return wm.Run(ctx, c.w)
}
