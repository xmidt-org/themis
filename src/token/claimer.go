package token

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"random"
	"time"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

var (
	ErrRemoteURLRequired  = errors.New("A URL for the remote claimer is required")
	ErrClaimValueRequired = errors.New("A value is required for static claims")
)

// Claimer represents a strategy for obtaining claims, typically through configuration
// or from some remote system.
type Claimer interface {
	Append(context.Context, *Request, map[string]interface{}) error
}

// requestClaimer is a Claimer that copies the Request.Claims
type requestClaimer struct{}

func (rc requestClaimer) Append(_ context.Context, r *Request, target map[string]interface{}) error {
	for k, v := range r.Claims {
		target[k] = v
	}

	return nil
}

// staticClaimer is a Claimer that simply appends a constant set of claims
type staticClaimer map[string]interface{}

func (sc staticClaimer) Append(_ context.Context, r *Request, target map[string]interface{}) error {
	for k, v := range sc {
		target[k] = v
	}

	return nil
}

// timeClaimer is a Claimer which handles time-based claims
type timeClaimer struct {
	now            func() time.Time
	duration       time.Duration
	notBeforeDelta *time.Duration
}

func (tc *timeClaimer) Append(_ context.Context, r *Request, target map[string]interface{}) error {
	now := tc.now().UTC()
	target["iat"] = now.Unix()

	if tc.duration > 0 {
		target["exp"] = now.Add(tc.duration).Unix()
	}

	if tc.notBeforeDelta != nil {
		target["nbf"] = now.Add(*tc.notBeforeDelta).Unix()
	}

	return nil
}

// nonceClaimer is a Claimer that appends a nonce (jti) claim
type nonceClaimer struct {
	n random.Noncer
}

func (nc nonceClaimer) Append(_ context.Context, r *Request, target map[string]interface{}) error {
	nonce, err := nc.n.Nonce()
	if err != nil {
		return err
	}

	target["jti"] = nonce
	return nil
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// remoteClaimer invokes a remote system to obtain claims.  The metadata from a token request
// is passed as the payload.
type remoteClaimer endpoint.Endpoint

func (rc remoteClaimer) Append(ctx context.Context, r *Request, target map[string]interface{}) error {
	result, err := rc(ctx, r.Metadata)
	if err != nil {
		return err
	}

	for k, v := range result.(map[string]interface{}) {
		target[k] = v
	}

	return nil
}

func newRemoteClaimer(r *RemoteClaims) (remoteClaimer, error) {
	if r == nil {
		return nil, nil
	}

	if len(r.URL) == 0 {
		return nil, ErrRemoteURLRequired
	}

	url, err := url.Parse(r.URL)
	if err != nil {
		return nil, err
	}

	method := r.Method
	if len(method) == 0 {
		method = http.MethodPost
	}

	c := kithttp.NewClient(
		method,
		url,
		kithttp.EncodeJSONRequest,
		DecodeRemoteClaimsResponse,
		kithttp.SetClient(new(http.Client)),
	)

	return remoteClaimer(c.Endpoint()), nil
}

// Claimers represents a set of Claimer strategies that are invoked in sequence.  A Claimers
// is an aggregate Claimer strategy.
type Claimers []Claimer

func (cs Claimers) Append(ctx context.Context, r *Request, target map[string]interface{}) error {
	for _, e := range cs {
		if err := e.Append(ctx, r, target); err != nil {
			return err
		}
	}

	return nil
}

// NewClaimers produces a slice of Claimer strategies that produces the basic claims
// defined in the Options.  HTTP-based claims are skipped by this function.
func NewClaimers(n random.Noncer, o Options) (Claimers, error) {
	var (
		claimers      = Claimers{requestClaimer{}}
		staticClaimer = make(staticClaimer)
		timeClaimer   = &timeClaimer{
			now: time.Now,
		}
	)

	if o.Remote != nil {
		remoteClaimer, err := newRemoteClaimer(o.Remote)
		if err != nil {
			return nil, err
		}

		claimers = append(claimers, remoteClaimer)
	}

	for name, value := range o.Claims {
		if value.IsHttp() {
			continue
		}

		if value.Value == nil {
			return nil, ErrClaimValueRequired
		}

		staticClaimer[name] = value.Value
	}

	if len(staticClaimer) > 0 {
		claimers = append(claimers, staticClaimer)
	}

	if o.Nonce && n != nil {
		claimers = append(claimers, nonceClaimer{n: n})
	}

	if len(o.Duration) > 0 {
		var err error
		timeClaimer.duration, err = time.ParseDuration(o.Duration)
		if err != nil {
			return nil, err
		}
	}

	if len(o.NotBeforeDelta) > 0 {
		var err error
		timeClaimer.notBeforeDelta = new(time.Duration)
		*timeClaimer.notBeforeDelta, err = time.ParseDuration(o.NotBeforeDelta)
		if err != nil {
			return nil, err
		}
	}

	claimers = append(claimers, timeClaimer)
	return claimers, nil
}
