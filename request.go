// Package aci is a a Cisco ACI client library.
package aci

import (
	"net/http"
)

// Req is an API request wrapper around http.Request.
type Req struct {
	httpReq *http.Request
	refresh bool
}

// NoRefresh prevents token refresh check.
func NoRefresh(req *Req) {
	req.refresh = false
}

// Query sets query parameters.
func Query(k, v string) func(req *Req) {
	return func(req *Req) {
		q := req.httpReq.URL.Query()
		q.Add(k, v)
		req.httpReq.URL.RawQuery = q.Encode()
	}
}
