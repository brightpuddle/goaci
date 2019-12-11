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

// Client is an HTTP API client.
type Client struct {
	httpClient  *http.Client
	url         string
	usr         string
	pwd         string
	lastRefresh time.Time
}

// NewClient creates a new ACI HTTP client.
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

// RequestTimeout modifies the HTTP request timeout from the default.
func RequestTimeout(x time.Duration) func(*Client) {
	return func(client *Client) {
		client.httpClient.Timeout = x * time.Second
	}
}

// Get makes a GET request and returns a GJSON result.
func (client *Client) Get(path string, mods ...func(*Req)) (Res, error) {
	req := NewReq("GET", client.url+path, nil, mods...)

	if req.refresh && time.Now().Sub(client.lastRefresh) > 480*time.Second {
		if err := client.Refresh(); err != nil {
			return Res{}, err
		}
	}

	httpRes, err := client.httpClient.Do(req.httpReq)
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
func (client *Client) GetClass(class string, mods ...func(*Req)) (Res, error) {
	res, err := client.Get(fmt.Sprintf("/api/class/%s", class), mods...)
	if err != nil {
		return res, err
	}
	return res.Get("imdata"), nil
}

// GetDn makes a GET request by DN.
func (client *Client) GetDn(dn string, mods ...func(*Req)) (Res, error) {
	res, err := client.Get(fmt.Sprintf("/api/mo/%s", dn), mods...)
	if err != nil {
		return res, err
	}
	return res.Get("imdata.0"), nil
}

// Post makes a POST request and returns a GJSON result.
func (client *Client) Post(path, data string, mods ...func(*Req)) (Res, error) {
	req := NewReq("POST", client.url+path, strings.NewReader(data), mods...)
	if req.refresh && time.Now().Sub(client.lastRefresh) > 480*time.Second {
		if err := client.Refresh(); err != nil {
			return Res{}, err
		}
	}

	httpRes, err := client.httpClient.Do(req.httpReq)
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

// Login authenticates to the Client.
func (client *Client) Login() error {
	data := fmt.Sprintf(`{"aaaUser":{"attributes":{"name":"%s","pwd":"%s"}}}`,
		client.usr,
		client.pwd,
	)
	res, err := client.Post("/api/aaaLogin", data, NoRefresh)
	if err != nil {
		return err
	}
	errText := res.Get("imdata|0|error|attributes|text").Str
	if errText != "" {
		return errors.New("authentication error")
	}
	client.lastRefresh = time.Now()
	return nil
}

// Refresh refreshes the authentication token.
func (client *Client) Refresh() error {
	_, err := client.Get("/api/aaaRefresh", NoRefresh)
	if err != nil {
		return err
	}
	client.lastRefresh = time.Now()
	return nil
}
