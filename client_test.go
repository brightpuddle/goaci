package goaci

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/sjson"
	"gopkg.in/h2non/gock.v1"
)

const (
	testUrl = "https://10.0.0.1"
)

func testClient() Client {
	client, _ := NewClient(testUrl, "usr", "pwd")
	gock.InterceptClient(client.httpClient)
	return client
}

// ErrReader implements the io.Reader interface and fails on Read.
type ErrReader struct{}

// Read mocks failing io.Reader test cases.
func (r ErrReader) Read(buf []byte) (int, error) {
	return 0, errors.New("fail")
}

// TestNewClient tests client initiation parameters.
func TestNewClient(t *testing.T) {
	client, _ := NewClient(testUrl, "usr", "pwd", RequestTimeout(120))
	assert.Equal(t, client.httpClient.Timeout, 120*time.Second)
}

// TestLogin tests the Login method.
func TestLogin(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Successful login
	gock.New(testUrl).Post("/api/aaaLogin.json").Reply(200)
	assert.NoError(t, client.Login())

	// Invalid HTTP status code
	gock.New(testUrl).Post("/api/aaaLogin.json").Reply(405)
	assert.Error(t, client.Login())

	// JSON error from APIC
	body, _ := sjson.Set("", "imdata.0.error.attributes.text", "error")
	gock.New(testUrl).
		Post("/api/aaaLogin.json").
		Reply(200).
		BodyString(body)
	assert.Error(t, client.Login())
}

// TestRefresh tests the Refresh method.
func TestRefresh(t *testing.T) {
	defer gock.Off()
	client := testClient()

	gock.New(testUrl).
		Get("/api/aaaRefresh.json").
		Reply(200)

	assert.NoError(t, client.Refresh())
}

// TestGet tests the Get method.
func TestGet(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Success
	gock.New(testUrl).Get("/url.json").Reply(200)
	_, err := client.Get("/url")
	assert.NoError(t, err)

	// HTTP error
	gock.New(testUrl).Get("/url.json").ReplyError(errors.New("fail"))
	_, err = client.Get("/url")
	assert.Error(t, err)

	// Invalid HTTP status code
	gock.New(testUrl).Get("/url.json").Reply(405)
	_, err = client.Get("/url")
	assert.Error(t, err)

	// Error decoding response body
	gock.New(testUrl).
		Get("/url.json").
		Reply(200).
		Map(func(res *http.Response) *http.Response {
			res.Body = ioutil.NopCloser(ErrReader{})
			return res
		})
	_, err = client.Get("/url")
	assert.Error(t, err)
}

// TestGetClass tests the GetClass method.
func TestGetClass(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Success
	var body string
	body, _ = sjson.Set(body, "imdata.0.test.attributes.i", 0)
	body, _ = sjson.Set(body, "imdata.1.test.attributes.i", 1)
	body, _ = sjson.Set(body, "imdata.2.other.attributes.i", 2)
	gock.New(testUrl).Get("/api/class/test.json").Reply(200).BodyString(body)
	res, _ := client.GetClass("test")
	assert.Len(t, res.Array(), 3)
	assert.Equal(t, `{"i":1}`, res.Get("1").Raw)

	// HTTP error
	gock.New(testUrl).Get("/api/class/test.json").ReplyError(errors.New("fail"))
	_, err := client.GetClass("test")
	assert.Error(t, err)
}

// TestGetDn tests the GetDn method.
func TestGetDn(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Success
	body, _ := sjson.Set("", "imdata.0.test.attributes.name", "test")
	gock.New(testUrl).Get("/api/mo/test.json").Reply(200).BodyString(body)
	res, _ := client.GetDn("test")
	assert.Equal(t, `{"name":"test"}`, res.Raw)

	// HTTP error
	gock.New(testUrl).Get("/api/class/test.json").ReplyError(errors.New("fail"))
	_, err := client.GetDn("test")
	assert.Error(t, err)
}

// TestPost tests the Post method.
func TestPost(t *testing.T) {
	defer gock.Off()
	client := testClient()

	var err error

	// Success
	gock.New(testUrl).Post("/url.json").Reply(200)
	_, err = client.Post("/url", "{}")
	assert.NoError(t, err)

	// HTTP error
	gock.New(testUrl).Post("/url.json").ReplyError(errors.New("fail"))
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)

	// Invalid HTTP status code
	gock.New(testUrl).Post("/url.json").Reply(405)
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)

	// Error decoding response body
	gock.New(testUrl).
		Post("/url.json").
		Reply(200).
		Map(func(res *http.Response) *http.Response {
			res.Body = ioutil.NopCloser(ErrReader{})
			return res
		})
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)
}
