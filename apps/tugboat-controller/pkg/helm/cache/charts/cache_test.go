package charts

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/testing/chartmuseum"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/testing/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

func Test_Cache_Charts(t *testing.T) {
	s := NewStatefulTest()
	defer s.Close()

	s.Run(t, s)
}

type StatefulTest struct {
	*chartmuseum.StatefulTest

	cachedir string
	settings *cli.EnvSettings
	rc       *repos.Cache
}

func NewStatefulTest() *StatefulTest {
	s := &StatefulTest{
		StatefulTest: chartmuseum.NewStatefulTest(),
	}

	s.cachedir = filepath.Join(s.ParentDir, "tarballs")

	s.settings = cli.New()
	s.settings.Debug = true
	s.settings.RepositoryCache = s.RepositoryCacheDir
	s.settings.RepositoryConfig = s.RepositoryConfigFile

	s.rc = repos.New()
	s.rc.Connect(repos.WithHelmEnvSettings(s.settings))
	s.rc.UpdateRepositories()

	return s
}

func (s *StatefulTest) Test_Repo_Cache_GetOne(t *testing.T) {
	c := New()
	err := c.Connect(WithCacheDepth(128), WithCacheDirectory(s.cachedir), WithHelmEnvSettings(s.settings), WithRepoCache(s.rc))
	if err != nil {
		t.Errorf("Internal error: failed to connect:\n\t%s\n", err.Error())
		return
	}

	f, err := c.Get(chartmuseum.TestRepositoryName, "app-foo", utils.WillMakeVersion("0.1.0"))
	if err != nil {
		t.Errorf("Did not download chart tarball: %s", err.Error())
	}

	if fi, err := os.Stat(f); err != nil {
		t.Errorf("Returned path is invalid: %s: %s", f, err.Error())
	} else if fi.IsDir() {
		t.Errorf("Returned path is a directory: %s", f)
	}
}

func (s *StatefulTest) Test_Cache_Charts_GetOneChartTwice(t *testing.T) {
	c := New()
	err := c.Connect(WithCacheDepth(128), WithCacheDirectory(s.cachedir), WithHelmEnvSettings(s.settings), WithRepoCache(s.rc))
	if err != nil {
		t.Errorf("Internal error: failed to connect:\n\t%s\n", err.Error())
		return
	}

	name := s.GenerateName(12)
	v := utils.WillMakeVersion("0.2.0")
	s.CreateTestChart(name, v)

	f0, err := c.Get(chartmuseum.TestRepositoryName, name, v)
	if err != nil {
		t.Errorf("Did not download chart tarball: %s", err.Error())
	}

	f1, err := c.Get(chartmuseum.TestRepositoryName, name, v)
	if err != nil {
		t.Errorf("Did not download chart tarball: %s", err.Error())
	}

	if f0 != f1 {
		t.Errorf("Second download returned different path: %s, %s", f0, f1)
	}

	if n, ok := s.Srv.Requests[fmt.Sprintf("/charts/%s-%s.tgz", name, v)]; !ok {
		s.Srv.DumpRequests()
		t.Errorf("No record of call")
	} else if n != 1 {
		s.Srv.DumpRequests()
		t.Errorf("Have record of %d calls", n)
	}
}

func (s *StatefulTest) Test_Cache_Charts_GetTwo(t *testing.T) {
	c := New()
	err := c.Connect(WithCacheDepth(128), WithCacheDirectory(s.cachedir), WithHelmEnvSettings(s.settings), WithRepoCache(s.rc))
	if err != nil {
		t.Errorf("Internal error: failed to connect:\n\t%s\n", err.Error())
		return
	}

	name := s.GenerateName(12)
	v0 := utils.WillMakeVersion("0.2.0")
	v1 := utils.WillMakeVersion("0.2.1")

	s.CreateTestChart(name, v0)
	s.CreateTestChart(name, v1)

	f0, err := c.Get(chartmuseum.TestRepositoryName, name, v0)
	f1, err := c.Get(chartmuseum.TestRepositoryName, name, v1)

	// The returned paths must be different.
	if f0 == f1 {
		t.Errorf("Paths returned for different charts are the same: '%s'", f0)
	}
}

func (s *StatefulTest) Test_Cache_Charts_GetBeyondCache(t *testing.T) {
	c := New()
	err := c.Connect(WithCacheDepth(2), WithCacheDirectory(s.cachedir), WithHelmEnvSettings(s.settings), WithRepoCache(s.rc))
	if err != nil {
		t.Errorf("Internal error: failed to connect:\n\t%s\n", err.Error())
		return
	}

	name := s.GenerateName(12)
	vs := map[*semver.Version]bool{
		utils.WillMakeVersion("0.2.0"): true,
		utils.WillMakeVersion("0.2.1"): false,
		utils.WillMakeVersion("0.2.2"): false,
	}

	for v := range vs {
		s.CreateTestChart(name, v)
	}

	for k := range vs {
		_, err = c.Get(chartmuseum.TestRepositoryName, name, k)
		if err != nil {
			t.Errorf("Did not download chart tarball: %s", err.Error())
		}
	}

	dir := filepath.Join(s.cachedir, chartmuseum.TestRepositoryName)

	for k, purged := range vs {
		filename := fmt.Sprintf("%s-%s.tgz", name, k)
		fullpath := filepath.Join(dir, filename)
		if _, err := os.Stat(fullpath); err == os.ErrNotExist && !purged {
			t.Errorf("File '%s' should not have been purged, but was", fullpath)
		} else if err == nil && purged {
			t.Errorf("File '%s' should have been purged, but was not", fullpath)
		}
	}
}

func (s *StatefulTest) Test_Cache_Charts_Unpack(t *testing.T) {
	c := New()
	err := c.Connect(WithCacheDepth(128), WithCacheDirectory(s.cachedir), WithHelmEnvSettings(s.settings), WithRepoCache(s.rc))
	if err != nil {
		t.Errorf("Internal error: failed to connect:\n\t%s\n", err.Error())
		return
	}

	name := s.GenerateName(12)
	v := utils.WillMakeVersion("0.3.0")
	s.CreateTestChart(name, v)

	ref := v1alpha1.ChartReference{
		Repository: chartmuseum.TestRepositoryName,
		Chart:      name,
		Version:    v,
	}
	unpackdir, clean, err := c.Unpack(ref)
	if err != nil {
		t.Errorf("%s\n", err.Error())
	}
	defer func() {
		if !t.Failed() {
			clean()
		}
	}()

	act := action.NewLint()
	res := act.Run([]string{unpackdir}, nil)
	if len(res.Errors) != 0 {
		for _, e := range res.Errors {
			t.Errorf(e.Error())
		}
	}
	for _, m := range res.Messages {
		t.Logf("%s: %d: %s", m.Path, m.Severity, m.Error())
	}
}
