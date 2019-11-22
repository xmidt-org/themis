package xconsul

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"
)

// GenerateID creates a random, base64-encoded identifier.  Used when an ID is not supplied.
func GenerateID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:0]); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(id[0:]), nil
}

// TTL describes how to setup a TTL check for services registered through a given consul client
type TTL struct {
	// ID is the check ID for the registered TTL check
	ID string

	// Name is the name of the registered TTL check
	Name string

	// Interval is registered TTL interval.  This is also the interval at which the TTL goroutine will run.
	Interval time.Duration
}

// RegistrationOptions stores the set of common configuration parameters for registering services
// using a given consul Client.
type RegistrationOptions struct {
	DeregisterCriticalServiceAfter time.Duration

	TTL *TTL

	Tags []string
	Meta map[string]string
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
