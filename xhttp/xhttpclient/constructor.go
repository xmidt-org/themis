package xhttpclient

import (
	"net/http"
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (rtf RoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return rtf(request)
}

// Constructor is an Alice-style constructor for RoundTrippers
type Constructor func(http.RoundTripper) http.RoundTripper

// Chain is a sequence of Constructors, just like github.com/justinas/alice.Chain
type Chain struct {
	c []Constructor
}

func NewChain(c ...Constructor) Chain {
	return Chain{c: c}
}

func (ch Chain) Append(more ...Constructor) Chain {
	merged := Chain{
		c: make([]Constructor, 0, len(ch.c)+len(more)),
	}

	merged.c = append(merged.c, ch.c...)
	merged.c = append(merged.c, more...)
	return merged
}

func (ch Chain) Extend(more Chain) Chain {
	return ch.Append(more.c...)
}

func (ch Chain) Then(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = new(http.Transport)
	}

	for i := range ch.c {
		rt = ch.c[len(ch.c)-1-i](rt)
	}

	return rt
}

func (ch Chain) ThenFunc(rtf RoundTripperFunc) http.RoundTripper {
	if rtf == nil {
		return ch.Then(nil)
	}

	return ch.Then(rtf)
}
