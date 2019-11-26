package aci

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/sjson"
	"gopkg.in/h2non/gock.v1"
)

const (
	testHost = "https://apic"
)

func testClient() Client {
	client, _ := NewClient(testHost, "usr", "pwd")
	gock.InterceptClient(client.httpClient)
	return client
}

// ErrReader implements the io.Reader interface and fails on Read.
type ErrReader struct{}

// Read mocks failing io.Reader test cases.
func (r ErrReader) Read(buf []byte) (int, error) {
	return 0, errors.New("fail")
}

// TestLogin tests the Login method.
func TestLogin(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Successful login
	gock.New(testHost).Post("/api/aaaLogin.json").Reply(200)
	assert.NoError(t, client.Login())

	// Invalid HTTP status code
	gock.New(testHost).Post("/api/aaaLogin.json").Reply(405)
	assert.Error(t, client.Login())

	// JSON error from APIC
	responseBody, _ := sjson.Set("", "imdata.0.error.attributes.text", "error")
	gock.New(testHost).
		Post("/api/aaaLogin.json").
		Reply(200).
		BodyString(responseBody)
	assert.Error(t, client.Login())
}

// TestRefresh tests the Refresh method.
func TestRefresh(t *testing.T) {
	defer gock.Off()
	client := testClient()

	gock.New(testHost).
		Get("/api/aaaRefresh.json").
		Reply(200)

	assert.NoError(t, client.Refresh())
}

// TestGet tests the Get method.
func TestGet(t *testing.T) {
	defer gock.Off()
	client := testClient()

	// Success
	gock.New(testHost).Get("/url.json").Reply(200)
	_, err := client.Get("/url")
	assert.NoError(t, err)

	// HTTP error
	gock.New(testHost).Get("/url.json").ReplyError(errors.New("fail"))
	_, err = client.Get("/url")
	assert.Error(t, err)

	// Invalid HTTP status code
	gock.New(testHost).Get("/url.json").Reply(405)
	_, err = client.Get("/url")
	assert.Error(t, err)

	// Error decoding response body
	gock.New(testHost).
		Get("/url.json").
		Reply(200).
		Map(func(res *http.Response) *http.Response {
			res.Body = ioutil.NopCloser(ErrReader{})
			return res
		})
	_, err = client.Get("/url")
	assert.Error(t, err)
}

// TestPost tests the Post method.
func TestPost(t *testing.T) {
	defer gock.Off()
	client := testClient()

	var err error

	// Success
	gock.New(testHost).Post("/url.json").Reply(200)
	_, err = client.Post("/url", "{}")
	assert.NoError(t, err)

	// HTTP error
	gock.New(testHost).Post("/url.json").ReplyError(errors.New("fail"))
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)

	// Invalid HTTP status code
	gock.New(testHost).Post("/url.json").Reply(405)
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)

	// Error decoding response body
	gock.New(testHost).
		Post("/url.json").
		Reply(200).
		Map(func(res *http.Response) *http.Response {
			res.Body = ioutil.NopCloser(ErrReader{})
			return res
		})
	_, err = client.Post("/url", "{}")
	assert.Error(t, err)
}

// TestQuery tests query parameters.
func TestQuery(t *testing.T) {
	defer gock.Off()
	client := testClient()

	gock.New(testHost).Get("/url").MatchParam("foo", "bar").Reply(200)
	_, err := client.Get("/url", Query("foo", "bar"))
	assert.NoError(t, err)
}
