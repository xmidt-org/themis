package xconsul

import (
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"
)

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
