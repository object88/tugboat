package repos

import (
	"time"

	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/cli"
)

type HelmOptions struct {
	cooldown  time.Duration
	driver    string
	logger    logr.Logger
	namespace string
	settings  *cli.EnvSettings
	timeout   time.Duration
}

type OptionFunc func(h *HelmOptions) error

func WithCooldown(cooldown time.Duration) OptionFunc {
	return func(h *HelmOptions) error {
		h.cooldown = cooldown
		return nil
	}
}

func WithDriver(driver string) OptionFunc {
	return func(h *HelmOptions) error {
		h.driver = driver
		return nil
	}
}

func WithHelmEnvSettings(settings *cli.EnvSettings) OptionFunc {
	return func(h *HelmOptions) error {
		h.settings = settings
		return nil
	}
}

func WithLogger(logger logr.Logger) OptionFunc {
	return func(h *HelmOptions) error {
		h.logger = logger
		return nil
	}
}

func WithNamespace(namespace string) OptionFunc {
	return func(h *HelmOptions) error {
		h.namespace = namespace
		return nil
	}
}

func WithTimeout(timeout time.Duration) OptionFunc {
	return func(h *HelmOptions) error {
		h.timeout = timeout
		return nil
	}
}
