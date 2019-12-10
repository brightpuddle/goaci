package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetRaw tests the Body::SetRaw method.
func TestSetRaw(t *testing.T) {
	name := Body{}.SetRaw("a", `{"name":"a"}`).gjson().Get("a.name").Str
	assert.Equal(t, "a", name)
}
