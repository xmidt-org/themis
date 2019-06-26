package token

import (
	"fmt"
	"key"
	"random"
	"sync/atomic"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Request is a token creation request.  Clients can pass in arbitrary claims, typically things like "iss",
// to merge and override anything set on the factory via configuration.
type Request struct {
	Claims map[string]interface{}
}

// Factory is a creation strategy for signed JWT tokens
type Factory interface {
	// NewToken uses a Request to produce a signed JWT token
	NewToken(Request) (string, error)
}

type factory struct {
	method jwt.SigningMethod
	claims map[string]interface{}

	// pair is an atomic value so that future updates can implement key rotation
	pair atomic.Value

	now      func() time.Time
	duration time.Duration
	noncer   random.Noncer
}

func (f *factory) NewToken(r Request) (string, error) {
	merged := make(jwt.MapClaims, len(f.claims)+len(r.Claims))
	for k, v := range f.claims {
		merged[k] = v
	}

	for k, v := range r.Claims {
		merged[k] = v
	}

	var (
		now  = f.now().UTC()
		pair = f.pair.Load().(key.Pair)
	)

	merged["kid"] = pair.KID()
	merged["iat"] = now.Unix()

	if f.duration > 0 {
		merged["exp"] = now.Add(f.duration).UTC().Unix()
	}

	if f.noncer != nil {
		nonce, err := f.noncer.Nonce()
		if err != nil {
			return "", err
		}

		merged["jti"] = nonce
	}

	token := jwt.NewWithClaims(f.method, merged)
	return token.SignedString(pair.Sign())
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
}

func parseAlg(alg string) (jwt.SigningMethod, error) {
	method := jwt.GetSigningMethod(alg)
	if method == nil {
		return nil, fmt.Errorf("Unrecognized signing method: %s", alg)
	}

	return method, nil
}

func parseDuration(v string) (time.Duration, error) {
	if len(v) > 0 {
		return time.ParseDuration(v)
	}

	return 0, nil
}

// NewFactory creates a token Factory from a Descriptor.  The supplied Noncer is used if and only
// if d.Nonce is true.  Alternatively, supplying a nil Noncer will disable nonce creation altogether.
// The token's key pair is registered with the given key Registry.
func NewFactory(n random.Noncer, kr key.Registry, d Descriptor) (Factory, error) {
	method, err := parseAlg(d.Alg)
	if err == nil {
		return nil, err
	}

	duration, err := parseDuration(d.Duration)
	if err != nil {
		return nil, err
	}

	f := &factory{
		method:   method,
		claims:   make(map[string]interface{}, len(d.Claims)),
		now:      time.Now,
		duration: duration,
	}

	if d.Nonce {
		f.noncer = n
	}

	pair, err := kr.Register(d.Key)
	if err != nil {
		return nil, err
	}

	f.pair.Store(pair)
	for k, v := range d.Claims {
		f.claims[k] = v
	}

	return f, nil
}
