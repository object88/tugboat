package cliflags

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

// CLI Flags
const (
	// museumConfigFileKey defines where to load a file with chart museum
	// configurations
	museumConfigFileKey = "museum-config-file"

	// MuseumsKey defines an array of simple museum name/URL pairs:
	// --museums local=http://localhost:8080
	museumsKey = "museums"

	// updateRepositoriesKey indicates whether to update the local repos when
	// running command
	updateRepositoriesKey = "update-repositories"
)

type FlagManager struct {
	// Do not access these directly; properties that are set via environment
	// configs (i.e. `viper.BindEnv`) will not get updated here.
	helmSettings       *cli.EnvSettings
	museumConfigFile   string
	museums            []string
	updateRepositories bool
}

func NewFlagManager() *FlagManager {
	return &FlagManager{}
}

// ConfigureHelm uses the helm packages to add helm-related flags to the CLI.
// This is paired with `HelmEnvSettings`.
func (fl *FlagManager) ConfigureHelm(flags *pflag.FlagSet) {
	fl.helmSettings = cli.New()
	fl.helmSettings.AddFlags(flags)
}

// HelmEnvSettings return the `helm.EnvSettings` struct constructed by the helm
// library to support helm interactions.  Use the `ConfigureHelm` func on the
// FlagManager struct to set up CLI flags to configure the EnvSettings.
func (fl *FlagManager) HelmEnvSettings() *cli.EnvSettings {
	return fl.helmSettings
}

func (fl *FlagManager) ConfigureMuseumConfigFileFlag(flags *pflag.FlagSet) {
	flags.StringVar(&fl.museumConfigFile, museumConfigFileKey, "", "Path to read museum chart configs to add to repositories.yaml")
	viper.BindPFlag(museumConfigFileKey, flags.Lookup(museumConfigFileKey))
	viper.BindEnv(museumConfigFileKey)
}

func (fl *FlagManager) MuseumConfigFile() (string, error) {
	s0 := viper.GetString(museumConfigFileKey)
	s := strings.TrimSpace(s0)
	if s == "" {
		return "", nil
	}

	// We have a path, ensure that it exists.
	s, err := filepath.Abs(s)
	if err != nil {
		return "", err
	}
	fi, err := os.Lstat(s)
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		return "", fmt.Errorf("file path '%s' is a directory", s0)
	}

	return s, nil
}

func (fl *FlagManager) ConfigureMuseumsFlag(flags *pflag.FlagSet) {
	flags.StringArrayVar(&fl.museums, museumsKey, []string{}, "")
	viper.BindPFlag(museumsKey, flags.Lookup(museumsKey))
}

func (fl *FlagManager) Museums() ([]*repo.Entry, error) {
	index := 0
	result := make([]*repo.Entry, len(fl.museums))
	for _, v := range fl.museums {
		split := strings.SplitN(v, "=", 2)

		if len(split) != 2 {
			return nil, fmt.Errorf("flag value '%s' is invalid; must be of shape \"name=repository_address\"", v)
		}

		result[index] = &repo.Entry{
			Name: split[0],
			URL:  split[1],
		}
		index++
	}
	return result, nil
}

func (fl *FlagManager) ConfigureUpdateRepositories(flags *pflag.FlagSet) {
	flags.BoolVar(&fl.updateRepositories, updateRepositoriesKey, false, "Update the helm repositories before running command")
	viper.BindPFlag(updateRepositoriesKey, flags.Lookup(updateRepositoriesKey))
}

func (fl *FlagManager) UpdateRepositories() bool {
	return viper.GetBool(updateRepositoriesKey)
}
