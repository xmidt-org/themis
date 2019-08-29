package config

import "go.uber.org/fx"

// Optioner is a strategy for producing one or more uber/fx options from an environment.
// This strategy is useful for components that must be created externally from the uber/fx App flow
// of execution, such as loggers and components that are optional based on configuration.
type Optioner func(Environment) fx.Option

// IfKeySet checks if a viper configuration key is set, either in configuration or using a default.
// If the key is set, the given options are returned.  If not, no options are returned into the enclosing container.
func IfKeySet(configKey string, options ...fx.Option) Optioner {
	return func(e Environment) fx.Option {
		if e.Unmarshaller.IsSet(configKey) {
			return fx.Options(options...)
		}

		return fx.Options()
	}
}

// IfKeyNotSet checks if a viper configuration key is set, either in configuration or using a default.
// If the key is not set, the given options are returned.  If it is set, no options are returned into the enclosing container.
func IfKeyNotSet(configKey string, options ...fx.Option) Optioner {
	return func(e Environment) fx.Option {
		if !e.Unmarshaller.IsSet(configKey) {
			return fx.Options(options...)
		}

		return fx.Options()
	}
}
