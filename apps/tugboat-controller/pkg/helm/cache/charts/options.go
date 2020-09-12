package charts

import (
	"github.com/go-logr/logr"
	"github.com/object88/tugboat/apps/tugboat-controller/pkg/helm/cache/repos"
	"helm.sh/helm/v3/pkg/cli"
)

type Options struct {
	cachedepth uint
	cachedir   string
	logger     logr.Logger
	repocache  *repos.Cache
	settings   *cli.EnvSettings
}

type OptionFunc func(o *Options) error

func WithCacheDepth(depth uint) OptionFunc {
	return func(o *Options) error {
		o.cachedepth = depth
		return nil
	}
}

func WithCacheDirectory(dir string) OptionFunc {
	return func(o *Options) error {
		o.cachedir = dir
		return nil
	}
}

func WithHelmEnvSettings(settings *cli.EnvSettings) OptionFunc {
	return func(o *Options) error {
		o.settings = settings
		return nil
	}
}

func WithLogger(logger logr.Logger) OptionFunc {
	return func(o *Options) error {
		o.logger = logger
		return nil
	}
}

func WithRepoCache(c *repos.Cache) OptionFunc {
	return func(o *Options) error {
		o.repocache = c
		return nil
	}
}
