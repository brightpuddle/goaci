package backup

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Body wraps SJSON for building JSON body strings.
// Example:
//  Body{}.Set("fvTenant.attributes.name", "mytenant").Str
// Creates:
//	{
//		"fvTenant":
//			"attributes": {
//			"name": "mytenant"
//		}
//	}
// These can also be chained:
//	Body{}.
//		Set("fvTenant.attributes.name", "mytenant").
//		Set("fvTenant.attributes.descr", "This is my tenant").
//		Str
// Creates:
//	{
//		"fvTenant": {
//			"attributes": {
//				"name": "mytenant",
//				"descr": "This is my tenant"
//			}
//		}
//	}
type Body struct {
	Str string
}

// Set sets a JSON path to a value.
// These can be a single key or nested path.
func (body Body) Set(path, value string) Body {
	res, _ := sjson.Set(body.Str, path, value)
	body.Str = res
	return body
}

// SetRaw sets a JSON path to a raw string value.
// Use this for nesting JSON, e.g.:
//	Body{}.SetRaw("fvTenant.attributes", Body{}.Set("name", "mytenant").Str).Str
func (body Body) SetRaw(path, rawValue string) Body {
	res, _ := sjson.SetRaw(body.Str, path, rawValue)
	body.Str = res
	return body
}

// Res generates a backup.Res (gjson.Result) object.
// Use the .Str struct property to access the string data, or .Res() to generate a GJSON result.
func (body Body) Res() Res {
	return gjson.Parse(body.Str)
}

// Req is a backup.Req request object.
type Req struct{}
