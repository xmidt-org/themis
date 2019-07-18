package xmetricshttp

import (
	"xmetrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHttpOut struct {
	xmetrics.MetricsOut

	Handler Handler
}

func Unmarshal(configKey string, ho promhttp.HandlerOpts) func(xmetrics.MetricsIn) (MetricsHttpOut, error) {
	return func(in xmetrics.MetricsIn) (MetricsHttpOut, error) {
		out, err := xmetrics.Unmarshal(configKey)(in)
		if err != nil {
			return MetricsHttpOut{}, err
		}

		return MetricsHttpOut{
			MetricsOut: out,
			Handler:    NewHandler(out.Gatherer, ho),
		}, nil
	}
}
