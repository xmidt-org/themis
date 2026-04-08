// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package token

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/xmidt-org/sallust"
	"github.com/xmidt-org/themis/random"
	"github.com/xmidt-org/themis/xhttp/xhttpclient"
	"github.com/xmidt-org/themis/xhttp/xhttpserver"
	"github.com/xmidt-org/themis/xzap"
	"go.uber.org/zap"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

const (
	// ClaimTrust is the name of the trust value within JWT claims issued
	// by themis. This claim will be written based upon TLS connection state.
	ClaimTrust = "trust"
)

var (
	ErrRemoteURLRequired                = errors.New("a URL for the remote claimer is required")
	ErrMissingKey                       = errors.New("a key is required for all claims and metadata values")
	ErrInvalidRemoteClaimsConfiguration = errors.New("invalid remote claims' configuration")
)

// ClaimBuilder is a strategy for building token claims, given a token Request
type ClaimBuilder interface {
	AddClaims(context.Context, *Request, map[string]interface{}) error
}

type ClaimBuilderFunc func(context.Context, *Request, map[string]interface{}) error

func (cbf ClaimBuilderFunc) AddClaims(ctx context.Context, tr *Request, target map[string]interface{}) error {
	return cbf(ctx, tr, target)
}

// ClaimBuilders implements a pipeline of ClaimBuilder instances, invoked in sequence.
type ClaimBuilders []ClaimBuilder

func (cbs ClaimBuilders) AddClaims(ctx context.Context, r *Request, target map[string]interface{}) error {
	for _, e := range cbs {
		if err := e.AddClaims(ctx, r, target); err != nil {
			return err
		}
	}

	return nil
}

// requestClaimBuilder is a ClaimBuilder that copies the Request.Claims
type requestClaimBuilder struct{}

func (rc requestClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) error {
	maps.Copy(target, r.Claims)

	return nil
}

// staticClaimBuilder is a ClaimBuilder that simply appends a constant set of claims
type staticClaimBuilder map[string]interface{}

func (sc staticClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) error {
	maps.Copy(target, sc)

	return nil
}

// timeClaimBuilder is a ClaimBuilder which handles time-based claims
type timeClaimBuilder struct {
	now              func() time.Time
	duration         time.Duration
	disableNotBefore bool
	notBeforeDelta   time.Duration
}

func (tc *timeClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) error {
	now := tc.now().UTC()
	target["iat"] = now.Unix()

	if tc.duration > 0 {
		target["exp"] = now.Add(tc.duration).Unix()
	}

	if !tc.disableNotBefore {
		target["nbf"] = now.Add(tc.notBeforeDelta).Unix()
	}

	return nil
}

// nonceClaimBuilder is a ClaimBuilder that appends a nonce (jti) claim
type nonceClaimBuilder struct {
	n random.Noncer
}

func (nc nonceClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) error {
	nonce, err := nc.n.Nonce()
	if err != nil {
		return err
	}

	target["jti"] = nonce
	return nil
}

// remoteClaimBuilder invokes a remote system to obtain claims.
type remoteClaimBuilder struct {
	endpoint    endpoint.Endpoint
	url         string
	extra       map[string]interface{}
	apiResults  *prometheus.CounterVec
	apiDuration prometheus.ObserverVec
}

