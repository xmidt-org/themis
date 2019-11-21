package service

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrRegistryFrozen = errors.New("The service registry cannot accept more service descriptors after startup")
)

const (
	registryOpen uint32 = iota
	registryFrozen
)

// Descriptor holds the common information about a single service endpoint within an application.
// A particular service discovery infrastructure will interpret a Descriptor in specific ways
// outside the scope of this package.
type Descriptor struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
}

// Registry represents the set of self-described services that should be made available via some
// external service discovery infrastructure.  A Registry instance will accept new service descriptors
// until Freeze is invoked.  At that point, Describe will fail with ErrRegistryFrozen.  This enforces
// the constraint that services cannot be described once the application has been started.
type Registry struct {
	state uint32

	lock        sync.Mutex
	descriptors []Descriptor
}

// Frozen returns true if this Registry has been frozen, i.e. if no further descriptors are allowed.
func (r *Registry) Frozen() bool {
	return atomic.LoadUint32(&r.state) == registryFrozen
}

// Describe allows application code to self-describe an externally available service.  A given application
// may have multiple services, e.g. more than one http.Server that should be discoverable.
//
// Once frozen, this method will fail with ErrRegistryFrozen.  No further self-describing services are allowed at that point.
func (r *Registry) Describe(d Descriptor) error {
	if atomic.LoadUint32(&r.state) == registryFrozen {
		return ErrRegistryFrozen
	}

	defer r.lock.Unlock()
	r.lock.Lock()

	if atomic.LoadUint32(&r.state) == registryFrozen {
		return ErrRegistryFrozen
	}

	r.descriptors = append(r.descriptors, d)
	return nil
}

// Freeze ensures that no further service descriptors are allowed and returns a copy of the internal
// set of self-described services.  This method is idempotent.  It can be called multiple times and concurrently,
// in which cases it will return distinct copies of the same slice.
func (r *Registry) Freeze() []Descriptor {
	atomic.StoreUint32(&r.state, registryFrozen)
	defer r.lock.Unlock()
	r.lock.Lock()

	clone := make([]Descriptor, len(r.descriptors))
	copy(clone, r.descriptors)
	return clone
}
