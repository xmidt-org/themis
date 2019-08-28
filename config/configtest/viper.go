package configtest

import (
	"strings"

	"github.com/spf13/viper"
)

// TB is the method set provided by the testing types
type TB interface {
	Fatalf(string, ...interface{})
}

// LoadJson is a convenience for initializing a Viper instance with a JSON configuration.
// If any errors occur, the surrounding test is failed.
func LoadJson(t TB, json string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("json")
	if err := v.ReadConfig(strings.NewReader(json)); err != nil {
		t.Fatalf("Unable to load JSON: %s", err)
	}

	return v
}

// LoadYaml is a convenience for initializing a Viper instance with a JSON configuration.
// If any errors occur, the surrounding test is failed.
func LoadYaml(t TB, yaml string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(yaml)); err != nil {
		t.Fatalf("Unable to load YAML: %s", err)
	}

	return v
}
