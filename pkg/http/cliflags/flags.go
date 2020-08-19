package cliflags

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	portKey string = "port"
)

type FlagManager struct {
	port int
}

func New() *FlagManager {
	return &FlagManager{}
}

func (fl *FlagManager) ConfigurePortFlag(flags *pflag.FlagSet) {
	flags.IntVar(&fl.port, portKey, 3000, "http port")
	viper.BindEnv(portKey)
	viper.BindPFlag(portKey, flags.Lookup(portKey))
}

func (fl *FlagManager) Port() int {
	return viper.GetInt(portKey)
}
