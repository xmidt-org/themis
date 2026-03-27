// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync/atomic"

	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/themis/key"
	"github.com/xmidt-org/themis/token/trust"
	"go.uber.org/zap"

	"github.com/golang-jwt/jwt"
)

const (
	DefaultAlg = "RS256"
)

// Request is a token creation request.  Clients can pass in arbitrary claims, typically things like "iss",
// to merge and override anything set on the factory via configuration.
type Request struct {
	// Logger is the zap logger which can be used to log information about the request.
	// This field is always set.
	Logger *zap.Logger

	// Claims holds the extra claims to add to tokens.  These claims will override any configured claims in a Factory,
	// but will not override time-based claims such as nbf or exp.
	Claims map[string]interface{}

	// TLS represents the state of any underlying TLS connection.
	// For non-tls connections, this field is unset.
	TLS *tls.ConnectionState

	// The following fields are for remote claims' requests.
	Metadata        map[string]any // Metadata is the request payload.
	PathWildCards   map[string]any // PathWildCards are the request path wildcards.
	QueryParameters map[string]any // QueryParameters are the request query parameters.
}

func (r *Request) SetTrustMetdata(tType trust.Type, tValue int) {
	r.Metadata[ClaimTrustType] = tType
	r.Metadata[ClaimTrust] = tValue
}

// NewRequest returns an empty, fully initialized token Request
func NewRequest() *Request {
	return &Request{
		Logger:          sallust.Default(),
		Claims:          make(map[string]interface{}),
		Metadata:        make(map[string]interface{}),
		PathWildCards:   make(map[string]interface{}),
		QueryParameters: make(map[string]interface{}),
	}
}

// Factory is a creation strategy for signed JWT tokens
type Factory interface {
	// NewToken uses a Request to produce a signed JWT token
	NewToken(context.Context, *Request) (string, error)
}

type factory struct {
	method       jwt.SigningMethod
	claimBuilder ClaimBuilder

	// pair is an atomic value so that future updates can implement key rotation
	pair atomic.Value
}

func (f *factory) NewToken(ctx context.Context, r *Request) (string, error) {
	merged := make(map[string]interface{}, len(r.Claims))
	if err := f.claimBuilder.AddClaims(ctx, r, merged); err != nil {
		return "", err
	}

	r.Logger.Info("new token", zap.Any("trust", merged[ClaimTrust]))
	token := jwt.NewWithClaims(f.method, jwt.MapClaims(merged))
	pair := f.pair.Load().(key.Pair)
	token.Header["kid"] = pair.KID()
	return token.SignedString(pair.Sign())
}

// NewFactory creates a token Factory from a Descriptor.  The supplied Noncer is used if and only
// if d.Nonce is true.  Alternatively, supplying a nil Noncer will disable nonce creation altogether.
// The token's key pair is registered with the given key Registry.
func NewFactory(o Options, cb ClaimBuilder, kr key.Registry) (Factory, error) {
	if len(o.Alg) == 0 {
		o.Alg = DefaultAlg
	}

	f := &factory{
		method:       jwt.GetSigningMethod(o.Alg),
		claimBuilder: cb,
	}

	if f.method == nil {
		return nil, fmt.Errorf("no such signing method: %s", o.Alg)
	}

	pair, err := kr.Register(o.Key)
	if err != nil {
		return nil, err
	}

	f.pair.Store(pair)
	return f, nil
}
