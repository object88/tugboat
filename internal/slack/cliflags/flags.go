package cliflags

import (
	"github.com/object88/tugboat/internal/slack/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// CLI Flags
const (
	// signingSecretKey verifies that a request has come from Slack
	signingSecretKey = "slack-signing-secret"

	// tokenKey provides the slack token
	tokenKey = "slack-token"

	verificationKey = "slack-verification"
)

// FlagManager maintains the state of Slack-related CLI flags
type FlagManager struct {
	// Do not access these directly; properties that are set via environment
	// configs (i.e. `viper.BindEnv`) will not get updated here.
	signingSecret string
	token         string
	verification  string
}

// New returns a new instance of FlagManager
func New() *FlagManager {
	return &FlagManager{}
}

func (fl *FlagManager) ConfigureFlags(flags *pflag.FlagSet) {
	flags.StringVar(&fl.signingSecret, signingSecretKey, "", "slack signing secret")
	viper.BindEnv(signingSecretKey)
	viper.BindPFlag(signingSecretKey, flags.Lookup(signingSecretKey))

	flags.StringVar(&fl.token, tokenKey, "", "slack token")
	viper.BindEnv(tokenKey)
	viper.BindPFlag(tokenKey, flags.Lookup(tokenKey))

	flags.StringVar(&fl.verification, verificationKey, "", "slack verification")
	viper.BindEnv(verificationKey)
	viper.BindPFlag(verificationKey, flags.Lookup(verificationKey))
}

func (fl *FlagManager) Config() config.Config {
	return config.Config{
		SigningSecret: viper.GetString(signingSecretKey),
		Token:         viper.GetString(tokenKey),
		Verification:  viper.GetString(verificationKey),
	}
}
