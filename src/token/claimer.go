package token

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"random"
	"time"
)

var (
	ErrRemoteURLRequired = errors.New("A URL for the remote claimer is required")
)

// Claimer represents a strategy for obtaining claims, typically through configuration
// or from some remote system.
type Claimer interface {
	Append(context.Context, *Request, map[string]interface{}) error
}

// localClaimer is a Claimer strategy that constructs some claims based on internal configuration
type localClaimer struct {
	claims map[string]interface{}

	now            func() time.Time
	noncer         random.Noncer
	duration       time.Duration
	notBeforeDelta *time.Duration
}

func (lc *localClaimer) Append(_ context.Context, r *Request, c map[string]interface{}) error {
	for k, v := range lc.claims {
		c[k] = v
	}

	for k, v := range r.Claims {
		c[k] = v
	}

	now := lc.now().UTC()
	c["iat"] = now.Unix()

	if lc.duration > 0 {
		c["exp"] = now.Add(lc.duration).Unix()
	}

	if lc.notBeforeDelta != nil {
		c["nbf"] = now.Add(*lc.notBeforeDelta).Unix()
	}

	if lc.noncer != nil {
		nonce, err := lc.noncer.Nonce()
		if err != nil {
			return err
		}

		c["jti"] = nonce
	}

	return nil
}

func newLocalClaimer(n random.Noncer, d Descriptor) (*localClaimer, error) {
	lc := &localClaimer{
		claims: make(map[string]interface{}, len(d.Claims)),
		now:    time.Now,
	}

	for k, v := range d.Claims {
		lc.claims[k] = v
	}

	var err error
	if len(d.Duration) > 0 {
		lc.duration, err = time.ParseDuration(d.Duration)
		if err != nil {
			return nil, err
		}
	}

	if len(d.NotBeforeDelta) > 0 {
		lc.notBeforeDelta = new(time.Duration)
		*lc.notBeforeDelta, err = time.ParseDuration(d.NotBeforeDelta)
		if err != nil {
			return nil, err
		}
	}

	if d.Nonce {
		lc.noncer = n
	}

	return lc, nil
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type remoteClaimer struct {
	method string
	url    string
	client httpClient
}

func (rc *remoteClaimer) Append(ctx context.Context, r *Request, c map[string]interface{}) error {
	method := rc.method
	if len(method) == 0 {
		method = http.MethodGet
	}

	client := rc.client
	if client == nil {
		client = http.DefaultClient
	}

	var body *bytes.Buffer
	if len(r.Meta) > 0 {
		meta, err := json.Marshal(r.Meta)
		if err != nil {
			return err
		}

		body = bytes.NewBuffer(meta)
	}

	hr, err := http.NewRequest(method, rc.url, body)
	if err != nil {
		return err
	}

	response, err := client.Do(hr.WithContext(ctx))
	if err != nil {
		return err
	}

	defer func() {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
	}()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("Remote claims system returned status %d", response.StatusCode)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var claims map[string]interface{}
	err = json.Unmarshal(responseBody, &claims)
	if err != nil {
		return err
	}

	for k, v := range claims {
		c[k] = v
	}

	return nil
}

func newRemoteClaimer(r *RemoteClaims) (*remoteClaimer, error) {
	if r == nil {
		return nil, nil
	}

	if len(r.URL) == 0 {
		return nil, ErrRemoteURLRequired
	}

	if _, err := url.Parse(r.URL); err != nil {
		return nil, err
	}

	rc := &remoteClaimer{
		method: r.Method,
		url:    r.URL,
		client: new(http.Client),
	}

	if len(rc.method) == 0 {
		rc.method = http.MethodPost
	}

	return rc, nil
}

type Claimers []Claimer

func (cs Claimers) Append(ctx context.Context, r *Request, c map[string]interface{}) error {
	for _, e := range cs {
		if err := e.Append(ctx, r, c); err != nil {
			return err
		}
	}

	return nil
}
