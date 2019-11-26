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
func (c Client) NewRequest(method, urn string, body io.Reader) Request {
	uri := fmt.Sprintf("%s%s.json", c.url, urn)
	httpReq, _ := http.NewRequest(method, uri, body)
	return Request{
		httpReq: httpReq,
		refresh: true,
	}
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
