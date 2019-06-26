package issuer

import (
	"token"
)

type Options struct {
	Claims map[string]interface{}
}

func New(tf token.Factory, o Options) Issuer {
	i := &issuer{
		factory: tf,
	}

	if len(o.Claims) > 0 {
		i.claims = make(map[string]interface{}, len(o.Claims))
		for k, v := range o.Claims {
			i.claims[k] = v
		}
	}

	return i
}

type Issuer interface {
	Issue() (string, error)
}

type issuer struct {
	claims  map[string]interface{}
	factory token.Factory
}

func (i *issuer) Issue() (string, error) {
	return i.factory.NewToken(token.Request{
		Claims: i.claims,
	})
}
