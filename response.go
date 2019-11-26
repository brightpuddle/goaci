// Package aci is a a Cisco ACI client library.
package aci

import (
	"github.com/tidwall/gjson"
)

// Res is an API response returned by HTTP requests.
type Res = gjson.Result
