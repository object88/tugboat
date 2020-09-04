package helm

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

type StatefulTest struct {
	ProjectRootDir string

	ParentDir            string
	MuseumStorageDir     string
	RepositoryCacheDir   string
	RepositoryConfigFile string

	GenerateName func(length int) string

	srv *testChartMuseum
}

const (
	TestRepositoryName string = "local"
)

// NewStatefulTest returns a pointer to a new StatefulTest instance.  NewStatefulTest and related
// initalization functions will panic upon failure.
func NewStatefulTest() *StatefulTest {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	s := &StatefulTest{}
	s.GenerateName = func(length int) string {
		b := make([]byte, length)
		for i := range b {
			b[i] = charset[seededRand.Intn(len(charset))]
		}
		return string(b)
	}

	s.findProjectRoot()
	s.initializeParentDir()
	s.initializeMuseumDir()
	s.initializeChartMuseum()
	s.initializeChartRepository()

	if err := s.createTestChart("app-foo", willMakeVersion("0.1.0")); err != nil {
		panic(err.Error())
	}

	s.refreshIndex()

	return s
}

func (s *StatefulTest) Run(t *testing.T) {
	val := reflect.ValueOf(s)
	tval := val.Type()
	for i := 0; i < tval.NumMethod(); i++ {
		mval := tval.Method(i)
		if !strings.HasPrefix(mval.Name, "Test_") {
			continue
		}
		fn := mval.Func
		tfn := fn.Type()
		if tfn.NumOut() != 0 || tfn.NumIn() != 2 || tfn.In(1) != reflect.TypeOf(t) {
			continue
		}

		t.Run(mval.Name, func(t *testing.T) {
			t.Logf("Starting '%s'", mval.Name)

			// Invoke the test
			fn.Call([]reflect.Value{reflect.ValueOf(s), reflect.ValueOf(t)})
			if !t.Failed() {
				// Exit early; test was successful.
				t.Logf("Completed '%s'", mval.Name)
				return
			}

			t.Logf("Completed with error '%s'", mval.Name)
		})
	}
}

func (s *StatefulTest) Close() error {
	if s.srv != nil {
		s.srv.Close()
		s.srv = nil
	}
	if s.ParentDir != "" {
		os.RemoveAll(s.ParentDir)
		s.ParentDir = ""
	}
	return nil
}

func (s *StatefulTest) findProjectRoot() {
	workingdir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("internal error: failed to get current working directory: %w", err))
	}
	var checkdir func(startdir string) string
	checkdir = func(startdir string) string {
		if startdir == "" {
			panic(fmt.Errorf("internal error: failed to find project root"))
		}
		fi, err := os.Stat(filepath.Join(startdir, "go.mod"))
		if err != nil {
			if !os.IsNotExist(err) {
				panic(fmt.Errorf"internal error: got unexpected non-os.PathError from os.Stat: %w", err))
			}
			// Not here, keep moving.
		} else if !fi.IsDir() {
			// Found it.
			return startdir
		}
		return checkdir(filepath.Dir(startdir))
	}
	s.ProjectRootDir = checkdir(workingdir)
}

func (s *StatefulTest) initializeMuseumDir() {
	s.MuseumStorageDir = filepath.Join(s.ParentDir, "charts")
	if err := os.MkdirAll(s.MuseumStorageDir, 0777); err != nil {
		panic(fmt.Errorf("internal error: failed to create museum storage directory: %w", err))
	}
}

func (s *StatefulTest) initializeParentDir() {
	chartname := fmt.Sprintf("test-%s", s.GenerateName(8))
	s.ParentDir, _ = ioutil.TempDir(os.TempDir(), chartname)
	s.RepositoryCacheDir, _ = ioutil.TempDir(s.ParentDir, "*-cache")
}

func (s *StatefulTest) initializeChartMuseum() {
	var err error
	s.srv, err = newTestChartMuseum(s.MuseumStorageDir)
	if err != nil {
		panic(err)
	}
	s.srv.Run()
}

func (s *StatefulTest) initializeChartRepository() {
	repositoryConfigDir, _ := ioutil.TempDir(s.ParentDir, "*-repository")
	s.RepositoryConfigFile = filepath.Join(repositoryConfigDir, "repository.yaml")

	f := repo.NewFile()
	f.Add(&repo.Entry{
		Name: TestRepositoryName,
		URL:  s.srv.httpserver.URL,
	})
	if err := f.WriteFile(s.RepositoryConfigFile, 0644); err != nil {
		panic(fmt.Errorf("internal error: failed to write repository config '%s': %w", s.RepositoryConfigFile, err))
	}

	helmSettings := cli.New()
	helmSettings.Debug = true
	// helmSettings.RegistryConfig = filepath.Join(repositoryConfigDir, "registry.json")
	helmSettings.RepositoryCache = s.RepositoryCacheDir
	helmSettings.RepositoryConfig = s.RepositoryConfigFile

	entry := f.Repositories[0]
	cr, err := repo.NewChartRepository(entry, getter.All(helmSettings))
	if err != nil {
		panic(fmt.Errorf("internal error: failed to create new chart repository: %w", err))
	}
	cr.CachePath = helmSettings.RepositoryCache
}

func (s *StatefulTest) createTestChart(chartname string, version *semver.Version) error {
	chartdir, _ := ioutil.TempDir(s.ParentDir, chartname)

	cfile := &chart.Metadata{
		Name:        chartname,
		Description: fmt.Sprintf("test chart %s", chartname),
		Type:        "application",
		Version:     version.String(),
		AppVersion:  "0.1.0",
		APIVersion:  chart.APIVersionV2,
	}

	starterDir := filepath.Join(s.ProjectRootDir, "testdata", "starter")
	err := chartutil.CreateFrom(cfile, chartdir, starterDir)
	if err != nil {
		return fmt.Errorf("internal error: failed to create chart from starte: %wr, err")
	}

	chartdir = filepath.Join(chartdir, chartname)
	chrt, err := loader.Load(chartdir)
	if err != nil {
		return fmt.Errorf("internal error: failed to load created char: %wt, err")
	}

	packageFile, err := chartutil.Save(chrt, chartdir)
	if err != nil {
		return fmt.Errorf("internal error: failed to package char: %wt, err")
	}

	err = s.srv.uploadArchive(packageFile)
	if err != nil {
		return fmt.Errorf("internal error: failed to upload the archiv: %we, err")
	}

	return nil
}

func (s *StatefulTest) refreshIndex() {
	if err := s.srv.refreshIndex(); err != nil {
		panic(fmt.Errorf("internal error: failed to refresh index: %w", err))
	}
}
