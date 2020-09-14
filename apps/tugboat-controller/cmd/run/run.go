package run

import (
	"context"
	"fmt"
	"strings"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/controller/launch"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/controller/repository"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/charts"
	chartcachecliflags "github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/charts/cliflags"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos"
	helmcliflags "github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cliflags"
	v1 "github.com/object88/tugboat/apps/tugboat-controller/pkg/http/router/v1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/validator"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	mgr    manager.Manager
	scheme *runtime.Scheme
	cc     *charts.Cache
	rc     *repos.Cache

	chartCacheFlagMgr *chartcachecliflags.FlagManager
	helmFlagMgr       *helmcliflags.FlagManager
	httpFlagMgr       *httpcliflags.FlagManager
	// k8sFlagMgr  *cliflags.FlagManager
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
		CommonArgs:        ca,
		chartCacheFlagMgr: chartcachecliflags.New(),
		helmFlagMgr:       helmcliflags.New(),
		httpFlagMgr:       httpcliflags.New(),
		// k8sFlagMgr:  cliflags.New(),
	}

	flags := c.Flags()

	c.chartCacheFlagMgr.ConfigureCacheDepthFlag(flags)
	c.chartCacheFlagMgr.ConfigureCacheDirectoryFlag(flags)
	c.helmFlagMgr.ConfigureFlags(flags)
	c.httpFlagMgr.ConfigureHttpFlag(flags)
	c.httpFlagMgr.ConfigureHttpsFlags(flags)
	// c.k8sFlagMgr.ConfigureKubernetesConfig(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	c.rc = repos.New()
	err := c.rc.Connect(
		repos.WithHelmEnvSettings(c.helmFlagMgr.EnvSettings()),
		repos.WithLogger(c.Log),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to repo cache: %w", err)
	}
	if err := c.rc.UpdateRepositories(); err != nil {
		return err
	}

	cdir, err := c.chartCacheFlagMgr.CacheDirectory()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}

	c.cc = charts.New()
	err = c.cc.Connect(
		charts.WithCacheDepth(c.chartCacheFlagMgr.CacheDepth()),
		charts.WithCacheDirectory(cdir),
		charts.WithHelmEnvSettings(c.helmFlagMgr.EnvSettings()),
		charts.WithLogger(c.Log),
		charts.WithRepoCache(c.rc),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to chart cache: %w", err)
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
	if err := apis.AddToScheme(c.scheme); err != nil {
		return err
	}

	c.Log.Info("Registering launch components.")

	cfg, err := c.helmFlagMgr.Client().ToRESTConfig()
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

	if err := (&launch.ReconcileLaunch{
		Cache:        c.cc,
		Client:       c.mgr.GetClient(),
		HelmSettings: c.helmFlagMgr.EnvSettings(),
		Log:          c.Log.WithName("controllers").WithName("Launch"),
		Scheme:       c.scheme,
	}).SetupWithManager(c.mgr); err != nil {
		return err
	}

	if err := (&repository.ReconcileRepository{
		Client:       c.mgr.GetClient(),
		HelmSettings: c.helmFlagMgr.EnvSettings(),
		Log:          c.Log.WithName("controllers").WithName("Repository"),
		Scheme:       c.scheme,
	}).SetupWithManager(c.mgr); err != nil {
		return err
	}

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	return common.Multiblock(c.Log, c.startHTTPServer, c.startManager)
}

func (c *command) startHTTPServer(ctx context.Context) error {
	v := validator.New(c.Log, c.rc, c.helmFlagMgr.EnvSettings())
	m, err := router.New(c.Log).Route(router.Defaults(v1.Defaults(c.Log, v)))
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

	h := http.New(c.Log, m, c.httpFlagMgr.HttpPort())
	if p := c.httpFlagMgr.HttpsPort(); p != 0 {
		if err = h.ConfigureTLS(p, cf, kf); err != nil {
			return err
		}
	}
	h.Serve(ctx)
	return nil
}

func (c *command) startManager(ctx context.Context) error {
	// And now, run.  And wait.
	c.Log.Info("Starting launch manager")
	return c.mgr.Start(ctx.Done())
}
