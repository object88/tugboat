package repoCache

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"helm.sh/helm/v3/pkg/cli"
)

func Test_ChartMuseum(t *testing.T) {
	s := NewStatefulTest()
	defer s.Close()

	s.Run(t)
}

func (s *StatefulTest) Test_Something(t *testing.T) {
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

	v := willMakeVersion("0.1.0")
	cm, err := h.GetChartMetadata(TestRepositoryName, "app-foo", v)
	if err != nil {
		t.Errorf("Failed to get chart metadata:\n\t%s\n", err.Error())
	}

	if cm == nil {
		t.Errorf("Failed to get chart metadata back for chart '%s' version '%s'", "app-foo", v.String())
	}
}

func willMakeVersion(in string) *semver.Version {
	v, _ := semver.NewVersion(in)
	return v
}
