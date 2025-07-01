// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"

	"os"

	"github.com/lestrrat-go/jwx/jwk"
)

const (
	DefaultRSABits    = 1024
	DefaultSecretBits = 512
)

var (
	ErrUnrecognizedKeyData = errors.New("unable to read key data")

	DefaultCurve = elliptic.P384()
)

type Pair interface {
	// KID is the key identifier for this Pair
	KID() string

	// Sign returns the signing key for generating signed JWT tokens.
	Sign() interface{}

	// WriteVerifyPEMto writes the PEM-encoded verify key to an arbitrary output sink.
	WriteVerifyPEMTo(io.Writer) (int64, error)

	WriteJWK(io.Writer) (int64, error)
}

type pair struct {
	kid        string
	sign       interface{}
	verifyPEM  []byte
	jsonWebKey []byte
}

func (p pair) KID() string {
	return p.kid
}

func (p pair) Sign() interface{} {
	return p.sign
}

func (p pair) WriteVerifyPEMTo(w io.Writer) (int64, error) {
	c, err := w.Write(p.verifyPEM)
	return int64(c), err
}

func (p pair) WriteJWK(w io.Writer) (int64, error) {
	c, err := w.Write(p.jsonWebKey)
	return int64(c), err
}

func NewPair(kid string, key interface{}) (Pair, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		verifyPEM, err := MarshalPKIXPublicKeyToPEM(&k.PublicKey)
		if err != nil {
			return nil, err
		}

		jwkKey, err := jwk.New(&k.PublicKey)
		if err != nil {
			return nil, err
		}
		jsonWebKey, err := json.MarshalIndent(jwkKey, "", "  ")
		if err != nil {
			return nil, err
		}

		return pair{
			kid:        kid,
			sign:       key,
			verifyPEM:  verifyPEM,
			jsonWebKey: jsonWebKey,
		}, nil

	case *ecdsa.PrivateKey:
		verifyPEM, err := MarshalPKIXPublicKeyToPEM(&k.PublicKey)
		if err != nil {
			return nil, err
		}

		jwkKey, err := jwk.New(&k.PublicKey)
		if err != nil {
			return nil, err
		}
		jsonWebKey, err := json.MarshalIndent(jwkKey, "", "  ")
		if err != nil {
			return nil, err
		}

		return pair{
			kid:        kid,
			sign:       key,
			verifyPEM:  verifyPEM,
			jsonWebKey: jsonWebKey,
		}, nil

	case []byte:
		jwkKey, err := jwk.New(k)
		if err != nil {
			return nil, err
		}
		jsonWebKey, err := json.MarshalIndent(jwkKey, "", "  ")
		if err != nil {
			return nil, err
		}

		return pair{
			kid:  kid,
			sign: key,
			verifyPEM: pem.EncodeToMemory(
				&pem.Block{
					Type:  "PUBLIC KEY",
					Bytes: k,
				},
			),
			jsonWebKey: jsonWebKey,
		}, nil

	case string:
		keyBytes := []byte(k)

		jwkKey, err := jwk.New(keyBytes)
		if err != nil {
			return nil, err
		}
		jsonWebKey, err := json.MarshalIndent(jwkKey, "", "  ")
		if err != nil {
			return nil, err
		}

		return pair{
			kid:  kid,
			sign: keyBytes,
			verifyPEM: pem.EncodeToMemory(
				&pem.Block{
					Type:  "PUBLIC KEY",
					Bytes: keyBytes,
				},
			),
			jsonWebKey: jsonWebKey,
		}, nil
	}
	return nil, fmt.Errorf("unsupported key type: %v", key)
}

func ReadPair(kid string, file string) (Pair, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return ReadPairBytes(kid, data)
}

func ReadPairBytes(kid string, data []byte) (Pair, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return NewPair(kid, data)
	}

	if rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return NewPair(kid, rsaKey)
	}

	if pkcs8, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return NewPair(kid, pkcs8)
	}

	return nil, ErrUnrecognizedKeyData
}

func GenerateRSAPair(kid string, random io.Reader, bits int) (Pair, error) {
	if bits <= 0 {
		bits = DefaultRSABits
	}

	key, err := rsa.GenerateKey(random, bits)
	if err != nil {
		return nil, err
	}

	return NewPair(kid, key)
}

func GenerateECDSAPair(kid string, random io.Reader, bits int) (Pair, error) {
	curve := DefaultCurve

	if bits > 0 {
		switch bits {
		case 224:
			curve = elliptic.P224()
		case 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()

		// oddity: the P521() method returns the curve for 512 bit JWT signing
		case 512:
			curve = elliptic.P521()

		default:
			return nil, fmt.Errorf("unsupported curve value: %d", curve)
		}
	}

	key, err := ecdsa.GenerateKey(curve, random)
	if err != nil {
		return nil, err
	}

	return NewPair(kid, key)
}

func GenerateSecretPair(kid string, random io.Reader, bits int) (Pair, error) {
	if bits <= 0 {
		bits = DefaultSecretBits
	}

	secret := make([]byte, bits)
	if _, err := random.Read(secret); err != nil {
		return nil, err
	}

	return NewPair(kid, secret)
}

// MarshalPKIXPublicKeyToPEM handles marshaling a public key in PKIX format which is
// then encoded as a PEM block
func MarshalPKIXPublicKeyToPEM(key interface{}) ([]byte, error) {
	pkix, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pkix,
		},
	), nil
}
