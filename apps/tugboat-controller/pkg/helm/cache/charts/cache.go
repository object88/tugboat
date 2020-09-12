package charts

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/apis/engineering.tugboat/v1alpha1"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	DefaultCacheDepth uint = 128
)

type Cache struct {
	logger    logr.Logger
	repocache *repos.Cache

	cachedir string

	lru     *simplelru.LRU
	lrulock sync.Mutex

	settings *cli.EnvSettings
}

// New returns a new pointer to a new instance of the Cache struct
func New() *Cache {
	return &Cache{}
}

func (c *Cache) Connect(opts ...OptionFunc) error {
	co := Options{}
	for _, opt := range opts {
		if err := opt(&co); err != nil {
			return fmt.Errorf("failed to configure options: %w", err)
		}
	}

	if co.logger == nil {
		z, err := zap.NewProduction()
		if err != nil {
			return fmt.Errorf("no logger provided and failed to create production looger: %w", err)
		}
		co.logger = zapr.NewLogger(z)
	}

	// This is a very simple implementation
	var err error
	if c.lru, err = simplelru.NewLRU(int(co.cachedepth), c.evict); err != nil {
		return err
	}

	c.cachedir = co.cachedir
	c.logger = co.logger
	c.repocache = co.repocache
	c.settings = co.settings

	if c.repocache == nil {
		return fmt.Errorf("missing required repo cache")
	}

	return nil
}

func (c *Cache) Unpack(ref v1alpha1.ChartReference) (string, func() error, error) {
	noop := func() error { return nil }

	source, err := c.Get(ref.Repository, ref.Chart, ref.Version)
	if err != nil {
		return "", noop, fmt.Errorf("failed to get chart tarbar: %w", err)
	}

	unpackdir := filepath.Join(c.cachedir, "unpacked", uuid.New().String())
	if err := os.MkdirAll(unpackdir, 0755); err != nil {
		return "", noop, fmt.Errorf("failed to create unpack dir at '%s': %w", unpackdir, err)
	}

	if err := chartutil.ExpandFile(unpackdir, source); err != nil {
		return "", noop, fmt.Errorf("failed to expand tarball '%s' to '%s': %w", unpackdir, source, err)
	}

	f := func() error {
		return os.RemoveAll(unpackdir)
	}

	return filepath.Join(unpackdir, ref.Chart), f, nil
}

func (c *Cache) Get(reponame string, chartname string, version *semver.Version) (string, error) {
	e, err := c.repocache.GetChartRepository(reponame)
	if err != nil {
		return "", fmt.Errorf("did not get repo.Entry for '%s': %w", reponame, err)
	}

	c.lrulock.Lock()
	defer c.lrulock.Unlock()

	destination, err := c.foo(e, chartname, version)
	if err != nil {
		return "", err
	}

	if _, ok := c.lru.Get(destination); ok {
		return destination, nil
	}

	if err := c.download(destination, e, chartname, version); err != nil {
		return "", err
	}

	c.lru.Add(destination, struct{}{})

	return destination, nil
}

func (c *Cache) download(destination string, r *repo.Entry, chartname string, version *semver.Version) error {
	cv, err := c.repocache.GetChartVersion(r.Name, chartname, version)
	if err != nil {
		return err
	}

	// Get the repo URL from the repo
	repoURL, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	// Get the chart URL
	u, err := url.Parse(cv.URLs[0])
	if err != nil {
		return err
	}

	// We need a trailing slash for ResolveReference to work, but make sure there isn't already one
	repoURL.Path = strings.TrimSuffix(repoURL.Path, "/") + "/"
	u = repoURL.ResolveReference(u)
	u.RawQuery = repoURL.Query().Encode()

	g, err := getter.All(c.settings).ByScheme(u.Scheme)
	if err != nil {
		return err
	}

	buf, err := g.Get(
		u.String(),
		getter.WithBasicAuth(r.Username, r.Password),
		getter.WithTLSClientConfig(r.CertFile, r.KeyFile, r.CAFile),
		getter.WithInsecureSkipVerifyTLS(r.InsecureSkipTLSverify),
	)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("downloaded chart '%s' but could not open file to write tarball: %w", destination, err)
	}
	f.Write(buf.Bytes())

	return nil
}

func (c *Cache) evict(key interface{}, value interface{}) {
	destination, ok := key.(string)
	if !ok {
		return
	}
	os.Remove(destination)
}

func (c *Cache) foo(r *repo.Entry, chartname string, version *semver.Version) (string, error) {
	filename := fmt.Sprintf("%s-%s.tgz", chartname, version)
	fullpath := filepath.Join(c.cachedir, r.Name, chartname)
	if err := os.MkdirAll(fullpath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory '%s': %w", fullpath, err)
	}
	fullname := filepath.Join(fullpath, filename)
	return fullname, nil
}
