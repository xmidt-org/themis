package service

// CheckState represents the healthiness of a self-described application service
// that is available for external discovery.
type CheckState int

const (
	Pass CheckState = iota
	Warn
	Fail
)

// CheckStatus represents the health status of a single self-described service
type CheckStatus struct {
	State   CheckState
	Message string
}

// Checker is a strategy interface for checking service health.  In general, all self-described
// services will share a single instance of this interface.  This interface is consumed by service discovery
// infrastructure, but implemented by health infrastructure.
type Checker interface {
	Check() CheckStatus
}
