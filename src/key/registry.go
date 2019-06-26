package key

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"sync"
)

const (
	KeyTypeRSA    = "rsa"
	KeyTypeECDSA  = "ecdsa"
	KeyTypeSecret = "secret"
)

var (
	ErrInvalidDescriptor = errors.New("Either generate or file must be set")
)

// Descriptor holds the configurable options for a key Pair
type Descriptor struct {
	// Kid is the key id to use initially.  If unset, the name of the key is used.  Note that the kid can
	// change is the key is rotated or updated during application execution.
	Kid string

	// Type indicates the type of key.  This field dictates both how the key File is read or how the key
	// is generated.  The default is "rsa".
	Type string

	// Bits indicates the bit size for a generated key
	Bits int

	// File is the system path to a file where the key is stored.  If set, this file must exist and contain
	// either a secret or a PEM-encoded key pair.  If this field is not set, a key is generated.
	File string
}

// Registry holds zero or more key Pairs
type Registry interface {
	// Get returns the Pair associated with a given key identifier
	Get(kid string) (Pair, bool)

	// Register creates a new Pair from a Descriptor and stores it in this registry
	Register(Descriptor) (Pair, error)

	// Update rotate a key Pair, deleting the old key identifier
	Update(old string, new Pair) bool
}

// NewRegistry creates a new key Registry backed by a given source of randomness for generation.
// If random is nil, crypto/rand.Reader is used.
func NewRegistry(random io.Reader) Registry {
	if random == nil {
		random = rand.Reader
	}

	return &registry{
		pairs:  make(map[string]Pair),
		random: random,
	}
}

type registry struct {
	lock   sync.RWMutex
	pairs  map[string]Pair
	random io.Reader
}

func (r *registry) Get(kid string) (Pair, bool) {
	r.lock.RLock()
	p, ok := r.pairs[kid]
	r.lock.RUnlock()
	return p, ok
}

func (r *registry) newPair(d Descriptor) (Pair, error) {
	if len(d.File) > 0 {
		return ReadPair(d.Kid, d.File)
	}

	switch d.Type {
	case "":
		fallthrough
	case KeyTypeRSA:
		return GenerateRSAPair(d.Kid, r.random, d.Bits)
	case KeyTypeECDSA:
		return GenerateECDSAPair(d.Kid, r.random, d.Bits)
	case KeyTypeSecret:
		return GenerateSecretPair(d.Kid, r.random, d.Bits)
	default:
		return nil, fmt.Errorf("Invalid key type: %s", d.Type)
	}

	return nil, ErrInvalidDescriptor
}

func (r *registry) Register(d Descriptor) (Pair, error) {
	p, err := r.newPair(d)
	if err != nil {
		return nil, err
	}

	defer r.lock.Unlock()
	r.lock.Lock()

	if _, ok := r.pairs[p.KID()]; ok {
		return nil, fmt.Errorf("Key id already used: %s", p.KID())
	}

	r.pairs[p.KID()] = p
	return p, nil
}

func (r *registry) Update(old string, new Pair) bool {
	defer r.lock.Unlock()
	r.lock.Lock()

	if _, ok := r.pairs[old]; !ok {
		return false
	}

	delete(r.pairs, old)
	r.pairs[new.KID()] = new
	return true
}