func (rc *remoteClaimBuilder) AddClaims(ctx context.Context, r *Request, target map[string]interface{}) error {
	rCopy := NewRequest()
	maps.Copy(rCopy.Metadata, r.Metadata)
	maps.Copy(rCopy.Metadata, rc.extra)
	maps.Copy(rCopy.PathWildCards, r.PathWildCards)
	maps.Copy(rCopy.QueryParameters, r.QueryParameters)
	startTime := time.Now()
	result, err := rc.endpoint(sallust.With(ctx, r.Logger), rCopy)
	duration := time.Since(startTime).Seconds()
	respErr := RemoteClaimsResponseError{}
	if err == nil { // Handle success outcomes.
		r.Logger.Info("successful response from remote claims endpoint")
		rc.apiDuration.With(prometheus.Labels{CodeLabelKey: strconv.Itoa(http.StatusOK), OutcomeLabelKey: SuccessOutcome}).Observe(duration)
		rc.apiResults.With(prometheus.Labels{CodeLabelKey: strconv.Itoa(http.StatusOK), OutcomeLabelKey: SuccessOutcome, ReasonLabelKey: ""}).Add(1)
		maps.Copy(target, result.(map[string]any))
	} else if errors.As(err, &respErr) { // Handle response related errors.
		code := respErr.StatusCode
		apiDuration := rc.apiDuration.MustCurryWith(prometheus.Labels{CodeLabelKey: strconv.Itoa(code)})
		apiResults := rc.apiResults.MustCurryWith(prometheus.Labels{CodeLabelKey: strconv.Itoa(code), ReasonLabelKey: GetRemoteClaimsReasonFromError(err)})
		if errors.Is(respErr, ErrRemoteClaimsResponseDecodingFailure) { // Handle decoding related errors.
			// Failure outcome.
			// Results in themis responding with a 500.
			ls := prometheus.Labels{OutcomeLabelKey: FailOutcome}
			apiDuration.With(ls).Observe(duration)
			apiResults.With(ls).Add(1)

			return respErr
		}

		switch codeCategory := code - code%100; codeCategory {
		case 500: // Success outcome.
			// 5XX HTTP responses from the remote claims endpoint
			// results in a 200 themis response, but no added remote claims to themis' jwt.
			r.Logger.Error(err.Error(), zap.Error(err))
		case 400, 300, 100, 0: // Success outcome.
			// 4XX, 3XX, 1XX or XX HTTP responses from the remote claims endpoint
			// results in a 200 themis response, but no added remote claims to themis' jwt.
			r.Logger.Warn(err.Error(), zap.Error(err))
		case 200: // Failure outcome.
			// Successful 2XX responses from remote claims that triggered a non-ErrRemoteClaimsResponseDecodingFailure error
			// results in themis responding with a 500.
			fallthrough
		default: // Failure outcome.
			// Results in themis responding with a 500.
			ls := prometheus.Labels{OutcomeLabelKey: FailOutcome}
			apiDuration.With(ls).Observe(duration)
			apiResults.With(ls).Add(1)
			r.Logger.Error(err.Error(), zap.Error(err))

			return respErr
		}

		// Success outcomes.
		// Results in a 200 themis response, but no added remote claims to themis' jwt.
		ls := prometheus.Labels{OutcomeLabelKey: SuccessOutcome}
		apiDuration.With(ls).Observe(duration)
		apiResults.With(ls).Add(1)
	} else if errors.Is(err, ErrRemoteClaimsRequestEncodingFailure) { // Handle request encoding related errors.
		// Failure outcome.
		// Results in themis responding with a 500.
		rc.apiDuration.With(prometheus.Labels{CodeLabelKey: "", OutcomeLabelKey: FailOutcome}).Observe(duration)
		rc.apiResults.With(prometheus.Labels{CodeLabelKey: "", OutcomeLabelKey: FailOutcome, ReasonLabelKey: GetRemoteClaimsReasonFromError(err)}).Add(1)
		r.Logger.Error("remote claims request encoding failure", zap.Error(err))

		return ErrRemoteClaimsRequestEncodingFailure
	} else if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) { // Handle request context related errors.
		rc.apiDuration.With(prometheus.Labels{CodeLabelKey: "", OutcomeLabelKey: SuccessOutcome}).Observe(duration)
		rc.apiResults.With(prometheus.Labels{CodeLabelKey: "", OutcomeLabelKey: SuccessOutcome, ReasonLabelKey: GetRemoteClaimsReasonFromError(err)}).Add(1)
		if errors.Is(err, context.DeadlineExceeded) {
			msg := "remote claims timeout"
			err = fmt.Errorf("%s: %s", msg, err.Error())
			r.Logger.Error(msg, zap.Error(err))
		} else {
			msg := "remote claims request canceled: themis token request was canceled"
			err = fmt.Errorf("%s: %s", msg, err.Error())
			r.Logger.Info(msg, zap.Error(err))
		}
	} else { // Handle gokit/configuration related errors.
		// Failure outcome.
		// Results in themis responding with a 500.
		rc.apiDuration.With(prometheus.Labels{CodeLabelKey: "", OutcomeLabelKey: FailOutcome}).Observe(duration)
		rc.apiResults.With(prometheus.Labels{CodeLabelKey: "", OutcomeLabelKey: FailOutcome, ReasonLabelKey: GetRemoteClaimsReasonFromError(err)}).Add(1)
		internalErr := fmt.Errorf("internal error details: %w: %s", ErrInvalidRemoteClaimsConfiguration, err.Error())
		r.Logger.Error("unknown remote claims failure", zap.Error(internalErr))

		return ErrInvalidRemoteClaimsConfiguration
	}

	return nil
}

func newRemoteClaimBuilder(client xhttpclient.Interface, metadata map[string]interface{}, r *RemoteClaims, apiResults *prometheus.CounterVec, duration prometheus.ObserverVec) (*remoteClaimBuilder, error) {
	if len(r.URL) == 0 {
		return nil, ErrRemoteURLRequired
	}

	url, err := url.Parse(r.URL)
	if err != nil {
		return nil, err
	}

	method := r.Method
	if len(method) == 0 {
		method = http.MethodPost
	}

	if client == nil {
		client = new(http.Client)
	}

	c := kithttp.NewClient(
		method,
		url,
		EncodeRemoteClaimsRequest,
		DecodeRemoteClaimsResponse,
		kithttp.SetClient(client),
		kithttp.ClientBefore(
			kithttp.SetRequestHeader("Content-Type", "application/json"),
		),
	)

	return &remoteClaimBuilder{endpoint: c.Endpoint(), url: r.URL, extra: metadata, apiResults: apiResults.MustCurryWith(prometheus.Labels{EndpointLabelKey: r.URL, MethodLabelKey: method}), apiDuration: duration.MustCurryWith(prometheus.Labels{EndpointLabelKey: r.URL, MethodLabelKey: method})}, nil
}

