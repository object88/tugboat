package run

import (
	"context"
	"fmt"
	"time"

	"github.com/object88/tugboat/apps/tugboat-watcher/pkg/informerhandlers"
	"github.com/object88/tugboat/internal/cmd/common"
	notificationsclient "github.com/object88/tugboat/internal/notifications/client"
	notificationscliflags "github.com/object88/tugboat/internal/notifications/cliflags"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/probes"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/object88/tugboat/pkg/k8s/client/clientset/versioned"
	"github.com/object88/tugboat/pkg/k8s/client/informers/externalversions"
	"github.com/object88/tugboat/pkg/k8s/cliflags"
	k8scliflags "github.com/object88/tugboat/pkg/k8s/cliflags"
	"github.com/object88/tugboat/pkg/k8s/informermanager"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	httpFlagMgr          *httpcliflags.FlagManager
	k8sFlagMgr           *k8scliflags.FlagManager
	notificationsFlagMgr *notificationscliflags.FlagManager

	versionedclientset *versioned.Clientset

	// w                      cache.SharedIndexInformer
	eventinformer          cache.SharedIndexInformer
	releasehistoryinformer cache.SharedIndexInformer
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

	getter := c.k8sFlagMgr.KubernetesConfig()

	cfg, err := getter.ToRESTConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	c.versionedclientset, err = versioned.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// r, err := labels.NewRequirement("tugboat.engineering/releasehistory", selection.Exists, nil)
	// if err != nil {
	//  // log.Info("failed to create requirement", "err0", err0.Error(), "err1", err1.Error())
	//  return err
	// }

	fact := informers.NewSharedInformerFactoryWithOptions(clientset, 1*time.Second, informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
		// lo.LabelSelector = labels.NewSelector().Add(*r).String()
		lo.FieldSelector = field.NewPath("involvedObject.namespace=default").String()
	}))

	c.eventinformer = fact.Core().V1().Events().Informer()

	c.eventinformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			evt, ok := obj.(*v1.Event)
			if !ok {
				return
			}
			fmt.Printf("thing added: %s \n", evt.InvolvedObject.Name)
		},
		DeleteFunc: func(obj interface{}) {
			// fmt.Printf("service deleted: %s \n", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// fmt.Printf("service changed \n")
		},
	})

	// v1.Event.

	// c.w = watcher.New(c.Log, clientset)

	factory := externalversions.NewSharedInformerFactory(c.versionedclientset, 10*time.Second)
	c.releasehistoryinformer = factory.Tugboat().V1alpha1().ReleaseHistories().Informer()

	handler, err := informerhandlers.NewReleaseHistory(c.Log)
	if err != nil {
		return err
	}
	c.releasehistoryinformer.AddEventHandler(handler)

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
		mgr := informermanager.New(c.Log)
		return mgr.Run(ctx, r, c.eventinformer, c.releasehistoryinformer)
	}

	return common.Multiblock(c.Log, p, f0, f1)
}
