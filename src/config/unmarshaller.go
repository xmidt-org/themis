package config

import "github.com/spf13/viper"

// Unmarshaller is a strategy for unmarshalling configuration, mostly in the form of structs
type Unmarshaller interface {
	Unmarshal(interface{}) error
}

// KeyUnmarshaller is a strategy for unmarshalling configuration keys, mostly in the form of structs
type KeyUnmarshaller interface {
	UnmarshalKey(string, interface{}) error
}

// ViperUnmarshaller is both an Unmarshaller and a KeyUnmarshaller backed by Viper.  This type
// allows a common place for decoder options.
type ViperUnmarshaller struct {
	Viper   *viper.Viper
	Options []viper.DecoderConfigOption
}

func (vu ViperUnmarshaller) Unmarshal(v interface{}) error {
	return vu.Viper.Unmarshal(v, vu.Options...)
}

func (vu ViperUnmarshaller) UnmarshalKey(k string, v interface{}) error {
	return vu.Viper.UnmarshalKey(k, v, vu.Options...)
}
