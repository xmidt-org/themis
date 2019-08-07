package xhttp

import "net/http"

// CanonicalizeHeaders returns a copy of the source with each key canonicalized via http.CanonicalHeaderKey.
// This is useful when reading headers from sources that are not guaranteed to have canonical header names,
// such as configuration files.  Each key's values are shallow copied.
func CanonicalizeHeaders(source http.Header) http.Header {
	target := make(http.Header, len(source))
	for key, values := range source {
		for _, value := range values {
			target.Add(key, value)
		}
	}

	return target
}

// CanonicalizeHeaderMap produces an http.Header from a map of strings.  Each key is canonicalized via
// http.CanonicalHeaderKey.  This function is primarily useful when reading in configuration or when
// processing non-header maps.
func CanonicalizeHeaderMap(source map[string]string) http.Header {
	target := make(http.Header, len(source))
	for key, value := range source {
		target.Set(key, value)
	}

	return target
}

// AddHeaders adds each source header onto the destination.  Existing headers in the target are preserved.
// Headers already present in the target have the source's values appended.
//
// This function assumes that the source is already canonicalized.
func AddHeaders(target, source http.Header) {
	for key, values := range source {
		target[key] = append(target[key], values...)
	}
}

// SetHeaders works like AddHeader, except that any headers already present in the target are overwritten.
//
// This function assumes that the source is already canonicalized.
func SetHeaders(target, source http.Header) {
	for key, values := range source {
		target[key] = values
	}
}
