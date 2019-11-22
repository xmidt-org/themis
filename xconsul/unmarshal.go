package xconsul

import (
	"github.com/hashicorp/consul/api"
	"github.com/xmidt-org/themis/config"
	"github.com/xmidt-org/themis/service"
	"go.uber.org/fx"
)

const (
	DefaultUnmarshalKey = "consul"
)

// ConsulUnmarshalIn defines all the dependencies required for creating a consul client
type ConsulUnmarshalIn struct {
	fx.In

	// Unmarshaller is the required configuration unmarshaller used to obtain Options instances
	Unmarshaller config.Unmarshaller

	// IDGenerator is the service-layer strategy for generating id's.  If supplied, this component
	// is used to generate service and check ids during registration.  If not supplied, service.DefaultIDGenerator()
	// is used.
	IDGenerator service.IDGenerator `optional:"true"`

	// Checker is an optional component that can report service-layer health.
	// Both this component must be supplied and a TTL configuration section must be set in order
	// for TTL checks to be autoregistered for all services.
	Checker service.Checker `optional:"true"`
}

// Unmarshal specifies all the non-component information for reading configuration to
// produce the components exposed by this package.
type Unmarshal struct {
	// Key is the unmarshal key used to obtain an Options.  If unset, DefaultUnmarshalKey is used.
	Key string

	// Name is the component name for this consul client, useful when more than one consul client
	// is used.  For Provide, this field is unused.  For Annotated, this is the Name field of the returned
	// fx.Annotated struct.  If unset, this field defaults to the configuration key used to unmarshal.
	Name string
}

func (u Unmarshal) key() string {
	if len(u.Key) > 0 {
		return u.Key
	}

	return DefaultUnmarshalKey
}

func (u Unmarshal) name() string {
	if len(u.Name) > 0 {
		return u.Name
	}

	return u.key()
}

// Provide is an fx.Provide function that creates an unnamed consul client.  This method is appropriate
// when only (1) consul client instance is used for a given fx.App.
func (u Unmarshal) Provide(in ConsulUnmarshalIn) (*api.Client, error) {
	var o Options
	if err := in.Unmarshaller.UnmarshalKey(u.key(), &o); err != nil {
		return nil, err
	}

	return New(o)
}

// Annotated returns an fx.Annotated instance for use with fx.Provide.  This method is appropriate when
// multiple consul clients are used within a single fx.App.
func (u Unmarshal) Annotated() fx.Annotated {
	return fx.Annotated{
		Name:   u.name(),
		Target: u.Provide,
	}
}
