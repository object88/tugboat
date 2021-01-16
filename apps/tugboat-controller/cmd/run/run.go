package run

import (
	"context"
	"time"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/controller/releasehistory"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/controller/secret"
	v1 "github.com/object88/tugboat/apps/tugboat-controller/pkg/http/router/v1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/validator"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/probes"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/object88/tugboat/pkg/k8s/apis"
	"github.com/object88/tugboat/pkg/k8s/client/clientset/versioned"
	"github.com/object88/tugboat/pkg/k8s/client/informers/externalversions"
	"github.com/object88/tugboat/pkg/k8s/client/listers/engineering.tugboat/v1alpha1"
	k8scliflags "github.com/object88/tugboat/pkg/k8s/cliflags"
	"github.com/object88/tugboat/pkg/k8s/informermanager"
	"github.com/spf13/cobra"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	dyn                    dynamic.Interface
	mapper                 *restmapper.DeferredDiscoveryRESTMapper
	mgr                    manager.Manager
	scheme                 *runtime.Scheme
	versionedclientset     *versioned.Clientset
	releasehistoryinformer cache.SharedIndexInformer

	httpFlagMgr *httpcliflags.FlagManager
	k8sFlagMgr  *k8scliflags.FlagManager

	probe *probes.Probe
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
		k8sFlagMgr:  k8scliflags.New(),
	}

	flags := c.Flags()

	c.httpFlagMgr.ConfigureHttpFlag(flags)
	c.httpFlagMgr.ConfigureHttpsFlags(flags)
	c.k8sFlagMgr.ConfigureKubernetesConfig(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	// namespace := "default"

	// // return watchers.New().Run(ctx, c.w)
	// // Set default manager options
	// options := manager.Options{
	// 	Namespace: namespace,
	// }

	// // Add support for MultiNamespace set in WATCH_NAMESPACE (e.g ns1,ns2)
	// // Note that this is not intended to be used for excluding namespaces, this is better done via a Predicate
	// // Also note that you may face performance issues when using this with a high number of namespaces.
	// // More Info: https://godoc.org/github.com/kubernetes-sigs/controller-runtime/pkg/cache#MultiNamespacedCacheBuilder
	// if strings.Contains(namespace, ",") {
	// 	c.Log.Info(fmt.Sprintf("Using multi-namespace: %s", namespace))
	// 	options.Namespace = ""
	// 	options.NewCache = cache.MultiNamespacedCacheBuilder(strings.Split(namespace, ","))
	// }

	var err error
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

	getter := c.k8sFlagMgr.KubernetesConfig()

	cfg, err := getter.ToRESTConfig()
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

	c.versionedclientset, err = versioned.NewForConfig(cfg)
	if err != nil {
		return err
	}
	factory := externalversions.NewSharedInformerFactory(c.versionedclientset, 10*time.Second)
	c.releasehistoryinformer = factory.Tugboat().V1alpha1().ReleaseHistories().Informer()

	dc, err := getter.ToDiscoveryClient()
	if err != nil {
		return err
	}
	c.dyn, err = dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}
	c.mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	c.probe = probes.New()

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	return common.Multiblock(c.Log, c.probe, c.startHTTPServer, c.startControllerManager, c.startInformerManager)
}

func (c *command) startHTTPServer(ctx context.Context, r probes.Reporter) error {
	lister := v1alpha1.NewReleaseHistoryLister(c.releasehistoryinformer.GetIndexer())

	m := validator.NewMutator(c.Log, c.versionedclientset, lister, c.dyn, c.mapper)
	v := validator.New(c.Log, c.scheme)
	v2 := validator.NewV2(c.Log, c.scheme, c.versionedclientset, lister)
	rts, err := router.New(c.Log).Route(router.LoggingDefaultRoute, router.Defaults(c.probe, v1.Defaults(c.Log, m, v, v2)))
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

	h.Serve(ctx, r)
	return nil
}

func (c *command) startControllerManager(ctx context.Context, r probes.Reporter) error {
	// And now, run.  And wait.
	c.Log.Info("starting controller manager")
	defer c.Log.Info("controller manager complete")

	if err := (&releasehistory.ReconcileReleaseHistory{
		Client: c.mgr.GetClient(),
		Log:    c.Log,
		Scheme: c.scheme,
	}).SetupWithManager(c.mgr); err != nil {
		return err
	}

	if err := (&secret.ReconcileSecret{
		Client:          c.mgr.GetClient(),
		VersionedClient: c.versionedclientset,
		Log:             c.Log,
	}).SetupWithManager(c.mgr); err != nil {
		return err
	}

	r.Ready()

	return c.mgr.Start(ctx)
}

func (c *command) startInformerManager(ctx context.Context, r probes.Reporter) error {
	c.Log.Info("starting watcher")
	defer c.Log.Info("watcher complete")

	mgr := informermanager.New(c.Log)
	return mgr.Run(ctx, r, c.releasehistoryinformer)
}
