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

func testAPIC() APIC {
	client, _ := NewAPIC(testHost, "usr", "pwd")
	client.lastRefresh = time.Now()
	gock.InterceptClient(client.httpClient)
	return client
}

// ErrReader implements the io.Reader interface and fails on Read.
type ErrReader struct{}

// Read mocks failing io.Reader test cases.
func (r ErrReader) Read(buf []byte) (int, error) {
	return 0, errors.New("fail")
}

// TestNewAPIC tests the NewAPIC function.
func TestNewAPIC(t *testing.T) {
	apic, _ := NewAPIC(testURL, "usr", "pwd", RequestTimeout(120))
	assert.Equal(t, apic.httpClient.Timeout, 120*time.Second)
}

// TestAPICLogin tests the APIC::Login method.
func TestAPICLogin(t *testing.T) {
	defer gock.Off()
	apic := testAPIC()

	// Successful login
	gock.New(testURL).Post("/api/aaaLogin.json").Reply(200)
	assert.NoError(t, apic.Login())

	// Invalid HTTP status code
	gock.New(testURL).Post("/api/aaaLogin.json").Reply(405)
	assert.Error(t, apic.Login())

	// JSON error from APIC
	gock.New(testURL).
		Post("/api/aaaLogin.json").
		Reply(200).
		BodyString(Body{}.Set("imdata.0.error.attributes.text", "error").Str)
	assert.Error(t, apic.Login())
}

// TestAPICRefresh tests the APIC::Refresh method.
func TestAPICRefresh(t *testing.T) {
	defer gock.Off()
	apic := testAPIC()

	gock.New(testURL).Get("/api/aaaRefresh.json").Reply(200)
	assert.NoError(t, apic.Refresh())
}

// TestAPICGet tests the APIC::Get method.
func TestAPICGet(t *testing.T) {
	defer gock.Off()
	apic := testAPIC()
	var err error

	// Success
	gock.New(testURL).Get("/url.json").Reply(200)
	_, err = apic.Get("/url")
	assert.NoError(t, err)

	// HTTP error
	gock.New(testURL).Get("/url.json").ReplyError(errors.New("fail"))
	_, err = apic.Get("/url")
	assert.Error(t, err)

	// Invalid HTTP status code
	gock.New(testURL).Get("/url.json").Reply(405)
	_, err = apic.Get("/url")
	assert.Error(t, err)

	// Error decoding response body
	gock.New(testURL).
		Get("/url.json").
		Reply(200).
		Map(func(res *http.Response) *http.Response {
			res.Body = ioutil.NopCloser(ErrReader{})
			return res
		})
	_, err = apic.Get("/url")
	assert.Error(t, err)

	// Force token refresh and throw an error
	apic.lastRefresh = time.Now().AddDate(0, 0, -1)
	gock.New(testURL).
		Get("/api/aaaRefresh.json").
		ReplyError(errors.New("fail"))
	_, err = apic.Get("/url")
	assert.Error(t, err)
}

// TestAPICGetClass tests the APIC::GetClass method.
func TestAPICGetClass(t *testing.T) {
	defer gock.Off()
	apic := testAPIC()

	// Success
	gock.New(testURL).
		Get("/api/class/fvTenant.json").
		Reply(200).
		BodyString(Body{}.
			Set("imdata.0.fvTenant.attributes.name", "zero").
			Set("imdata.1.fvTenant.attributes.name", "one").
			Str)
	res, _ := apic.GetClass("fvTenant")
	if !assert.Len(t, res.Array(), 2) {
		fmt.Println(res.Get("@pretty"))
	}
	if !assert.Equal(t, "one", res.Get("1.fvTenant.attributes.name").Str) {
		fmt.Println(res.Get("@pretty"))
	}

	// HTTP error
	gock.New(testURL).Get("/api/class/test.json").ReplyError(errors.New("fail"))
	_, err := apic.GetClass("test")
	assert.Error(t, err)
}

// TestAPICGetDn tests the APIC::GetDn method.
func TestAPICGetDn(t *testing.T) {
	defer gock.Off()
	apic := testAPIC()

	// Success
	gock.New(testURL).
		Get("/api/mo/uni/tn-test.json").
		Reply(200).
		BodyString(Body{}.Set("imdata.0.fvTenant.attributes.name", "test").Str)
	res, _ := apic.GetDn("uni/tn-test")
	if !assert.Equal(t, "test", res.Get("fvTenant.attributes.name").Str) {
		fmt.Println(res.Get("@pretty"))
	}

	// HTTP error
	gock.New(testURL).
		Get("/api/mo/uni/fail.json").
		ReplyError(errors.New("fail"))
	_, err := apic.GetDn("uni/fail")
	assert.Error(t, err)
}

// TestAPICPost tests the APIC::Post method.
func TestAPICPost(t *testing.T) {
	defer gock.Off()
	apic := testAPIC()

	var err error

	// Success
	gock.New(testURL).Post("/url.json").Reply(200)
	_, err = apic.Post("/url", "{}")
	assert.NoError(t, err)

	// HTTP error
	gock.New(testURL).Post("/url.json").ReplyError(errors.New("fail"))
	_, err = apic.Post("/url", "{}")
	assert.Error(t, err)

	// Invalid HTTP status code
	gock.New(testURL).Post("/url.json").Reply(405)
	_, err = apic.Post("/url", "{}")
	assert.Error(t, err)

	// Error decoding response body
	gock.New(testURL).
		Post("/url.json").
		Reply(200).
		Map(func(res *http.Response) *http.Response {
			res.Body = ioutil.NopCloser(ErrReader{})
			return res
		})
	_, err = apic.Post("/url", "{}")
	assert.Error(t, err)

	// Force token refresh and throw an error
	apic.lastRefresh = time.Now().AddDate(0, 0, -1)
	gock.New(testURL).Get("/api/aaaRefresh.json").ReplyError(errors.New("fail"))
	_, err = apic.Post("/url", "{}")
	assert.Error(t, err)
}
