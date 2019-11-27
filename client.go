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
func NewClient(url, usr, pwd string, mods ...func(*Client)) (Client, error) {

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
		httpClient: &httpClient,
		url:        url,
		usr:        usr,
		pwd:        pwd,
	}
	for _, mod := range mods {
		mod(&client)
	}
	return client, nil
}

// RequestTimeout modifies the request timeout from the default.
func RequestTimeout(x time.Duration) func(*Client) {
	return func(client *Client) {
		client.httpClient.Timeout = x * time.Second
	}
}

// Get makes a GET request and returns a GJSON result.
func (c Client) Get(path string, mods ...func(*Request)) (Response, error) {
	req := NewRequest("GET", c.url+path, nil, mods...)

	// TODO caching option.
	if req.refresh && time.Now().Sub(c.lastRefresh) > 480*time.Second {
		c.Refresh()
	}

	httpRes, err := c.httpClient.Do(req.httpReq)
	if err != nil {
		return Response{}, err
	}
	defer httpRes.Body.Close()
	if httpRes.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("received HTTP status %d", httpRes.StatusCode)
	}
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return Response{}, errors.New("cannot decode response body")
	}
	return Response(gjson.ParseBytes(body)), nil
}

// GetClass makes a GET request by class and unwraps the results.
func (c Client) GetClass(class string, mods ...func(*Request)) (Response, error) {
	res, err := c.Get(fmt.Sprintf("/api/class/%s", class), mods...)
	if err != nil {
		return res, err
	}
	return res.Get("imdata.#.*.attributes"), nil
}

// GetDn makes a GET request by DN and unwraps the result.
func (c Client) GetDn(dn string, mods ...func(*Request)) (Response, error) {
	res, err := c.Get(fmt.Sprintf("/api/mo/%s", dn), mods...)
	if err != nil {
		return res, err
	}
	return res.Get("imdata.0.*.attributes"), nil
}

// Post makes a POST request and returns a GJSON result.
func (c Client) Post(path, data string, mods ...func(*Request)) (Response, error) {
	req := NewRequest("POST", c.url+path, strings.NewReader(data), mods...)
	if req.refresh && time.Now().Sub(c.lastRefresh) > 480*time.Second {
		c.Refresh()
	}

	httpRes, err := c.httpClient.Do(req.httpReq)
	if err != nil {
		return Response{}, err
	}
	defer httpRes.Body.Close()
	if httpRes.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("HTTP response: %s", httpRes.Status)
	}
	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return Response{}, err
	}
	return Response(gjson.ParseBytes(body)), nil
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
