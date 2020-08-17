package common

import (
	"strings"

	"github.com/object88/tugboat/internal/cmd/cliflags"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type CommonArgs struct {
	Log *logrus.Logger

	FlagMgr *cliflags.FlagManager
}

func NewCommonArgs() *CommonArgs {
	return &CommonArgs{
		Log:     logrus.New(),
		FlagMgr: cliflags.New(),
	}
}

func (ca *CommonArgs) Setup(flags *pflag.FlagSet) {
}

func (ca *CommonArgs) Evaluate() error {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("TUGBOAT")

	if ca.FlagMgr.Verbose() {
		ca.Log.Level = logrus.DebugLevel
	}

	return nil
}
