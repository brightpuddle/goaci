// Package goaci is a a Cisco ACI client library for Go.
package goaci

import (
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/sjson"
)

// Body wraps SJSON for building JSON body strings.
type Body struct {
	Str string
}

// Set sets a JSON path to a value.
func (body Body) Set(path, value string) Body {
	res, _ := sjson.Set(body.Str, path, value)
	body.Str = res
	return body
}

// SetRaw sets a JSON path to a raw string value.
func (body Body) SetRaw(path, rawValue string) Body {
	res, _ := sjson.SetRaw(body.Str, path, rawValue)
	body.Str = res
	return body
}

// Req wraps http.Request for API requests.
type Req struct {
	httpReq *http.Request
	refresh bool
}

// NewReq creates a new Req.
func NewReq(method, uri string, body io.Reader, mods ...func(*Req)) Req {
	httpReq, _ := http.NewRequest(method, fmt.Sprintf("%s.json", uri), body)
	req := Req{
		httpReq: httpReq,
		refresh: true,
	}
	for _, mod := range mods {
		mod(&req)
	}
	return req
}

// NoRefresh prevents token refresh check.
func NoRefresh(req *Req) {
	req.refresh = false
}

// Query sets an HTTP query parameter.
func Query(k, v string) func(req *Req) {
	return func(req *Req) {
		q := req.httpReq.URL.Query()
		q.Add(k, v)
		req.httpReq.URL.RawQuery = q.Encode()
	}
}
