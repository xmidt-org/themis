package service

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

// IDGenerator is a strategy interface for generating identifiers appropriate for services
// and service discovery
type IDGenerator interface {
	NewID() (string, error)
}

// Base64IDGenerator is a UUID-like implementation of IDGenerator.  It generates base64-encoded
// random bytes.
type Base64IDGenerator struct {
	// Encoding is the base64 Encoding to use for the final id.  If not set, base64.RawURLEncoding is used.
	Encoding *base64.Encoding

	// Random is the source of randomness.  If not set, math/rand.Read is used.
	Random io.Reader

	// Length is the number of random bytes to read prior to encoding.  If nonpositive, 16 bytes are read, which is
	// the length of UUIDs.
	Length int
}

func (bg Base64IDGenerator) NewID() (string, error) {
	l := bg.Length
	if l < 1 {
		l = 16
	}

	b := make([]byte, l)
	if bg.Random != nil {
		_, err := io.ReadFull(bg.Random, b)
		if err != nil {
			return "", err
		}
	} else {
		rand.Read(b)
	}

	e := bg.Encoding
	if e == nil {
		e = base64.RawURLEncoding
	}

	return e.EncodeToString(b), nil
}
