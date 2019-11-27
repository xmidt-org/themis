package xconsul

import (
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"
)

// RegistrationOptions stores the set of common configuration parameters for registering services
// using a given consul Client.
type RegistrationOptions struct {
	// Tags is set of tags to append to all services
	Tags []string

	// Meta is a set of metadata to merge into all services
	Meta map[string]string

	// Services is a set of consul registration objects, which can include checks.
	// Any TTL checks are automatically managed in separate goroutines.
	Services []api.AgentServiceRegistration
}

// Options describes the set of configuration options for a consul client.  These fields
// mostly mirror api.Config, but use the xhttpclient package to create the client.
type Options struct {
	Address    string
	Scheme     string
	Datacenter string
	Transport  *xhttpclient.Transport
	HttpAuth   *api.HttpBasicAuth
	WaitTime   time.Duration
	Token      string
	TokenFile  string
	Tls        api.TLSConfig

	Registration RegistrationOptions
}

// New returns a consul client given a set of Options
func New(o Options) (*api.Client, error) {
	httpClient, err := api.NewHttpClient(
		xhttpclient.NewHttpTransport(o.Transport),
		o.Tls,
	)

	if err != nil {
		return nil, err
	}

	config := api.Config{
		Address:    o.Address,
		Scheme:     o.Scheme,
		Datacenter: o.Datacenter,
		HttpClient: httpClient,
		HttpAuth:   o.HttpAuth,
		WaitTime:   o.WaitTime,
		Token:      o.Token,
		TokenFile:  o.TokenFile,
		TLSConfig:  o.Tls,
	}

	return api.NewClient(&config)
}
