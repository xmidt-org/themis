package main

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func newViper(name string, file string, fs *pflag.FlagSet) (*viper.Viper, error) {
	v := viper.New()
	if len(file) > 0 {
		v.SetConfigFile(file)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath(fmt.Sprintf("$HOME/.%s", name))
		v.AddConfigPath(fmt.Sprintf("/etc/%s", name))
		v.SetConfigName(name)
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix(name)
	v.AutomaticEnv()

	if err := v.BindPFlags(fs); err != nil {
		return nil, err
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return v, nil
}
