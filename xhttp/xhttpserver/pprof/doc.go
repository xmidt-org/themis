// Package pprof exposes simple integrations between net/http/pprof, gorilla/mux, and uber/fx.
// This has to be in its own package to avoid side effects from importing net/http/pprof in the
// cases where a client does not want pprof.
package pprof
