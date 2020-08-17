package cliflags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	outputKey  string = "output"
	verboseKey        = "verbose"
)

type FlagManager struct {
	output  string
	verbose bool
}

func New() *FlagManager {
	return &FlagManager{}
}

// CreateOutputFlag adds the `--output` flag to the flagset
func (fl *FlagManager) ConfigureOutputFlag(flags *pflag.FlagSet) {
	annotations := map[string][]string{
		cobra.BashCompCustom: {"__tugboat_get_outputs"},
	}

	var def Output
	flags.StringVar(&fl.output, outputKey, def.String(), Values())
	flg := flags.Lookup(outputKey)
	flg.Annotations = annotations
	viper.BindPFlag(outputKey, flg)
	viper.BindEnv(outputKey)
}

func (fl *FlagManager) ConfigureVerboseFlag(flags *pflag.FlagSet) {
	flags.BoolVarP(&fl.verbose, verboseKey, "v", false, "Emit debug messages")
	viper.BindPFlag(verboseKey, flags.Lookup(verboseKey))
	viper.BindEnv(verboseKey)
}

func (fl *FlagManager) Output() Output {
	var o Output
	if err := o.UnmarshalText([]byte(fl.output)); err != nil {
		o = Unknown
	}
	return o
}

func (fl *FlagManager) Verbose() bool {
	return viper.GetBool(verboseKey)
}
