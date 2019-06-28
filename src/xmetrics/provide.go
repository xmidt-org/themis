package xmetrics

import (
	"net/http"
	"xhttp/xhttpserver"

	"github.com/go-kit/kit/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

type ProvideOut struct {
	fx.Out

	Registerer prometheus.Registerer
	Gatherer   prometheus.Gatherer
	Registry   Registry
	Handler    http.Handler `name:"metricsHandler"`
	Router     *mux.Router  `name:"metricsRouter"`
}

func Provide(serverConfigKey, metricsConfigKey string, ho promhttp.HandlerOpts) func(xhttpserver.ProvideIn) (ProvideOut, error) {
	return func(serverIn xhttpserver.ProvideIn) (ProvideOut, error) {
		router, err := xhttpserver.Unmarshal(serverConfigKey, serverIn)
		if err != nil {
			return ProvideOut{}, err
		}

		var o Options
		if err := serverIn.Viper.UnmarshalKey(metricsConfigKey, &o); err != nil {
			return ProvideOut{}, err
		}

		registry, err := New(o)
		if err != nil {
			return ProvideOut{}, err
		}

		return ProvideOut{
			Registerer: registry,
			Gatherer:   registry,
			Registry:   registry,
			Handler:    promhttp.HandlerFor(registry, ho),
			Router:     router,
		}, nil
	}
}

type InvokeIn struct {
	fx.In

	Handler http.Handler `name:"metricsHandler"`
	Router  *mux.Router  `name:"metricsRouter"`
}

func RunServer(path string) func(InvokeIn) {
	return func(in InvokeIn) {
		in.Router.Handle(path, in.Handler)
	}
}

func ProvideCounter(o prometheus.CounterOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Counter, error) {
			return r.NewCounter(o, labelNames)
		},
	}
}

func ProvideGauge(o prometheus.GaugeOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Gauge, error) {
			return r.NewGauge(o, labelNames)
		},
	}
}

func ProvideHistogram(o prometheus.HistogramOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Histogram, error) {
			return r.NewHistogram(o, labelNames)
		},
	}
}

func ProvideSummary(o prometheus.SummaryOpts, labelNames ...string) fx.Annotated {
	return fx.Annotated{
		Name: o.Name,
		Target: func(r Registry) (metrics.Histogram, error) {
			return r.NewSummary(o, labelNames)
		},
	}
}
