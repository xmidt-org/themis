package token

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/xmidt-org/themis/random"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

var (
	ErrRemoteURLRequired = errors.New("A URL for the remote claimer is required")
	ErrMissingKey        = errors.New("A key is required for all claims and metadata values")
)

// ClaimBuilder is a strategy for building token claims, given a token Request
type ClaimBuilder interface {
	AddClaims(context.Context, *Request, map[string]interface{}) error
}

type ClaimBuilderFunc func(context.Context, *Request, map[string]interface{}) error

func (cbf ClaimBuilderFunc) AddClaims(ctx context.Context, tr *Request, target map[string]interface{}) error {
	return cbf(ctx, tr, target)
}

// ClaimBuilders implements a pipeline of ClaimBuilder instances, invoked in sequence.
type ClaimBuilders []ClaimBuilder

func (cbs ClaimBuilders) AddClaims(ctx context.Context, r *Request, target map[string]interface{}) error {
	for _, e := range cbs {
		if err := e.AddClaims(ctx, r, target); err != nil {
			return err
		}
	}

	return nil
}

// requestClaimBuilder is a ClaimBuilder that copies the Request.Claims
type requestClaimBuilder struct{}

func (rc requestClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) error {
	for k, v := range r.Claims {
		target[k] = v
	}

	return nil
}

// staticClaimBuilder is a ClaimBuilder that simply appends a constant set of claims
type staticClaimBuilder map[string]interface{}

func (sc staticClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) error {
	for k, v := range sc {
		target[k] = v
	}

	return nil
}

// timeClaimBuilder is a ClaimBuilder which handles time-based claims
type timeClaimBuilder struct {
	now              func() time.Time
	duration         time.Duration
	disableNotBefore bool
	notBeforeDelta   time.Duration
}

func (tc *timeClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) error {
	now := tc.now().UTC()
	target["iat"] = now.Unix()

	if tc.duration > 0 {
		target["exp"] = now.Add(tc.duration).Unix()
	}

	if !tc.disableNotBefore {
		target["nbf"] = now.Add(tc.notBeforeDelta).Unix()
	}

	return nil
}

// nonceClaimBuilder is a ClaimBuilder that appends a nonce (jti) claim
type nonceClaimBuilder struct {
	n random.Noncer
}

func (nc nonceClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) error {
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
type remoteClaimBuilder struct {
	endpoint endpoint.Endpoint
	url      string
	extra    map[string]interface{}
}

func (rc *remoteClaimBuilder) AddClaims(ctx context.Context, r *Request, target map[string]interface{}) error {
	metadata := r.Metadata
	if len(rc.extra) > 0 {
		metadata = make(map[string]interface{}, len(r.Metadata)+len(rc.extra))
		for k, v := range r.Metadata {
			metadata[k] = v
		}

		for k, v := range rc.extra {
			metadata[k] = v
		}
	}

	result, err := rc.endpoint(ctx, metadata)
	if err != nil {
		return err
	}

	for k, v := range result.(map[string]interface{}) {
		target[k] = v
	}

	return nil
}

func newRemoteClaimBuilder(client xhttpclient.Interface, metadata map[string]interface{}, r *RemoteClaims) (*remoteClaimBuilder, error) {
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

	if client == nil {
		client = new(http.Client)
	}

	c := kithttp.NewClient(
		method,
		url,
		kithttp.EncodeJSONRequest,
		DecodeRemoteClaimsResponse,
		kithttp.SetClient(client),
		kithttp.ClientBefore(
			kithttp.SetRequestHeader("Content-Type", "application/json"),
		),
	)

	return &remoteClaimBuilder{endpoint: c.Endpoint(), url: r.URL, extra: metadata}, nil
}

// NewClaimBuilders constructs a ClaimBuilders from configuration.  The returned instance is typically
// used in configuration a token Factory.  It can be used as a standalone service component with an endpoint.
//
// The returned builders do not include those claims derived from HTTP requests.  Claims derived from HTTP
// requests are handled by NewRequestBuilders and DecodeServerRequest.
func NewClaimBuilders(n random.Noncer, client xhttpclient.Interface, o Options) (ClaimBuilders, error) {
	var (
		// at a minimum, the claims from the request will be copied
		builders           = ClaimBuilders{requestClaimBuilder{}}
		staticClaimBuilder = make(staticClaimBuilder)
	)

	if o.Remote != nil {
		// scan the metadata looking for static values that should be applied when invoking the remote server
		metadata := make(map[string]interface{})
		for _, value := range o.Metadata {
			switch {
			case len(value.Key) == 0:
				return nil, ErrMissingKey

			case value.IsFromHTTP():
				continue

			case !value.IsStatic():
				return nil, fmt.Errorf("A value is required for the static metadata: %s", value.Key)

			default:
				msg, err := value.RawMessage()
				if err != nil {
					return nil, err
				}

				metadata[value.Key] = msg
			}
		}

		remoteClaimBuilder, err := newRemoteClaimBuilder(client, metadata, o.Remote)
		if err != nil {
			return nil, err
		}

		builders = append(builders, remoteClaimBuilder)
	}

	for _, value := range o.Claims {
		switch {
		case len(value.Key) == 0:
			return nil, ErrMissingKey

		case value.IsFromHTTP():
			continue

		case !value.IsStatic():
			return nil, fmt.Errorf("A value is required for the static claim: %s", value.Key)

		default:
			msg, err := value.RawMessage()
			if err != nil {
				return nil, err
			}

			staticClaimBuilder[value.Key] = msg
		}
	}

	if len(staticClaimBuilder) > 0 {
		builders = append(builders, staticClaimBuilder)
	}

	if o.Nonce && n != nil {
		builders = append(builders, nonceClaimBuilder{n: n})
	}

	if !o.DisableTime {
		builders = append(
			builders,
			&timeClaimBuilder{
				now:              time.Now,
				duration:         o.Duration,
				disableNotBefore: o.DisableNotBefore,
				notBeforeDelta:   o.NotBeforeDelta,
			},
		)
	}

	return builders, nil
}
