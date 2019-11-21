package service

// HealthState represents the healthiness of a self-described application service
// that is available for external discovery.
type HealthState int

const (
	Pass HealthState = iota
	Warn
	Fail
)

// Status represents the health status of a single self-described service
type Status struct {
	State   HealthState
	Message string
}

// HealthChecker is a strategy interface for checking service health.  In general, all self-described
// services will share a single instance of this interface.  This interface is consumed by service discovery
// infrastructure, but implemented by health infrastructure.
type HealthChecker interface {
	Check() Status
}
