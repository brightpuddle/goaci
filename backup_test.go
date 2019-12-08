// Package goaci is a a Cisco ACI client library for Go.
package goaci

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

// Memoize backup client to only read the testdata once.
func newTestBackup() func() (Backup, error) {
	bkup, err := NewBackup("./testdata/config.tar.gz")
	return func() (Backup, error) {
		return bkup, err
	}
}

var testBackup = newTestBackup()

// A testing convenience for generating GJSON from the Body.
func (body Body) gjson() gjson.Result {
	return gjson.Parse(body.Str)
}

// TestFmtRN tests the fmtRn function.
func TestFmtRN(t *testing.T) {
	record := gjson.Parse("{}")
	assert.Equal(t, "simple", fmtRn("simple", record))

	record = Body{}.Set("key2", "two").gjson()
	assert.Equal(t, "one-two", fmtRn("one-{key2}", record))

	record = Body{}.Set("key2", "two").gjson()
	assert.Equal(t, "one-[two]", fmtRn("one-{[key2]}", record))

	record = Body{}.Set("key2", "two").Set("key3", "three").gjson()
	assert.Equal(t, "one-two-three", fmtRn("one-{key2}-{key3}", record))
}

// TestBuildDN tests the buildDn function.
func TestBuildDN(t *testing.T) {
	uniPath := []string{"uni"}

	// Record has existing DN so just return that
	path, _ := buildDn(Body{}.Set("dn", "uni/a/b").gjson(), uniPath, "fvTenant")
	assert.ElementsMatch(t, path, []string{"uni", "a", "b"})

	// Simple key
	path, _ = buildDn(Body{}.Set("name", "a").gjson(), uniPath, "fvTenant")
	assert.ElementsMatch(t, path, []string{"uni", "tn-a"})

	// Missing key
	_, err := buildDn(gjson.Parse("{}"), uniPath, "FakeTestClass")
	assert.Error(t, err)
}

// TestNewBackup tests the NewBackup function.
func TestNewBackup(t *testing.T) {
	// Success use case already tested
	// Missing file
	_, err := NewBackup("non.existent.file")
	assert.Error(t, err)

	// Not a gzip file
	_, err = NewBackup("./testdata/config.json")
	assert.Error(t, err)

	// Valid gzip; invalid tar
	_, err = NewBackup("./testdata/valid.gz.invalid.tar")
	assert.Error(t, err)

}

// TestBackupGetDn tests the Backup::GetDn method.
func TestBackupGetDn(t *testing.T) {
	bkup, _ := testBackup()

	// Valid dn
	res, _ := bkup.GetDn("uni/tn-a")
	if !assert.Equal(t, "a", res.Get("fvTenant.attributes.name").Str) {
		fmt.Println(res.Get("@pretty"))
	}

	// Invalid dn
	_, err := bkup.GetDn("uni/non-existent")
	assert.Error(t, err)
}

// TestBackupGetClass test the Backup::GetClass method.
func TestBackupGetClass(t *testing.T) {
	bkup, _ := testBackup()

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
