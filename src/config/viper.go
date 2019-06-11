package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ViperBuilder func(*viper.Viper) error

type ViperOptions struct {
	ApplicationName string
	FileFlag        string
	Builders        []ViperBuilder
}

func Viper(o ViperOptions) func(*pflag.FlagSet) (*viper.Viper, error) {
	return func(fs *pflag.FlagSet) (*viper.Viper, error) {
		var (
			v  = viper.New()
			ff = fs.Lookup(o.FileFlag)
		)

		if ff != nil {
			v.SetConfigFile(ff.Value.String())
		} else {
			v.AddConfigPath(".")
			v.AddConfigPath(fmt.Sprintf("$HOME/.%s", o.ApplicationName))
			v.AddConfigPath(fmt.Sprintf("/etc/%s", o.ApplicationName))
			v.SetConfigName(o.ApplicationName)
		}

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.SetEnvPrefix(o.ApplicationName)
		v.AutomaticEnv()

		for _, b := range o.Builders {
			if err := b(v); err != nil {
				return nil, err
			}
		}

		if err := v.BindPFlags(fs); err != nil {
			return nil, err
		}

		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}

		return v, nil
	}
}
