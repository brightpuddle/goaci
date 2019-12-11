package backup

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// Memoize backup client to only read the testdata once.
func newTestClient() func() (Client, error) {
	bkup, err := NewClient("./testdata/config.tar.gz")
	return func() (Client, error) {
		return bkup, err
	}
}

var testClient = newTestClient()

// TestFmtRN tests the fmtRn function.
func TestFmtRN(t *testing.T) {
	record := gjson.Parse("{}")
	assert.Equal(t, "simple", fmtRn("simple", record))

	record = Body{}.Set("key2", "two").Res()
	assert.Equal(t, "one-two", fmtRn("one-{key2}", record))

	record = Body{}.Set("key2", "two").Res()
	assert.Equal(t, "one-[two]", fmtRn("one-{[key2]}", record))

	record = Body{}.Set("key2", "two").Set("key3", "three").Res()
	assert.Equal(t, "one-two-three", fmtRn("one-{key2}-{key3}", record))
}

// TestBuildDN tests the buildDn function.
func TestBuildDN(t *testing.T) {
	uniPath := []string{"uni"}

	// Record has existing DN so just return that
	path, _ := buildDn(Body{}.Set("dn", "uni/a/b").Res(), uniPath, "fvTenant")
	assert.ElementsMatch(t, path, []string{"uni", "a", "b"})

	// Simple key
	path, _ = buildDn(Body{}.Set("name", "a").Res(), uniPath, "fvTenant")
	assert.ElementsMatch(t, path, []string{"uni", "tn-a"})

	// Missing key
	_, err := buildDn(gjson.Parse("{}"), uniPath, "FakeTestClass")
	assert.Error(t, err)
}

// TestNewClient tests the NewClient function.
func TestNewClient(t *testing.T) {
	// Success use case already tested
	// Missing file
	_, err := NewClient("non.existent.file")
	assert.Error(t, err)

	// Not a gzip file
	_, err = NewClient("./testdata/config.json")
	assert.Error(t, err)

	// Valid gzip; invalid tar
	_, err = NewClient("./testdata/valid.gz.invalid.tar")
	assert.Error(t, err)

}

// TestClientGetDn tests the Client::GetDn method.
func TestClientGetDn(t *testing.T) {
	bkup, _ := testClient()

	// Valid dn
	res, _ := bkup.GetDn("uni/tn-a")
	if !assert.Equal(t, "a", res.Get("fvTenant.attributes.name").Str) {
		fmt.Println(res.Get("@pretty"))
	}

	// Valid dn constructed as child of uni
	res, _ = bkup.GetDn("uni/tn-b")
	if !assert.Equal(t, "uni/tn-b", res.Get("fvTenant.attributes.dn").Str) {
		fmt.Println(res.Get("@pretty"))
	}

	// Invalid dn
	_, err := bkup.GetDn("uni/non-existent")
	assert.Error(t, err)
}

// TestClientGetClass test the Client::GetClass method.
func TestClientGetClass(t *testing.T) {
	bkup, _ := testClient()

	// Valid class
	res, err := bkup.GetClass("fvTenant")
	assert.NoError(t, err)
	if !assert.Equal(t, 2, len(res.Array())) {
		fmt.Println(res.Get("@pretty"))
	}

	// Invalid class
	res, _ = bkup.GetClass("notExist")
	assert.NoError(t, err)
	if !assert.Equal(t, 0, len(res.Array())) {
		fmt.Println(res.Get("@pretty"))
	}
}
