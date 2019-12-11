package goaci

import (
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Body wraps SJSON for building JSON body strings.
// Usage example:
//   Body{}.Set("fvTenant.attributes.name", "mytenant").Str
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
// This is primarily used for building up nested structures, e.g.:
//   Body{}.SetRaw("fvTenant.attributes", Body{}.Set("name", "mytenant").Str).Str
func (body Body) SetRaw(path, rawValue string) Body {
	res, _ := sjson.SetRaw(body.Str, path, rawValue)
	body.Str = res
	return body
}

// Res creates a Res object, i.e. a GJSON result object.
func (body Body) Res() Res {
	return gjson.Parse(body.Str)
}

// Req wraps http.Request for API requests.
type Req struct {
	// HttpReq is the *http.Request obejct.
	HttpReq *http.Request
	// Refresh indicates whether token refresh should be checked for this request.
	// Pass NoRefresh to disable Refresh check.
	Refresh bool
}

// NewReq creates a new Req request.
func NewReq(method, uri string, body io.Reader, mods ...func(*Req)) Req {
	httpReq, _ := http.NewRequest(method, fmt.Sprintf("%s.json", uri), body)
	req := Req{
		HttpReq: httpReq,
		Refresh: true,
	}
	for _, mod := range mods {
		mod(&req)
	}
	return req
}

// NoRefresh prevents token refresh check.
// Primarily used by the Login and Refresh methods where this would be redundant.
func NoRefresh(req *Req) {
	req.Refresh = false
}

// Query sets an HTTP query parameter.
//   client.GetClass("fvBD", goaci.Query("query-target-filter", `eq(fvBD.name,"bd-name")`))
// Or set multiple parameters:
//   client.GetClass("fvBD",
//     goaci.Query("rsp-subtree-include", "faults"),
//     goaci.Query("query-target-filter", `eq(fvBD.name,"bd-name")`))
func Query(k, v string) func(req *Req) {
	return func(req *Req) {
		q := req.HttpReq.URL.Query()
		q.Add(k, v)
		req.HttpReq.URL.RawQuery = q.Encode()
	}
}
