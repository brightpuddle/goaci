// Package goaci is a a Cisco ACI client library for Go.
package goaci

import (
	"github.com/tidwall/gjson"
)

// Res is an API response returned by client requests.
type Res = gjson.Result
