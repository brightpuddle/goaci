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

// APIC is an HTTP API client.
type APIC struct {
	httpClient  *http.Client
	url         string
	usr         string
	pwd         string
	lastRefresh time.Time
}

// NewAPIC creates a new ACI HTTP client.
func NewAPIC(url, usr, pwd string, mods ...func(*APIC)) (APIC, error) {

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

	apic := APIC{
		httpClient: &httpClient,
		url:        url,
		usr:        usr,
		pwd:        pwd,
	}
	for _, mod := range mods {
		mod(&apic)
	}
	return apic, nil
}

// RequestTimeout modifies the HTTP request timeout from the default.
func RequestTimeout(x time.Duration) func(*APIC) {
	return func(apic *APIC) {
		apic.httpClient.Timeout = x * time.Second
	}
}

// Get makes a GET request and returns a GJSON result.
func (apic APIC) Get(path string, mods ...func(*Req)) (Res, error) {
	req := NewReq("GET", apic.url+path, nil, mods...)

	if req.refresh && time.Now().Sub(apic.lastRefresh) > 480*time.Second {
		if err := apic.Refresh(); err != nil {
			return Res{}, err
		}
	}

	httpRes, err := apic.httpClient.Do(req.httpReq)
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
func (apic APIC) GetClass(class string, mods ...func(*Req)) (Res, error) {
	res, err := apic.Get(fmt.Sprintf("/api/class/%s", class), mods...)
	if err != nil {
		return res, err
	}
	return res.Get("imdata"), nil
}

// GetDn makes a GET request by DN.
func (apic APIC) GetDn(dn string, mods ...func(*Req)) (Res, error) {
	res, err := apic.Get(fmt.Sprintf("/api/mo/%s", dn), mods...)
	if err != nil {
		return res, err
	}
	return res.Get("imdata.0"), nil
}

// Post makes a POST request and returns a GJSON result.
func (apic APIC) Post(path, data string, mods ...func(*Req)) (Res, error) {
	req := NewReq("POST", apic.url+path, strings.NewReader(data), mods...)
	if req.refresh && time.Now().Sub(apic.lastRefresh) > 480*time.Second {
		if err := apic.Refresh(); err != nil {
			return Res{}, err
		}
	}

	httpRes, err := apic.httpClient.Do(req.httpReq)
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
func (apic APIC) Login() error {
	data := fmt.Sprintf(`{"aaaUser":{"attributes":{"name":"%s","pwd":"%s"}}}`,
		apic.usr,
		apic.pwd,
	)
	res, err := apic.Post("/api/aaaLogin", data, NoRefresh)
	if err != nil {
		return err
	}
	errText := res.Get("imdata|0|error|attributes|text").Str
	if errText != "" {
		return errors.New("authentication error")
	}
	apic.lastRefresh = time.Now()
	return nil
}

// Refresh refreshes the authentication token.
func (apic APIC) Refresh() error {
	_, err := apic.Get("/api/aaaRefresh", NoRefresh)
	return err
}
