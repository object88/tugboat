package flags

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// OutputKey determines the output format
	OutputKey = "output"

	// VerboseKey turns on verbose output to STDERR
	VerboseKey = "verbose"
)

// CreateOutputFlag adds the `--output` flag to the flagset
func CreateOutputFlag(flgs *pflag.FlagSet) {
	annotations := map[string][]string{
		cobra.BashCompCustom: []string{"__churl_get_outputs"},
	}

	var def Output
	flgs.String(OutputKey, def.String(), Values())
	flg := flgs.Lookup(OutputKey)
	flg.Annotations = annotations
	viper.BindPFlag(OutputKey, flg)
	viper.BindEnv(OutputKey)
}

// ReadOutputFlag gets the specified output setting, and verifies that it is a
// legitimate value
func ReadOutputFlag() (Output, error) {
	raw := viper.GetString(OutputKey)
	var o Output
	if err := o.UnmarshalText([]byte(raw)); err != nil {
		return Unknown, err
	}
	return o, nil
}
