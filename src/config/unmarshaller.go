package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// KeyUnmarshaller is a strategy for unmarshalling objects from hierarchial configuration sources.
type KeyUnmarshaller interface {
	// IsSet checks if a configuration key is present, either in the source or as a default.
	// See https://godoc.org/github.com/spf13/viper#IsSet
	IsSet(string) bool

	// UnmarshalKey unmarshals an object, normally a pointer to struct, from a given configuration key.
	// No attempt is made to verify the existence of the configuration key.  If no such key exists,
	// unmarshalling will still happen against a default set of options.
	// See https://godoc.org/github.com/spf13/viper#UnmarshalKey
	UnmarshalKey(string, interface{}) error
}

// Unmarshaller is a strategy for unmarshalling configuration, mostly in the form of structs.
type Unmarshaller interface {
	KeyUnmarshaller
	Unmarshal(interface{}) error
}

// ViperUnmarshaller is an Unmarshaller backed by Viper.
type ViperUnmarshaller struct {
	// Viper is the required viper instance
	Viper *viper.Viper

	// Options are passed to viper for each unmarshal operation.  This field is optional.
	Options []viper.DecoderConfigOption
}

func (vu ViperUnmarshaller) IsSet(k string) bool {
	return vu.Viper.IsSet(k)
}

func (vu ViperUnmarshaller) Unmarshal(v interface{}) error {
	return vu.Viper.Unmarshal(v, vu.Options...)
}

func (vu ViperUnmarshaller) UnmarshalKey(k string, v interface{}) error {
	return vu.Viper.UnmarshalKey(k, v, vu.Options...)
}

// MissingKeyError is returned when a required key was not found in the configuration
type MissingKeyError interface {
	error
	Key() string
}

type missingKeyError struct {
	k string
}

func (mke missingKeyError) Key() string {
	return mke.k
}

func (mke missingKeyError) Error() string {
	return fmt.Sprintf("Missing configuration key: %s", mke.k)
}

// NewMissingKeyError creates an error that indicates the given configuration key is not
// present in the underlying configuration source
func NewMissingKeyError(k string) MissingKeyError {
	return missingKeyError{k: k}
}

// UnmarshalRequired is a helper function that returns an error if the given key is not present
func UnmarshalRequired(ku KeyUnmarshaller, k string, v interface{}) error {
	if !ku.IsSet(k) {
		return NewMissingKeyError(k)
	}

	return ku.UnmarshalKey(k, v)
}
