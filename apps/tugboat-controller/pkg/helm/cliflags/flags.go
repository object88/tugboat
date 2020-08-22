package cliflags

import (
	"github.com/spf13/pflag"
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

func (fl *FlagManager) ConfigureFlags(flags *pflag.FlagSet) {
	fl.env.AddFlags(flags)
}

func (fl *FlagManager) Client() genericclioptions.RESTClientGetter {
	return fl.env.RESTClientGetter()
}
