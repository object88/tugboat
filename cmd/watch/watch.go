package watch

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/object88/tugboat/internal/cmd/cliflags"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	v1 "k8s.io/api/core/v1"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	cflags *genericclioptions.ConfigFlags

	output cliflags.Output
}

// CreateCommand returns the watch command
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c *command
	c = &command{
		Command: cobra.Command{
			Use:   "watch",
			Short: "report the version of the tool",
			Args:  cobra.NoArgs,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				return c.Preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.Execute(cmd, args)
			},
		},
		CommonArgs: ca,
	}

	flags := c.Flags()

	c.FlagMgr.ConfigureOutputFlag(flags)

	c.cflags = genericclioptions.NewConfigFlags(false)
	c.cflags.AddFlags(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) Preexecute(cmd *cobra.Command, args []string) error {
	c.output = c.FlagMgr.Output()

	config, err := c.cflags.ToRESTConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to get REST config")
	}

	// f := cmdutil.NewFactory(c.cflags)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "Failed to get clientset")
	}

	client := clientset.CoreV1().RESTClient()

	listWatch := cache.NewListWatchFromClient(client, "events", "", fields.Everything())

	funcs := cache.ResourceEventHandlerFuncs{
		AddFunc:    printEvent(clientset, client),
		DeleteFunc: printEvent(clientset, client),
	}
	_, controller := cache.NewInformer(listWatch, &v1.Event{}, time.Second*0, funcs)

	settings := cli.New()

	actionConfig := new(action.Configuration)
	err = actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), c.Log.Infof)
	if err != nil {
		// debug("%+v", err)
		// os.Exit(1)
		return errors.Wrapf(err, "Failed to init actionConfig")
	}

	release, err := action.NewGet(actionConfig).Run("foo")
	if err != nil {
		return errors.Wrapf(err, "Failed to get chart foo")
	}

	// values, err := action.NewGetValues(actionConfig).Run("foo")
	// if err != nil {
	// 	return errors.Wrapf(err, "Failed to get values for chart foo")
	// }

	_, err = chartutil.ToRenderValues(release.Chart, release.Config, chartutil.ReleaseOptions{}, nil)
	if err != nil {
		return errors.Wrapf(err, "Failed to render values")
	}
	// vals.

	fmt.Printf("Templates:\n")
	for k, v := range release.Chart.Templates {
		var y yaml.MapSlice
		dec := yaml.NewDecoder(bytes.NewReader(v.Data))
		err := dec.Decode(y)
		if err != nil {
			return errors.Wrapf(err, "Failed to decode '%s': %s", v.Name, v.Data)
		}

		fmt.Printf("\t%s (%d): %#v\n", v.Name, k, y)
	}

	// Start the controller:
	// go controller.Run(wait.NeverStop)
	controller.Run(wait.NeverStop)

	return nil
}

func (c *command) Execute(cmd *cobra.Command, args []string) error {

	return nil
}

func printEvent(clientset *kubernetes.Clientset, client rest.Interface) func(interface{}) {
	// v1client, err := clientv1.NewForConfig(client.)
	return func(obj interface{}) {
		// "k8s.io/apimachinery/pkg/apis/meta/v1" provides an Object
		// interface that allows us to get metadata easily
		// mObj := obj.(metav1.Object)
		// if release, ok := mObj.GetLabels()["Release"]; ok {
		// 	fmt.Printf("*** %s Event:\n\t%s\n", release, .String())
		// } else {
		// 	fmt.Printf("Event:\n\t%s\n", x.String())
		// }

		switch x := obj.(type) {
		case *v1.Event:
			// x.Object
			objref := x.InvolvedObject
			switch objref.Kind {
			case "service":

				// client.Get().
				// clientset.CoreV1().Services
			}

			// objref.UID
			if release, ok := x.ObjectMeta.GetLabels()["Release"]; ok {
				fmt.Printf("*** %s Event:\n\t%s\n", release, x.String())
			} else {
				fmt.Printf("Event:\n\t%s\n", x.String())
			}
		default:
			fmt.Printf("%s\n", obj)
		}
	}
}
