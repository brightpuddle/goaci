package aci

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/tidwall/gjson"
	"golang.org/x/crypto/ssh/terminal"
)

// Config is the required APIC info to create a client.
// If not provided, the library will prompt for input.
type Config struct {
	// IP is the APIC IP address or resolvable hostname
	IP string
	// Password is the APIC password
	Password string
	// Username is the APIC username
	Username string
	// RequestTimeout defaults to 60 seconds
	RequestTimeout time.Duration
}

// Client is an HTTP client for the ACI API.
type Client struct {
	httpClient *http.Client
	cfg        *Config
}

// Req is an API request.
type Req struct {
	// URI is the fragment of the URL between the IP and .json.
	// e.g. /api/class/fvTenant
	URI string
	// Query is a list of query parameters in string form.
	// e.g. rsp-subtree-include=count
	Query []string
}

// Res is an API result.
// Alias for gjson.Result
type Res = gjson.Result

func input(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s ", prompt)
	input, _ := reader.ReadString('\n')
	return strings.Trim(input, "\r\n")
}

func (c *Config) validate() {
	if c.IP == "" {
		c.IP = input("APIC IP:")
	}
	if c.Username == "" {
		c.Username = input("Username:")
	}
	if c.Password == "" {
		fmt.Print("Password: ")
		pwd, _ := terminal.ReadPassword(int(syscall.Stdin))
		c.Password = string(pwd)
	}
	if c.RequestTimeout == 0 {
		c.RequestTimeout = 30
	}
}

// NewClient creates a new ACI API client
func NewClient(cfg Config) Client {
	cfg.validate()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	cookieJar, _ := cookiejar.New(nil)
	httpClient := http.Client{
		Timeout: time.Second * cfg.RequestTimeout,
		Jar:     cookieJar,
	}
	return Client{
		httpClient: &httpClient,
		cfg:        &cfg,
	}
}

func (c *Client) newURL(req Req) string {
	result := fmt.Sprintf("https://%s%s.json", c.cfg.IP, req.URI)
	if len(req.Query) > 0 {
		return fmt.Sprintf("%s?%s", result, strings.Join(req.Query, "&"))
	}
	return result
}

// GetURI returns a GJSON result. Shortcut for Get with no query parameters.
func (c *Client) GetURI(s string) (Res, error) {
	return c.Get(Req{URI: s})
}

// Get makes a request and returns a GJSON result.
func (c *Client) Get(req Req) (Res, error) {
	url := c.newURL(req)
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

// Login authenticates to the APIC and returns an error
func (c *Client) Login() error {
	uri := "/api/aaaLogin"
	url := c.newURL(Req{URI: uri})
	data := fmt.Sprintf(`{"aaaUser":{"attributes":{"name":"%s","pwd":"%s"}}}`,
		c.cfg.Username, c.cfg.Password)
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
	return nil
}

// Refresh updates the authentication token.
// By default, this will time out after ten minutes.
func (c *Client) Refresh() error {
	_, err := c.Get(Req{URI: "/api/aaaRefresh"})
	return err
}
