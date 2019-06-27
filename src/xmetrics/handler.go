package xmetrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler http.Handler

func NewHandler(g prometheus.Gatherer, o promhttp.HandlerOpts) Handler {
	return promhttp.HandlerFor(g, o)
}
