package token

import (
	"context"
	"fmt"
	"key"
	"sync/atomic"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	DefaultAlg = "RS256"
)

// Request is a token creation request.  Clients can pass in arbitrary claims, typically things like "iss",
// to merge and override anything set on the factory via configuration.
type Request struct {
	// Claims holds the extra claims to add to tokens.  These claims will override any configured claims in a Factory,
	// but will not override time-based claims such as nbf or exp.
	Claims map[string]interface{}

	// Metadata holds non-claim information about the request, usually garnered from the original HTTP request.  This
	// metadata is available to lower levels of infrastructure used by the Factory.
	Metadata map[string]interface{}
}

// Factory is a creation strategy for signed JWT tokens
type Factory interface {
	// NewToken uses a Request to produce a signed JWT token
	NewToken(context.Context, *Request) (string, error)
}

type factory struct {
	method  jwt.SigningMethod
	claimer Claimer

	// pair is an atomic value so that future updates can implement key rotation
	pair atomic.Value
}

func (f *factory) NewToken(ctx context.Context, r *Request) (string, error) {
	merged := make(map[string]interface{}, len(r.Claims))
	if err := f.claimer.Append(ctx, r, merged); err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(f.method, jwt.MapClaims(merged))
	pair := f.pair.Load().(key.Pair)
	token.Header["kid"] = pair.KID()
	return token.SignedString(pair.Sign())
}

// NewFactory creates a token Factory from a Descriptor.  The supplied Noncer is used if and only
// if d.Nonce is true.  Alternatively, supplying a nil Noncer will disable nonce creation altogether.
// The token's key pair is registered with the given key Registry.
func NewFactory(c Claimer, kr key.Registry, o Options) (Factory, error) {
	if len(o.Alg) == 0 {
		o.Alg = DefaultAlg
	}

	f := &factory{
		method:  jwt.GetSigningMethod(o.Alg),
		claimer: c,
	}

	if f.method == nil {
		return nil, fmt.Errorf("No such signing method: %s", o.Alg)
	}

	pair, err := kr.Register(o.Key)
	if err != nil {
		return nil, err
	}

	f.pair.Store(pair)
	return f, nil
}
