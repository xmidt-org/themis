package pprof

import (
	"net/http/pprof"

	"github.com/gorilla/mux"
)

// BuildRoutes adds the handlers from net/http/pprof, using the standard paths,
// to the given Router.
func BuildRoutes(r *mux.Router) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
