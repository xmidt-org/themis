// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package xmetricshttp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler http.Handler

func NewHandler(g prometheus.Gatherer, o promhttp.HandlerOpts) Handler {
	return promhttp.HandlerFor(g, o)
}
