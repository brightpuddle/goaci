package goaci

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

// TestQuery tests query parameters.
func TestQuery(t *testing.T) {
	defer gock.Off()
	client := testClient()

	gock.New(testUrl).Get("/url").MatchParam("foo", "bar").Reply(200)
	_, err := client.Get("/url", Query("foo", "bar"))
	assert.NoError(t, err)
}
