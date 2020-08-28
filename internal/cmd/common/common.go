package common

import (
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/object88/tugboat/internal/cmd/cliflags"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type CommonArgs struct {
	Log logr.Logger

	FlagMgr *cliflags.FlagManager
}

func NewCommonArgs() *CommonArgs {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("TUGBOAT")

	ca := &CommonArgs{
		Log:     zapr.NewLogger(zap.NewNop()),
		FlagMgr: cliflags.New(),
	}
	return ca
}

func (ca *CommonArgs) Setup(flags *pflag.FlagSet) {
	ca.FlagMgr.ConfigureVerboseFlag(flags)
}

func (ca *CommonArgs) Evaluate() error {
	var err error
	var z *zap.Logger
	if ca.FlagMgr.Verbose() {
		z, err = zap.NewDevelopment()
	} else {
		z, err = zap.NewProduction()
	}
	if err != nil {
		return err
	}
	ca.Log = zapr.NewLogger(z)

	return nil
}

func (ca *CommonArgs) ReportDuration(cmd *cobra.Command, start time.Time) {
	duration := time.Since(start)

	segments := []string{}
	var f func(c1 *cobra.Command)
	f = func(c1 *cobra.Command) {
		parent := c1.Parent()
		if parent != nil {
			f(parent)
		}
		segments = append(segments, c1.Name())
	}
	f(cmd)

	ca.Log.Info("Executed command", "command", strings.Join(segments, " "), "duration", duration)
}
