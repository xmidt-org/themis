package token

import (
	"context"
	"fmt"
	"key"
	"random"
	"sync/atomic"

	jwt "github.com/dgrijalva/jwt-go"
)

// Request is a token creation request.  Clients can pass in arbitrary claims, typically things like "iss",
// to merge and override anything set on the factory via configuration.
type Request struct {
	// Claims holds the extra claims to add to tokens.  These claims will override any configured claims in a Factory,
	// but will not override time-based claims such as nbf or exp.
	Claims map[string]interface{}

	// Meta holds metadata about the request, usually garnered from the original HTTP request.  This
	// metadata is available to lower levels of infrastructure used by the Factory.
	Meta map[string]interface{}
}

// Factory is a creation strategy for signed JWT tokens
type Factory interface {
	// NewToken uses a Request to produce a signed JWT token
	NewToken(context.Context, *Request) (string, error)

	// NewClaims returns the claims that this factory would produce in a token,
	// given a specific Request.
	NewClaims(context.Context, *Request) (map[string]interface{}, error)
}

type factory struct {
	method   jwt.SigningMethod
	claimers Claimers

	// pair is an atomic value so that future updates can implement key rotation
	pair atomic.Value
}

func (f *factory) NewToken(ctx context.Context, r *Request) (string, error) {
	claims, err := f.NewClaims(ctx, r)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(f.method, jwt.MapClaims(claims))
	pair := f.pair.Load().(key.Pair)
	token.Header["kid"] = pair.KID()
	return token.SignedString(pair.Sign())
}

func (f *factory) NewClaims(ctx context.Context, r *Request) (map[string]interface{}, error) {
	merged := make(map[string]interface{}, len(r.Claims))
	for _, c := range f.claimers {
		if err := c.Append(ctx, r, merged); err != nil {
			return nil, err
		}
	}

	return merged, nil
}

// RemoteClaims describes a remote HTTP endpoint that can produce claims given the
// metadata from a token request.
type RemoteClaims struct {
	Method string
	URL    string
}

// Descriptor holds the configurable information for a token Factory
type Descriptor struct {
	// Alg is the required JWT signing algorithm to use
	Alg string

	// Key describes the signing key to use
	Key key.Descriptor

	// Claims is an optional map of claims to add to every token emitted by this factory.
	// Any claims here can be overridden by claims within a token Request.
	Claims map[string]interface{}

	// Duration specifies how long the token should be valid for.  An exp claim is set
	// using this duration from the current time if this field is positive.
	Duration string

	// Nonce indicates whether a nonce (jti) should be applied to each token emitted
	// by this factory.
	Nonce bool

	// NotBeforeDelta is a golang duration that determines the nbf field.  If set, this field
	// is parsed and added to the current time at the moment a token is issued.  The result
	// is set as an nbf claim.  Note that the duration may be zero or negative.
	//
	// If unset, then now nbf claim is issued.
	NotBeforeDelta string

	// HeaderClaims maps HTTP headers onto claims in the issued tokens
	HeaderClaims map[string]string

	// ParameterClaims maps query and form parameters onto claims in issued tokens
	ParameterClaims map[string]string

	// MetaHeaders maps HTTP headers onto metadata in token requests
	MetaHeaders map[string]string

	// MetaParameters maps HTTP parameters onto metadata in token requests
	MetaParameters map[string]string

	// Remote specifies an optional external system that takes metadata from a token request
	// and returns a set of claims to be merged into tokens returned by the Factory.  Returned
	// claims from the remote system do not override claims configured on the Factory.
	Remote *RemoteClaims
}

// NewFactory creates a token Factory from a Descriptor.  The supplied Noncer is used if and only
// if d.Nonce is true.  Alternatively, supplying a nil Noncer will disable nonce creation altogether.
// The token's key pair is registered with the given key Registry.
func NewFactory(n random.Noncer, kr key.Registry, d Descriptor) (Factory, error) {
	lc, err := newLocalClaimer(n, d)
	if err != nil {
		return nil, err
	}

	if len(d.Alg) == 0 {
		d.Alg = "RS256"
	}

	f := &factory{
		method: jwt.GetSigningMethod(d.Alg),
		claimers: Claimers{
			lc,
		},
	}

	if f.method == nil {
		return nil, fmt.Errorf("No such signing method: %s", d.Alg)
	}

	if d.Remote != nil {
		rc, err := newRemoteClaimer(d.Remote)
		if err != nil {
			return nil, err
		}

		f.claimers = append(f.claimers, rc)
	}

	pair, err := kr.Register(d.Key)
	if err != nil {
		return nil, err
	}

	f.pair.Store(pair)

	return f, nil
}
