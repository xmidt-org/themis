package main

import "github.com/spf13/pflag"

const (
	FileFlag      = "file"
	FileFlagShort = "f"
	FileFlagUsage = "the configuration file to use.  If unset, a standard search will be used to find the configuration file."
)

func parseCommandLine(name string, arguments []string) (*pflag.FlagSet, error) {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.StringP(FileFlag, FileFlagShort, "", FileFlagUsage)

	if err := fs.Parse(arguments); err != nil {
		return nil, err
	}

	return fs, nil
}
