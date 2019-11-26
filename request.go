// Package aci is a a Cisco ACI client library.
package aci

import (
	"fmt"
	"io"
	"net/http"
)

// Request is an API request wrapper around http.Request.
type Request struct {
	httpReq *http.Request
	refresh bool
}

// NewRequest creates a new Request against this client.
func NewRequest(method, uri string, body io.Reader, mods ...func(*Request)) Request {
	httpReq, _ := http.NewRequest(method, fmt.Sprintf("%s.json", uri), body)
	req := Request{
		httpReq: httpReq,
		refresh: true,
	}
	for _, mod := range mods {
		mod(&req)
	}
	return req
}

// NoRefresh prevents token refresh check.
func NoRefresh(req *Request) {
	req.refresh = false
}

// Query sets query parameters.
func Query(k, v string) func(req *Request) {
	return func(req *Request) {
		q := req.httpReq.URL.Query()
		q.Add(k, v)
		req.httpReq.URL.RawQuery = q.Encode()
	}
}
