package helm

import (
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/sirupsen/logrus"
)

type Installer struct {
	log *logrus.Logger
}

func New(log *logrus.Logger) *Installer {
	return &Installer{
		log: log,
	}
}

func Deploy(launch *v1alpha1.LaunchSpec) error {
	// act := action.NewPull()
	// act.LocateChart()
	// act.Run()
	return nil
}
