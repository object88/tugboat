package cliflags

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	portKey string = "grpc-port"
)

type FlagManager struct {
	port uint
}

func New() *FlagManager {
	return &FlagManager{}
}

func (fm *FlagManager) ConfigureGrpcPortFlag(flags *pflag.FlagSet) {
	flags.UintVar(&fm.port, portKey, 5678, "Port for GRPC communications")
	viper.BindEnv(portKey)
	viper.BindPFlag(portKey, flags.Lookup(portKey))
}

func (fm *FlagManager) GRPCPort() uint {
	return viper.GetUint(portKey)
}