// newClientCertificateClaimBuiler creates a claim builder that sets trust based
// on client certificates.  This functional always returns a non-nil claimbuilder.
// Regular HTTP always results in a NoCertificates trust level.
func newClientCertificateClaimBuiler(cc *ClientCertificates) (cb *clientCertificateClaimBuilder, err error) {
	cb = new(clientCertificateClaimBuilder)
	if cc == nil {
		cb.trust = Trust{}.enforceDefaults()
		return
	}

	cb.trust = cc.Trust.enforceDefaults()

	if len(cc.RootCAFile) > 0 {
		cb.roots, err = xhttpserver.ReadCertPool(cc.RootCAFile)
	}

	if err == nil && len(cc.IntermediatesFile) > 0 {
		cb.intermediates, err = xhttpserver.ReadCertPool(cc.IntermediatesFile)
	}

	return
}

type clientCertificateClaimBuilder struct {
	roots         *x509.CertPool
	intermediates *x509.CertPool
	trust         Trust
}

func (cb *clientCertificateClaimBuilder) AddClaims(_ context.Context, r *Request, target map[string]interface{}) (err error) {
	// simplest case: this request either (1) didn't come from a TLS connection,
	// or (2) the client sent no certificates
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		target[ClaimTrust] = cb.trust.NoCertificates
		return
	}

	now := time.Now()
	var trust int
	for i, pc := range r.TLS.PeerCertificates {
		if i < len(r.TLS.VerifiedChains) && len(r.TLS.VerifiedChains[i]) > 0 {
			// the TLS layer already verified this certificate, so we're done
			// we assume Trusted is the highest trust level
			target[ClaimTrust] = cb.trust.Trusted
			return
		}

		// special logic around expired certificates
		expired := now.Before(pc.NotBefore) || now.After(pc.NotAfter)
		vo := x509.VerifyOptions{
			// always set the current time so that we disambiguate expired
			// from untrusted.
			CurrentTime:   pc.NotAfter.Add(-time.Second),
			Roots:         cb.roots,
			Intermediates: cb.intermediates,
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		}

		_, verifyErr := pc.Verify(vo)
		if verifyErr != nil {
			r.Logger.Warn(
				"certificate verification failed",
				zap.Error(verifyErr),
				xzap.Certificate("cert", pc),
			)
		}

		switch {
		case expired && verifyErr != nil:
			if trust < cb.trust.ExpiredUntrusted {
				trust = cb.trust.ExpiredUntrusted
			}

		case !expired && verifyErr != nil:
			if trust < cb.trust.Untrusted {
				trust = cb.trust.Untrusted
			}

		case expired && verifyErr == nil:
			if trust < cb.trust.ExpiredTrusted {
				trust = cb.trust.ExpiredTrusted
			}

		case !expired && verifyErr == nil:
			// we assume Trusted is the highest trust level
			target[ClaimTrust] = cb.trust.Trusted
			return
		}
	}

	// take the highest, non-Trusted level
	target[ClaimTrust] = trust
	return
}

// NewClaimBuilders constructs a ClaimBuilders from configuration.  The returned instance is typically
// used in configuration a token Factory.  It can be used as a standalone service component with an endpoint.
//
// The returned builders do not include those claims derived from HTTP requests.  Claims derived from HTTP
// requests are handled by NewRequestBuilders and DecodeServerRequest.
func NewClaimBuilders(n random.Noncer, client xhttpclient.Interface, o Options, remoteResults *prometheus.CounterVec, remoteDuration prometheus.ObserverVec) (ClaimBuilders, error) {
	builders := ClaimBuilders{requestClaimBuilder{}}
	staticClaims, err := getStaticValues(o.Claims)
	if err != nil {
		return nil, fmt.Errorf("static claim builder configuration failure: %w", err)
	}

	builders = append(builders, staticClaimBuilder(staticClaims))
	if o.Nonce && n != nil {
		builders = append(builders, nonceClaimBuilder{n: n})
	}

	if !o.DisableTime {
		builders = append(
			builders,
			&timeClaimBuilder{
				now:              time.Now,
				duration:         o.Duration,
				disableNotBefore: o.DisableNotBefore,
				notBeforeDelta:   o.NotBeforeDelta,
			})
	}

	cb, err := newClientCertificateClaimBuiler(o.ClientCertificates)
	if err == nil {
		builders = append(
			builders,
			cb,
		)
	}

	if o.Remote != nil {
		metadata, err := getStaticValues(o.Metadata)
		if err != nil {
			return nil, fmt.Errorf("remote claim builder configuration failure: metadata error: %w", err)
		}

		remoteClaimBuilder, err := newRemoteClaimBuilder(client, metadata, o.Remote, remoteResults, remoteDuration)
		if err != nil {
			return nil, err
		}

		builders = append(builders, remoteClaimBuilder)
	}

	return builders, err
}

func getStaticValues(vals []Value) (map[string]any, error) {
	var errs []error

	m := make(map[string]any)
	for _, v := range vals {
		errs = append(errs, v.Validate())
		if !v.IsStatic() {
			continue
		}

		msg, err := v.RawMessage()
		errs = append(errs, err)
		m[v.Key] = msg
	}

	return m, errors.Join(errs...)
}
