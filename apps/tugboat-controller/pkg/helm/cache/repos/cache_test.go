package repos

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/testing/chartmuseum"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/testing/utils"
	"github.com/object88/tugboat/pkg/logging/testlogger"
	"helm.sh/helm/v3/pkg/cli"
)

func Test_Cache_Repos(t *testing.T) {
	s := NewStatefulTest(t, testlogger.TestLogger{T: t})
	defer s.Close()

	s.Run(t, s)
}

type StatefulTest struct {
	*chartmuseum.StatefulTest
}

func NewStatefulTest(t *testing.T, logger logr.Logger) *StatefulTest {
	return &StatefulTest{
		StatefulTest: chartmuseum.NewStatefulTest(t, logger),
	}
}

func (s *StatefulTest) Test_Cache_Repos_GetVersion(t *testing.T) {
	helmSettings := cli.New()
	helmSettings.Debug = true
	helmSettings.RepositoryCache = s.RepositoryCacheDir
	helmSettings.RepositoryConfig = s.RepositoryConfigFile

	rc := New()
	err := rc.Connect(
		WithCooldown(5*time.Microsecond),
		WithHelmEnvSettings(helmSettings),
		WithTimeout(10*time.Microsecond),
	)
	if err != nil {
		t.Errorf("Internal error: failed to connect:\n\t%s\n", err.Error())
		return
	}

	rc.logger.Info("About to update repositories")

	if err = rc.UpdateRepositories(); err != nil {
		t.Errorf("Failed to update repositories:\n\t%s\n", err.Error())
	}

	rc.logger.Info("Updated")

	v := utils.WillMakeVersion("0.1.0")
	cm, err := rc.GetChartVersion(chartmuseum.TestRepositoryName, "app-foo", v)
	if err != nil {
		t.Errorf("Failed to get chart metadata:\n\t%s\n", err.Error())
	}

	if cm == nil {
		t.Errorf("Failed to get chart metadata back for chart '%s' version '%s'", "app-foo", v.String())
	}
}
