package xconsul

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
)

const (
	RegisterNotYet uint32 = iota
	RegisterSuccess
	RegisterFailure
)

type RegistrationStatus struct {
	state       uint32
	registerErr error
}

func (rs *RegistrationStatus) updateRegistrationStatus(err error) {
}

type TTLUpdater interface {
	UpdateTTL(checkID, output, status string) error
}

type ServiceRegisterer interface {
	ServiceRegister(*api.AgentServiceRegistration) error
	ServiceDeregister(string) error
}

type Registrar struct {
	logger     log.Logger
	ttlUpdater TTLUpdater
	registerer ServiceRegisterer
	services   map[string]api.AgentServiceRegistration
}

func (r *Registrar) register(service *api.AgentServiceRegistration, bo backoff.BackOff, status *RegistrationStatus) error {
	registerErr := backoff.RetryNotify(
		func() error {
			return r.registerer.ServiceRegister(service)
		},
		bo,
		func(err error, d time.Duration) {
			// status.updateRegistrationStatus(err)
		},
	)

	if registerErr == nil {
		// status.updateRegistrationStatus(nil)
		bo.Reset()

		// TODO: start the TTL check
	}

	return registerErr
}

func (r *Registrar) updateTTL(checkID string) {
}
