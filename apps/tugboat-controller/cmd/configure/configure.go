package configure

import (
	"context"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/client/clientset/versioned"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos"
	helmcliflags "github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cliflags"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/repo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	helm        *repos.Cache
	helmFlagMgr *helmcliflags.FlagManager
}

func CreateCommand(cmmn *common.CommonArgs) *cobra.Command {
	var c command
	c = command{
		Command: cobra.Command{
			Use: "configure",
			// Short: "Configure various aspects of the tugboat controller",
			Args: cobra.NoArgs,
			PreRunE: func(cmd *cobra.Command, args []string) error {
				return c.preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.execute(cmd, args)
			},
		},
		CommonArgs:  cmmn,
		helmFlagMgr: helmcliflags.New(),
	}

	flags := c.Flags()

	c.helmFlagMgr.ConfigureFlags(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	c.helm = repos.New()

	err := c.helm.Connect(
		repos.WithHelmEnvSettings(c.helmFlagMgr.EnvSettings()),
		repos.WithLogger(c.Log),
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	cfg, err := c.helmFlagMgr.EnvSettings().RESTClientGetter().ToRESTConfig()
	if err != nil {
		return err
	}
	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return err
	}
	list, err := cs.TugboatV1alpha1().Repositories("").List(context.Background(), v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, v := range list.Items {
		r := repo.Entry{
			Name: v.Spec.Name,
			URL:  v.Spec.URL,
		}

		c.helm.UpsertRepo(&r)
	}
	return nil
}
