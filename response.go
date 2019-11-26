// Package aci is a a Cisco ACI client library.
package aci

import (
	"github.com/tidwall/gjson"
)

// Response is an API response returned by HTTP requests.
type Response = gjson.Result
