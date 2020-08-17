package cliflags

import (
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type FlagManager struct {
	kubeConfigFlags *genericclioptions.ConfigFlags
}

func New() *FlagManager {
	return &FlagManager{}
}

func (fl *FlagManager) ConfigureKubernetesConfig(flags *pflag.FlagSet) {
	fl.kubeConfigFlags = genericclioptions.NewConfigFlags(false)
	fl.kubeConfigFlags.AddFlags(flags)
}

func (fl *FlagManager) KubernetesConfig() genericclioptions.RESTClientGetter {
	return fl.kubeConfigFlags
}
