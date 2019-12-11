package backup

import (
	"github.com/tidwall/gjson"
)

// Res is an API response returned by client requests.
// This is a gjson.Result.
// https://github.com/tidwall/gjson
type Res = gjson.Result
