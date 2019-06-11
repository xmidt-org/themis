package config

import "github.com/spf13/pflag"

type FlagSetBuilder func(*pflag.FlagSet) error

type FlagSetOptions struct {
	ApplicationName string
	Builders        []FlagSetBuilder
	Arguments       []string
}

func FlagSet(o FlagSetOptions) func() (*pflag.FlagSet, error) {
	return func() (*pflag.FlagSet, error) {
		fs := pflag.NewFlagSet(o.ApplicationName, pflag.ContinueOnError)
		for _, b := range o.Builders {
			if err := b(fs); err != nil {
				return nil, err
			}
		}

		if err := fs.Parse(o.Arguments); err != nil {
			return nil, err
		}

		return fs, nil
	}
}
