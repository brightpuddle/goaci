// Package goaci is a a Cisco ACI client library for Go.
package goaci

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// Client is an HTTP ACI API client.
// Use goaci.NewClient to initiate a client.
// This will ensure proper cookie handling and processing of modifiers.
type Client struct {
	// HttpClient is the *http.Client used for API requests.
	HttpClient *http.Client
	// Url is the APIC IP or hostname, e.g. 10.0.0.1:80 (port is optional).
	Url string
	// Usr is the APIC username.
	Usr string
	// Pwd is the APIC password.
	Pwd string
	// LastRefresh is the timestamp of the last token refresh interval.
	LastRefresh time.Time
}

// NewClient creates a new ACI HTTP client.
// Pass modifiers in to modify the behavior of the client, e.g.
//  client, _ := NewClient("apic", "user", "password", RequestTimeout(120))
func NewClient(url, usr, pwd string, mods ...func(*Client)) (Client, error) {

	// Normalize the URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cookieJar, _ := cookiejar.New(nil)
	httpClient := http.Client{
		Timeout:   60 * time.Second,
		Transport: tr,
		Jar:       cookieJar,
	}

	client := Client{
		HttpClient: &httpClient,
		Url:        url,
		Usr:        usr,
		Pwd:        pwd,
	}
	for _, mod := range mods {
		mod(&client)
	}
	return client, nil
}

// RequestTimeout modifies the HTTP request timeout from the default of 60 seconds.
func RequestTimeout(x time.Duration) func(*Client) {
	return func(client *Client) {
		client.HttpClient.Timeout = x * time.Second
	}
}

// Get makes a GET request and returns a GJSON result.
// Results will be the raw data structure as returned by the APIC, wrapped in imdata, e.g.
//
//  {
// 	 "imdata": [
// 	  {
// 		 "fvTenant": {
// 		  "attributes": {
// 			 "dn": "uni/tn-mytenant",
// 			 "name": "mytenant",
// 		  }
// 		 }
// 	  }
// 	 ],
// 	  "totalCount": "1"
//  }
func (client *Client) Get(path string, mods ...func(*Req)) (Res, error) {
	req := NewReq("GET", client.Url+path, nil, mods...)

	if req.Refresh && time.Now().Sub(client.LastRefresh) > 480*time.Second {
		if err := client.Refresh(); err != nil {
			return Res{}, err
		}
	}

	httpRes, err := client.HttpClient.Do(req.HttpReq)
	if err != nil {
		return Res{}, err
	}
	defer httpRes.Body.Close()
	if httpRes.StatusCode != http.StatusOK {
		return Res{}, fmt.Errorf("received HTTP status %d", httpRes.StatusCode)
	}
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return Res{}, errors.New("cannot decode response body")
	}
	return Res(gjson.ParseBytes(body)), nil
}

// GetClass makes a GET request by class and unwraps the results.
// Result is removed from imdata, but still wrapped in Class.attributes, e.g.
//  [
//   {
// 	   "fvTenant": {
//	    "attributes": {
//       "dn": "uni/tn-mytenant",
//       "name": "mytenant",
//      }
//     }
//    }
//  ]
func (client *Client) GetClass(class string, mods ...func(*Req)) (Res, error) {
	res, err := client.Get(fmt.Sprintf("/api/class/%s", class), mods...)
	if err != nil {
		return res, err
	}
	return res.Get("imdata"), nil
}

// GetDn makes a GET request by DN.
// Result is removed from imdata and first result is removed from the list, e.g.
//  {
//   "fvTenant": {
//    "attributes": {
//     "dn": "uni/tn-mytenant",
//     "name": "mytenant",
//    }
//   }
// 	}
func (client *Client) GetDn(dn string, mods ...func(*Req)) (Res, error) {
	res, err := client.Get(fmt.Sprintf("/api/mo/%s", dn), mods...)
	if err != nil {
		return res, err
	}
	return res.Get("imdata.0"), nil
}

// Post makes a POST request and returns a GJSON result.
// Hint: Use the Body struct to easily create POST body data.
func (client *Client) Post(path, data string, mods ...func(*Req)) (Res, error) {
	req := NewReq("POST", client.Url+path, strings.NewReader(data), mods...)
	if req.Refresh && time.Now().Sub(client.LastRefresh) > 480*time.Second {
		if err := client.Refresh(); err != nil {
			return Res{}, err
		}
	}

	httpRes, err := client.HttpClient.Do(req.HttpReq)
	if err != nil {
		return Res{}, err
	}
	defer httpRes.Body.Close()
	if httpRes.StatusCode != http.StatusOK {
		return Res{}, fmt.Errorf("HTTP response: %s", httpRes.Status)
	}
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return Res{}, err
	}
	return Res(gjson.ParseBytes(body)), nil
}

// Login authenticates to the APIC.
func (client *Client) Login() error {
	data := fmt.Sprintf(`{"aaaUser":{"attributes":{"name":"%s","pwd":"%s"}}}`,
		client.Usr,
		client.Pwd,
	)
	res, err := client.Post("/api/aaaLogin", data, NoRefresh)
	if err != nil {
		return err
	}
	errText := res.Get("imdata|0|error|attributes|text").Str
	if errText != "" {
		return errors.New("authentication error")
	}
	client.LastRefresh = time.Now()
	return nil
}

// Refresh refreshes the authentication token.
// Note that this will be handled automatically be default.
// Refresh will be checked every request and the token will be refreshed after 8 minutes.
// Pass goaci.NoRefresh to prevent automatic refresh handling and handle it directly instead.
func (client *Client) Refresh() error {
	_, err := client.Get("/api/aaaRefresh", NoRefresh)
	if err != nil {
		return err
	}
	client.LastRefresh = time.Now()
	return nil
}
