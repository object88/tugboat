package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/pkg/logging"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type Installer struct {
	log       logr.Logger
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
			// An ErrReleaseNotFound just indicates that the chart is not installed.
			// so clear the error.
			err = nil
		}
		return false, err
	}
	return rel != nil, nil
}

func (i *Installer) Delete(name string) error {
	act := action.NewUninstall(i.actionConfig())
	if _, err := act.Run(name); err != nil {
		return err
	}

	return nil
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

func (i *Installer) Lint(location string, launch *v1alpha1.Launch) error {
	lintAct := action.NewLint()
	lintAct.Strict = true

	// figure out how to turn launch.Spec.Values into a map[string]interface{}
	// for linting.
	var errs *multierror.Error
	res := lintAct.Run([]string{location}, nil)
	if len(res.Errors) != 0 {
		for _, err := range res.Errors {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (i *Installer) Pull(launch *v1alpha1.Launch) (string, error) {
	if err := os.MkdirAll(os.TempDir(), 0755); err != nil {
		return "", err
	}
	destination, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	pullAct := action.NewPull()
	pullAct.DestDir = destination
	pullAct.Settings = i.settings
	pullAct.Untar = true
	pullAct.UntarDir = launch.Name
	pullAct.Version = launch.Spec.Version.String()
	if _, err = pullAct.Run(launch.Spec.Chart); err != nil {
		os.RemoveAll(destination)
		return "", err
	}

	return destination, nil
}

func (i *Installer) Update(name string, launch *v1alpha1.LaunchSpec) error {
	chartName := fmt.Sprintf("%s:%s", launch.Chart, launch.Version)
	if launch.Repository != "" {
		trimmed := launch.Chart[strings.Index(launch.Chart, "/"):]
		chartName = fmt.Sprintf("%s%s-%s.tgz", launch.Repository, trimmed, launch.Version.String())
	}
	i.log.Info("Referenced chart", "address", chartName)

	upgradeAct := action.NewUpgrade(i.actionConfig())

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
