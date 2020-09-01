package cliflags

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	httpPortKey      string = "http-port"
	httpsCertFileKey        = "https-cert-file"
	httpsKeyFileKey         = "https-key-file"
	httpsPortKey            = "https-port"
)

type FlagManager struct {
	httpPort      int
	httpsCertFile string
	httpsKeyFile  string
	httpsPort     int
}

func New() *FlagManager {
	return &FlagManager{}
}

func (fl *FlagManager) ConfigureHttpFlag(flags *pflag.FlagSet) {
	flags.IntVar(&fl.httpPort, httpPortKey, 3000, "HTTP port")
	viper.BindEnv(httpPortKey)
	viper.BindPFlag(httpPortKey, flags.Lookup(httpPortKey))
}

func (fl *FlagManager) ConfigureHttpsFlags(flags *pflag.FlagSet) {
	flags.StringVar(&fl.httpsCertFile, httpsCertFileKey, "", "path to certificate file for HTTPS")
	viper.BindEnv(httpsCertFileKey)
	viper.BindPFlag(httpsCertFileKey, flags.Lookup(httpsCertFileKey))

	flags.StringVar(&fl.httpsKeyFile, httpsKeyFileKey, "", "path to key file for HTTPS")
	viper.BindEnv(httpsKeyFileKey)
	viper.BindPFlag(httpsKeyFileKey, flags.Lookup(httpsKeyFileKey))

	flags.IntVar(&fl.httpsPort, httpsPortKey, 0, "HTTPS port")
	viper.BindEnv(httpsPortKey)
	viper.BindPFlag(httpsPortKey, flags.Lookup(httpsPortKey))
}

func (fl *FlagManager) HttpPort() int {
	return viper.GetInt(httpPortKey)
}

func (fl *FlagManager) HttpsCertFile() (string, error) {
	p := viper.GetString(httpsCertFileKey)
	if p == "" {
		return p, nil
	}
	cleanP := path.Clean(p)
	if fi, err := os.Stat(cleanP); err != nil {
		return "", err
	} else if fi.IsDir() {
		return "", fmt.Errorf("Certificate file '%s' is a directory", p)
	}
	return cleanP, nil
}

func (fl *FlagManager) HttpsKeyFile() (string, error) {
	p := viper.GetString(httpsKeyFileKey)
	if p == "" {
		return p, nil
	}
	cleanP := path.Clean(p)
	if fi, err := os.Stat(cleanP); err != nil {
		return "", err
	} else if fi.IsDir() {
		return "", fmt.Errorf("Key file '%s' is a directory", p)
	}
	return cleanP, nil
}

func (fl *FlagManager) HttpsPort() int {
	return viper.GetInt(httpsPortKey)
}
