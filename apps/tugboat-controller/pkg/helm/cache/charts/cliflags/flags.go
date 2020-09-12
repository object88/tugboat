package cliflags

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/charts"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cacheDepthKey     string = "cache-depth"
	cacheDirectoryKey        = "cache-directory"
)

type FlagManager struct {
	cacheDepth            uint
	cacheDirectory        string
	checkedCacheDirectory bool
}

func New() *FlagManager {
	return &FlagManager{}
}

func (fl *FlagManager) ConfigureCacheDepthFlag(flags *pflag.FlagSet) {
	flags.UintVar(&fl.cacheDepth, cacheDepthKey, charts.DefaultCacheDepth, "Number of chart tarballs to cache")
	viper.BindPFlag(cacheDepthKey, flags.Lookup(cacheDepthKey))
	viper.BindEnv(cacheDepthKey)
}

// ConfigureCacheDirectoryFlag is the location of the chart cache
func (fl *FlagManager) ConfigureCacheDirectoryFlag(flags *pflag.FlagSet) {
	flags.StringVar(&fl.cacheDirectory, cacheDirectoryKey, "", "Location for cached chart files.  If left blank, controller will create a directory under the TMP directory")
	viper.BindPFlag(cacheDirectoryKey, flags.Lookup(cacheDirectoryKey))
	viper.BindEnv(cacheDirectoryKey)
}

func (fl *FlagManager) CacheDepth() uint {
	return viper.GetUint(cacheDepthKey)
}

// CacheDirectory gets the directory to use to cache chart tarballs. If no
// flag was specified, a directory will be generated.
func (fl *FlagManager) CacheDirectory() (string, error) {
	if fl.checkedCacheDirectory {
		return fl.cacheDirectory, nil
	}

	var err error
	p := viper.GetString(cacheDirectoryKey)
	if p == "" {
		// Ensure that we have an appropriate temporary directory to work in
		if err := os.MkdirAll(os.TempDir(), 0777); err != nil {
			return "", fmt.Errorf("failed to create temporary directory: %w", err)
		}
		p, err = ioutil.TempDir(os.TempDir(), "chartcache-****")
		if err != nil {
			return "", fmt.Errorf("failed to get temp directory for chart cache directory: %w", err)
		}
	}

	// Ensure that we have a clean path.
	p, err = filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for chart cache directory: %w", err)
	}

	// Ensure that we have a directory.
	fi, err := os.Stat(p)
	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(p, 0644); err != nil {
			return "", fmt.Errorf("failed to create chart cache directory at '%s': %w", p, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to stat chart cache directory: %w", err)
	} else if !fi.IsDir() {
		return "", fmt.Errorf("chart cache directory at '%s' is a file", p)
	}

	fl.cacheDirectory = p
	fl.checkedCacheDirectory = true
	return p, nil
}
