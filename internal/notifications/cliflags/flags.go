package cliflags

import (
	"net/url"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	listenersKey string = "listeners"
)

type FlagManager struct {
	listeners []string
}

func New() *FlagManager {
	return &FlagManager{}
}

func (fm *FlagManager) ConfigureListenersFlag(flags *pflag.FlagSet) {
	flags.StringSliceVar(&fm.listeners, listenersKey, nil, "URLs to services implementing a Listener interface")
	viper.BindEnv(listenersKey)
	viper.BindPFlag(listenersKey, flags.Lookup(listenersKey))
}

func (fl *FlagManager) Listeners() ([]*url.URL, error) {
	raw := viper.GetStringSlice(listenersKey)
	result := make([]*url.URL, len(raw))
	var err error
	for k, v := range raw {
		result[k], err = url.Parse(v)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
