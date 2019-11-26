// Package aci is a a Cisco ACI client library.
package aci

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Client is an API client.
type Client struct {
	httpClient  *http.Client
	url         string
	usr         string
	pwd         string
	lastRefresh time.Time
}

// NewClient creates a new API client.
func NewClient(url, usr, pwd string) (Client, error) {
	var requestTimeout time.Duration = 60

	// disable unsigned cert check
	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
	// 	InsecureSkipVerify: true,
	// }
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cookieJar, _ := cookiejar.New(nil)
	httpClient := http.Client{
		Timeout:   time.Second * requestTimeout,
		Transport: tr,
		Jar:       cookieJar,
	}

	return Client{
		httpClient: &httpClient,
		url:        url,
		usr:        usr,
		pwd:        pwd,
	}, nil
}

// NewReq createa a new Req against this client.
func (c Client) NewReq(method, urn string, body io.Reader) Req {
	uri := fmt.Sprintf("%s%s.json", c.url, urn)
	httpReq, _ := http.NewRequest(method, uri, body)
	return Req{
		httpReq: httpReq,
		refresh: true,
	}
}

// Get makes a GET request and returns a GJSON result.
func (c Client) Get(urn string, options ...func(*Req)) (Res, error) {
	req := c.NewReq("GET", urn, nil)
	for _, option := range options {
		option(&req)
	}
	// TODO caching option.
	if req.refresh && time.Now().Sub(c.lastRefresh) > 480*time.Second {
		c.Refresh()
	}

	httpRes, err := c.httpClient.Do(req.httpReq)
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

// Post makes a POST request and returns a GJSON result.
func (c Client) Post(urn, data string, options ...func(*Req)) (Res, error) {
	req := c.NewReq("POST", urn, strings.NewReader(data))
	for _, option := range options {
		option(&req)
	}
	if req.refresh && time.Now().Sub(c.lastRefresh) > 480*time.Second {
		c.Refresh()
	}

	httpRes, err := c.httpClient.Do(req.httpReq)
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
func (c Client) Login() error {
	data, _ := sjson.Set("", "aaaUser.attributes", map[string]string{
		"name": c.usr,
		"pwd":  c.pwd,
	})
	res, err := c.Post("/api/aaaLogin", data, NoRefresh)
	if err != nil {
		return err
	}
	errText := res.Get("imdata|0|error|attributes|text").Str
	if errText != "" {
		return errors.New("authentication error")
	}
	c.lastRefresh = time.Now()
	return nil
}

// Refresh refreshes the authentication token.
func (c Client) Refresh() error {
	_, err := c.Get("/api/aaaRefresh", NoRefresh)
	return err
}
