package service

import "go.uber.org/fx"

// ServiceProvideOut describes the components emitted by this package's standard provider
type ServiceProvideOut struct {
	fx.Out

	IDGenerator IDGenerator
	Registry    *Registry
}

// Provide creates the standard components for this package
func Provide() ServiceProvideOut {
	return ServiceProvideOut{
		IDGenerator: Base64IDGenerator{},
		Registry:    new(Registry),
	}
}
