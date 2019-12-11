package goaci

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

const (
	testHost = "10.0.0.1"
	testURL  = "https://" + testHost
)

func testClient() Client {
	client, _ := NewClient(testHost, "usr", "pwd")
	client.LastRefresh = time.Now()
	gock.InterceptClient(client.HttpClient)
	return client
}

// ErrReader implements the io.Reader interface and fails on Read.
type ErrReader struct{}

// Read mocks failing io.Reader test cases.
func (r ErrReader) Read(buf []byte) (int, error) {
	return 0, errors.New("fail")
}

// TestNewClient tests the NewClient function.
func TestNewClient(t *testing.T) {
	client, _ := NewClient(testURL, "usr", "pwd", RequestTimeout(120))
	assert.Equal(t, client.HttpClient.Timeout, 120*time.Second)
}

// TestClientLogin tests the Client::Login method.
func TestClientLogin(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Successful login
	gock.New(testURL).Post("/api/aaaLogin.json").Reply(200)
	assert.NoError(t, client.Login())

	// Invalid HTTP status code
	gock.New(testURL).Post("/api/aaaLogin.json").Reply(405)
	assert.Error(t, client.Login())

	// JSON error from Client
	gock.New(testURL).
		Post("/api/aaaLogin.json").
		Reply(200).
		BodyString(Body{}.Set("imdata.0.error.attributes.text", "error").Str)
	assert.Error(t, client.Login())
}

// TestClientRefresh tests the Client::Refresh method.
func TestClientRefresh(t *testing.T) {
	defer gock.Off()
	client := testClient()

	gock.New(testURL).Get("/api/aaaRefresh.json").Reply(200)
	assert.NoError(t, client.Refresh())
}

// TestClientGet tests the Client::Get method.
func TestClientGet(t *testing.T) {
	defer gock.Off()
	client := testClient()
	var err error

	// Success
	gock.New(testURL).Get("/url.json").Reply(200)
	_, err = client.Get("/url")
	assert.NoError(t, err)

	// HTTP error
	gock.New(testURL).Get("/url.json").ReplyError(errors.New("fail"))
	_, err = client.Get("/url")
	assert.Error(t, err)

	// Invalid HTTP status code
	gock.New(testURL).Get("/url.json").Reply(405)
	_, err = client.Get("/url")
	assert.Error(t, err)

	// Error decoding response body
	gock.New(testURL).
		Get("/url.json").
		Reply(200).
		Map(func(res *http.Response) *http.Response {
			res.Body = ioutil.NopCloser(ErrReader{})
			return res
		})
	_, err = client.Get("/url")
	assert.Error(t, err)

	// Force token refresh and throw an error
	client.LastRefresh = time.Now().AddDate(0, 0, -1)
	gock.New(testURL).
		Get("/api/aaaRefresh.json").
		ReplyError(errors.New("fail"))
	_, err = client.Get("/url")
	assert.Error(t, err)
}

// TestClientGetClass tests the Client::GetClass method.
func TestClientGetClass(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Success
	gock.New(testURL).
		Get("/api/class/fvTenant.json").
		Reply(200).
		BodyString(Body{}.
			Set("imdata.0.fvTenant.attributes.name", "zero").
			Set("imdata.1.fvTenant.attributes.name", "one").
			Str)
	res, _ := client.GetClass("fvTenant")
	if !assert.Len(t, res.Array(), 2) {
		fmt.Println(res.Get("@pretty"))
	}
	if !assert.Equal(t, "one", res.Get("1.fvTenant.attributes.name").Str) {
		fmt.Println(res.Get("@pretty"))
	}

	// HTTP error
	gock.New(testURL).Get("/api/class/test.json").ReplyError(errors.New("fail"))
	_, err := client.GetClass("test")
	assert.Error(t, err)
}

// TestClientGetDn tests the Client::GetDn method.
func TestClientGetDn(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Success
	gock.New(testURL).
		Get("/api/mo/uni/tn-test.json").
		Reply(200).
		BodyString(Body{}.Set("imdata.0.fvTenant.attributes.name", "test").Str)
	res, _ := client.GetDn("uni/tn-test")
	if !assert.Equal(t, "test", res.Get("fvTenant.attributes.name").Str) {
		fmt.Println(res.Get("@pretty"))
	}

	// HTTP error
	gock.New(testURL).
		Get("/api/mo/uni/fail.json").
		ReplyError(errors.New("fail"))
	_, err := client.GetDn("uni/fail")
	assert.Error(t, err)
}

// TestClientPost tests the Client::Post method.
func TestClientPost(t *testing.T) {
	defer gock.Off()
	client := testClient()

	var err error

	// Success
	gock.New(testURL).Post("/url.json").Reply(200)
	_, err = client.Post("/url", "{}")
	assert.NoError(t, err)

	// HTTP error
	gock.New(testURL).Post("/url.json").ReplyError(errors.New("fail"))
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)

	// Invalid HTTP status code
	gock.New(testURL).Post("/url.json").Reply(405)
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)

	// Error decoding response body
	gock.New(testURL).
		Post("/url.json").
		Reply(200).
		Map(func(res *http.Response) *http.Response {
			res.Body = ioutil.NopCloser(ErrReader{})
			return res
		})
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)

	// Force token refresh and throw an error
	client.LastRefresh = time.Now().AddDate(0, 0, -1)
	gock.New(testURL).Get("/api/aaaRefresh.json").ReplyError(errors.New("fail"))
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)
}
