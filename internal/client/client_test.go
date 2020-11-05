package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckCacheBool(t *testing.T) {
	c := NewTestAPIClient()

	const key = "fred"
	uu := map[string]struct {
		key                  string
		val                  interface{}
		found, actual, sleep bool
	}{
		"setTrue": {
			key:    key,
			val:    true,
			found:  true,
			actual: true,
		},
		"setFalse": {
			key:   key,
			val:   false,
			found: true,
		},
		"missing": {
			key: "blah",
			val: false,
		},
		"expired": {
			key:   key,
			val:   true,
			sleep: true,
		},
	}

	expiry := 1 * time.Millisecond
	for k := range uu {
		u := uu[k]
		c.cache.Add(key, u.val, expiry)
		if u.sleep {
			time.Sleep(expiry)
		}
		t.Run(k, func(t *testing.T) {
			val, ok := c.checkCacheBool(u.key)
			assert.Equal(t, u.found, ok)
			assert.Equal(t, u.actual, val)
		})
	}
}
