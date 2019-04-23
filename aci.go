package aci

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// Config : APIC configuration object from CLI, config file, etc
type Config struct {
	IP       string
	Password string
	Username string
	Logger   Logger
}

// Logger : common logging interface for logrus, etc
type Logger interface {
	Debug(interface{})
	Info(interface{})
	Warn(interface{})
	Error(interface{})
	Panic(interface{})
}

// Client : httpClient wrapper
type Client struct {
	httpClient *http.Client
	config     Config
	log        Logger
}

// Req : API request
type Req struct {
	URI   string
	Query []string
}

// Res : API request result
type Res = gjson.Result

// NewClient : Create new ACI client struct
func NewClient(config Config) Client {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		log.Panic(err)
	}
	httpClient := http.Client{
		Timeout: time.Second * 30,
		Jar:     cookieJar,
	}
	return Client{
		httpClient: &httpClient,
		config:     config,
		log:        config.Logger,
	}
}

func (c Client) newURL(req Req) string {
	result := fmt.Sprintf("https://%s%s.json", c.config.IP, req.URI)
	if len(req.Query) > 0 {
		return fmt.Sprintf("%s?%s", result, strings.Join(req.Query, "&"))
	}
	return result
}

// GetURI : Shortcut for GET request with no query parameters
func (c *Client) GetURI(s string) (Res, error) {
	return c.Get(Req{URI: s})
}

// Get : APIC get request
func (c *Client) Get(req Req) (Res, error) {
	url := c.newURL(req)
	c.log.Debug(fmt.Sprintf("GET request to %s", req.URI))
	httpRes, err := c.httpClient.Get(url)
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
	return Res(gjson.GetBytes(body, "imdata")), nil
}

// Login : Login to the APIC
func (c Client) Login() error {
	uri := "/api/aaaLogin"
	url := c.newURL(Req{URI: uri})
	data := fmt.Sprintf(`{"aaaUser":{"attributes":{"name":"%s","pwd":"%s"}}}`,
		c.config.Username, c.config.Password)
	c.log.Debug(fmt.Sprintf("GET request to %s", uri))
	res, err := c.httpClient.Post(url, "json", strings.NewReader(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP response: %s", res.Status)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	errText := gjson.GetBytes(body, "imdata|0|error|attributes|text").Str
	if errText != "" {
		return errors.New("authentication error")
	}
	c.log.Info("Authentication successful.")
	return nil
}

// Refresh : Refresh the auth token
func (c Client) Refresh() error {
	_, err := c.Get(Req{URI: "/api/aaaRefresh"})
	return err
}
