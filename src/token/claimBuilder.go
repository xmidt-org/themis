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

// ClaimBuilder is a strategy for building token claims, given a token Request
type ClaimBuilder interface {
	Append(context.Context, *Request, map[string]interface{}) error
}

type ClaimBuilderFunc func(context.Context, *Request, map[string]interface{}) error

func (cbf ClaimBuilderFunc) Append(ctx context.Context, tr *Request, target map[string]interface{}) error {
	return cbf(ctx, tr, target)
}

// ClaimBuilders implements a pipeline of ClaimBuilder instances, invoked in sequence.
type ClaimBuilders []ClaimBuilder

func (cbs ClaimBuilders) Append(ctx context.Context, r *Request, target map[string]interface{}) error {
	for _, e := range cbs {
		if err := e.Append(ctx, r, target); err != nil {
			return err
		}
	}

	return nil
}

// requestClaimBuilder is a ClaimBuilder that copies the Request.Claims
type requestClaimBuilder struct{}

func (rc requestClaimBuilder) Append(_ context.Context, r *Request, target map[string]interface{}) error {
	for k, v := range r.Claims {
		target[k] = v
	}

	return nil
}

// staticClaimBuilder is a ClaimBuilder that simply appends a constant set of claims
type staticClaimBuilder map[string]interface{}

func (sc staticClaimBuilder) Append(_ context.Context, r *Request, target map[string]interface{}) error {
	for k, v := range sc {
		target[k] = v
	}

	return nil
}

// timeClaimBuilder is a ClaimBuilder which handles time-based claims
type timeClaimBuilder struct {
	now            func() time.Time
	duration       time.Duration
	notBeforeDelta *time.Duration
}

func (tc *timeClaimBuilder) Append(_ context.Context, r *Request, target map[string]interface{}) error {
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

// nonceClaimBuilder is a ClaimBuilder that appends a nonce (jti) claim
type nonceClaimBuilder struct {
	n random.Noncer
}

func (nc nonceClaimBuilder) Append(_ context.Context, r *Request, target map[string]interface{}) error {
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

// remoteClaimBuilder invokes a remote system to obtain claims.  The metadata from a token request
// is passed as the payload.
type remoteClaimBuilder endpoint.Endpoint

func (rc remoteClaimBuilder) Append(ctx context.Context, r *Request, target map[string]interface{}) error {
	result, err := rc(ctx, r.Metadata)
	if err != nil {
		return err
	}

	for k, v := range result.(map[string]interface{}) {
		target[k] = v
	}

	return nil
}

func newRemoteClaimBuilder(r *RemoteClaims) (remoteClaimBuilder, error) {
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
		kithttp.ClientBefore(
			kithttp.SetRequestHeader("Content-Type", "application/json"),
		),
	)

	return remoteClaimBuilder(c.Endpoint()), nil
}

// NewClaimBuilders constructs a ClaimBuilders from configuration.  The returned instance is typically
// used in configuration a token Factory.  It can be used as a standalone service component with an endpoint.
func NewClaimBuilders(n random.Noncer, o Options) (ClaimBuilders, error) {
	var (
		builders           = ClaimBuilders{requestClaimBuilder{}}
		staticClaimBuilder = make(staticClaimBuilder)
		timeClaimBuilder   = &timeClaimBuilder{
			now: time.Now,
		}
	)

	if o.Remote != nil {
		remoteClaimBuilder, err := newRemoteClaimBuilder(o.Remote)
		if err != nil {
			return nil, err
		}

		builders = append(builders, remoteClaimBuilder)
	}

	for name, value := range o.Claims {
		if len(value.Header) != 0 || len(value.Parameter) != 0 || len(value.Variable) != 0 {
			// skip any claims derived from HTTP requests
			continue
		}

		if value.Value == nil {
			return nil, ErrClaimValueRequired
		}

		staticClaimBuilder[name] = value.Value
	}

	if len(staticClaimBuilder) > 0 {
		builders = append(builders, staticClaimBuilder)
	}

	if o.Nonce && n != nil {
		builders = append(builders, nonceClaimBuilder{n: n})
	}

	if len(o.Duration) > 0 {
		var err error
		timeClaimBuilder.duration, err = time.ParseDuration(o.Duration)
		if err != nil {
			return nil, err
		}
	}

	if len(o.NotBeforeDelta) > 0 {
		var err error
		timeClaimBuilder.notBeforeDelta = new(time.Duration)
		*timeClaimBuilder.notBeforeDelta, err = time.ParseDuration(o.NotBeforeDelta)
		if err != nil {
			return nil, err
		}
	}

	builders = append(builders, timeClaimBuilder)
	return builders, nil
}
