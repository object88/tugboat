package helm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/pkg/logging"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type Installer struct {
	log logr.Logger
	// getter    genericclioptions.RESTClientGetter
	namespace string
	settings  *cli.EnvSettings
}

func New(log logr.Logger, settings *cli.EnvSettings) *Installer {
	return &Installer{
		log:       log,
		settings:  settings,
		namespace: "default",
	}
}

func (i *Installer) IsDeployed(name string) (bool, error) {
	rel, err := action.NewStatus(i.actionConfig()).Run(name)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			err = nil
		}
		return false, err
	}
	return rel != nil, nil
}

func (i *Installer) Deploy(name string, launch *v1alpha1.LaunchSpec) error {
	chartName := fmt.Sprintf("%s:%s", launch.Chart, launch.Version)
	if launch.Repository != "" {
		trimmed := launch.Chart[strings.Index(launch.Chart, "/"):]
		chartName = fmt.Sprintf("%s%s-%s.tgz", launch.Repository, trimmed, launch.Version.String())
	}
	i.log.Info("Referenced chart", "address", chartName)

	act := action.NewInstall(i.actionConfig())
	act.ReleaseName = name

	chartPath, err := act.ChartPathOptions.LocateChart(chartName, i.settings)
	if err != nil {
		return err
	}

	// vals, err := valueOpts.MergeValues(getter.All(settings))
	// if err != nil {
	// 	return err
	// }

	// Check chart dependencies to make sure all are present in /charts
	ch, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return err
		}
	}

	if _, err := act.Run(ch, nil); err != nil {
		return err
	}

	return nil
}

func (i *Installer) Update(name string, launch *v1alpha1.LaunchSpec) error {
	chartName := fmt.Sprintf("%s:%s", launch.Chart, launch.Version)
	if launch.Repository != "" {
		trimmed := launch.Chart[strings.Index(launch.Chart, "/"):]
		chartName = fmt.Sprintf("%s%s-%s.tgz", launch.Repository, trimmed, launch.Version.String())
	}
	i.log.Info("Referenced chart", "address", chartName)

	ac := i.actionConfig()
	// act := action.NewChartPull(ac)

	// if err := act.Run(os.Stderr, chartName); err != nil {
	// 	return err
	// }

	upgradeAct := action.NewUpgrade(ac)

	chartPath, err := upgradeAct.ChartPathOptions.LocateChart(chartName, i.settings)
	if err != nil {
		return err
	}

	// vals, err := valueOpts.MergeValues(getter.All(settings))
	// if err != nil {
	// 	return err
	// }

	// Check chart dependencies to make sure all are present in /charts
	ch, err := loader.Load(chartPath)
	if err != nil {
		return err
	}
	if req := ch.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(ch, req); err != nil {
			return err
		}
	}

	if _, err := upgradeAct.Run(name, ch, nil); err != nil {
		return err
	}

	return nil
}

func (i *Installer) actionConfig() *action.Configuration {
	a := &action.Configuration{
		Capabilities: chartutil.DefaultCapabilities,
	}
	a.Init(i.settings.RESTClientGetter(), i.namespace, "", (&logging.Adapter{Log: i.log}).Logf)
	return a
}
