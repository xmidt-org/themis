package service

import "go.uber.org/fx"

// ServiceProvideOut describes the components emitted by this package's standard provider
type ServiceProvideOut struct {
	fx.Out

	Registry *Registry
}

// Provide creates the standard components for this package
func Provide() ServiceProvideOut {
	return ServiceProvideOut{
		Registry: new(Registry),
	}
}
