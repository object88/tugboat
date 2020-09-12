package cliflags

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/pkg/logging"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type FlagManager struct {
	env *cli.EnvSettings
}

func New() *FlagManager {
	return &FlagManager{
		env: cli.New(),
	}
}

// ConfigureFlags uses the helm packages to add helm-related flags to the CLI.
// This is paired with `HelmEnvSettings`.
func (fl *FlagManager) ConfigureFlags(flags *pflag.FlagSet) {
	// fl.helmSettings = cli.New()
	fl.env.AddFlags(flags)
}

func (fl *FlagManager) ActionConfig(log logr.Logger) (*action.Configuration, error) {
	a := &action.Configuration{}
	helmDriver := os.Getenv("HELM_DRIVER")
	if err := a.Init(fl.env.RESTClientGetter(), fl.env.Namespace(), helmDriver, (&logging.Adapter{Log: log}).Logf); err != nil {
		return nil, err
	}
	return a, nil
}

func (fl *FlagManager) Client() genericclioptions.RESTClientGetter {
	return fl.env.RESTClientGetter()
}

func (fl *FlagManager) EnvSettings() *cli.EnvSettings {
	return fl.env
}
