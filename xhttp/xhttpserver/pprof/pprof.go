// SPDX-FileCopyrightText: 2017 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0
package pprof

import (
	httppprof "net/http/pprof"
	"runtime/pprof"

	"github.com/gorilla/mux"
)

// BuildRoutes adds the handlers from net/http/pprof, using the standard paths,
// to the given Router.
func BuildRoutes(r *mux.Router) {
	r.HandleFunc("/debug/pprof/", httppprof.Index)
	for _, p := range pprof.Profiles() {
		r.HandleFunc("/debug/pprof/"+p.Name(), httppprof.Index)
	}

	r.HandleFunc("/debug/pprof/cmdline", httppprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", httppprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", httppprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", httppprof.Trace)
}
