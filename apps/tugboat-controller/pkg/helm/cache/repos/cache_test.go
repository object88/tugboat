package repos

import (
	"testing"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/testing/chartmuseum"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/testing/utils"
	"helm.sh/helm/v3/pkg/cli"
)

func Test_Cache_Repos(t *testing.T) {
	s := NewStatefulTest()
	defer s.Close()

	s.Run(t, s)
}

type StatefulTest struct {
	*chartmuseum.StatefulTest
}

func NewStatefulTest() *StatefulTest {
	return &StatefulTest{
		StatefulTest: chartmuseum.NewStatefulTest(),
	}
}

func (s *StatefulTest) Test_Cache_Repos_GetVersion(t *testing.T) {
	helmSettings := cli.New()
	helmSettings.Debug = true
	helmSettings.RepositoryCache = s.RepositoryCacheDir
	helmSettings.RepositoryConfig = s.RepositoryConfigFile

	h := New()
	err := h.Connect(WithHelmEnvSettings(helmSettings))
	if err != nil {
		t.Errorf("Internal error: failed to connect:\n\t%s\n", err.Error())
		return
	}

	h.logger.Info("About to update repositories")

	if err = h.UpdateRepositories(); err != nil {
		t.Errorf("Failed to update repositories:\n\t%s\n", err.Error())
	}

	h.logger.Info("Updated")

	v := utils.WillMakeVersion("0.1.0")
	cm, err := h.GetChartVersion(chartmuseum.TestRepositoryName, "app-foo", v)
	if err != nil {
		t.Errorf("Failed to get chart metadata:\n\t%s\n", err.Error())
	}

	if cm == nil {
		t.Errorf("Failed to get chart metadata back for chart '%s' version '%s'", "app-foo", v.String())
	}
}
