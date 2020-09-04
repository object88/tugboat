package repositories

import (
	"io/ioutil"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/repoCache"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/repoCache/cliflags"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/repo"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	helm        *repoCache.Helm
	helmFlagMgr *cliflags.FlagManager
}

func CreateCommand(cmmn *common.CommonArgs) *cobra.Command {
	var c command
	c = command{
		Command: cobra.Command{
			Use:   "repositories",
			Short: "Configure helm repositories",
			Args:  cobra.NoArgs,
			PreRunE: func(cmd *cobra.Command, args []string) error {
				return c.preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.execute(cmd, args)
			},
		},
		CommonArgs:  cmmn,
		helmFlagMgr: cliflags.NewFlagManager(),
	}

	flags := c.Flags()

	c.helmFlagMgr.ConfigureMuseumConfigFileFlag(flags)
	c.helmFlagMgr.ConfigureMuseumsFlag(flags)
	c.helmFlagMgr.ConfigureHelm(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	c.helm = repoCache.New()

	err := c.helm.Connect(
		repoCache.WithHelmEnvSettings(c.helmFlagMgr.HelmEnvSettings()),
		repoCache.WithLogger(c.Log),
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	filepath, err := c.helmFlagMgr.MuseumConfigFile()
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	x := struct {
		Charts []*repo.Entry `yaml:"chartRepositories"`
	}{}
	if yaml.Unmarshal(buf, &x); err != nil {
		return err
	}

	c.Log.Info("have repositories from config", "count", len(x.Charts))

	for _, entry := range x.Charts {
		c.Log.Info("ensuring repository", "name", entry.Name)
		if err = c.helm.EnsureRepo(entry); err != nil {
			return err
		}
	}

	museums, err := c.helmFlagMgr.Museums()
	if err != nil {
		return err
	}

	c.Log.Info("have repositories from explicit flag", "count", len(museums))

	for _, entry := range museums {
		c.Log.Info("ensuring repository", "name", entry.Name)
		if err = c.helm.EnsureRepo(entry); err != nil {
			return err
		}
	}

	c.Log.Info("updating all repositories.")

	err = c.helm.UpdateRepositories()
	if err != nil {
		return err
	}

	c.Log.Info("done")

	return nil
}
