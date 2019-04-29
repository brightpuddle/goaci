package aci

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/tidwall/gjson"
	"golang.org/x/crypto/ssh/terminal"
)

// Config : APIC configuration object from CLI, config file, etc
type Config struct {
	IP             string
	Password       string
	Username       string
	Logger         Logger
	RequestTimeout time.Duration
}

// Logger : common logging interface for logrus, etc
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})
}

// Client : httpClient wrapper
type Client struct {
	httpClient *http.Client
	cfg        *Config
	log        Logger
}

// Req : API request
type Req struct {
	URI   string
	Query []string
}

// Res : API request result
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

// NewClientFromCLI : Create new ACI client struct
func NewClientFromCLI(cfg Config) Client {
	cfg.validate()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		log.Panic(err)
	}
	httpClient := http.Client{
		Timeout: time.Second * cfg.RequestTimeout,
		Jar:     cookieJar,
	}
	return Client{
		httpClient: &httpClient,
		cfg:        &cfg,
		log:        cfg.Logger,
	}
}

func (c *Client) newURL(req Req) string {
	result := fmt.Sprintf("https://%s%s.json", c.cfg.IP, req.URI)
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
func (c *Client) Login() error {
	uri := "/api/aaaLogin"
	url := c.newURL(Req{URI: uri})
	data := fmt.Sprintf(`{"aaaUser":{"attributes":{"name":"%s","pwd":"%s"}}}`,
		c.cfg.Username, c.cfg.Password)
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
func (c *Client) Refresh() error {
	_, err := c.Get(Req{URI: "/api/aaaRefresh"})
	return err
}
