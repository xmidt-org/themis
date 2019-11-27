package service

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

// Registry represents a way for application-layer code to self-describe services at runtime,
// usually during startup.
type Registry interface {
	Describe(Descriptor) error
}
