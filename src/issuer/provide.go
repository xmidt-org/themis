package issuer

import (
	"net/http"
	"token"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type In struct {
	fx.In

	Factory token.Factory
	Viper   *viper.Viper
}

type Out struct {
	fx.Out

	Issuer  Issuer
	Handler http.Handler `name:"issueHandler"`
}

func Provide(key string) func(In) (Out, error) {
	return func(in In) (Out, error) {
		var o Options
		if err := in.Viper.UnmarshalKey(key, &o); err != nil {
			return Out{}, err
		}

		i := New(in.Factory, o)

		return Out{
			Issuer: i,
			Handler: NewHandler(
				NewEndpoint(i),
			),
		}, nil
	}
}
