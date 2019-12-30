package client_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestNamespaced(t *testing.T) {
	uu := []struct {
		p, ns, n string
	}{
		{"fred/blee", "fred", "blee"},
		{"blee", "", "blee"},
	}

	for _, u := range uu {
		ns, n := client.Namespaced(u.p)
		assert.Equal(t, u.ns, ns)
		assert.Equal(t, u.n, n)
	}
}

func TestFQN(t *testing.T) {
	uu := []struct {
		ns, n string
		e     string
	}{
		{"fred", "blee", "fred/blee"},
		{"", "blee", "blee"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, client.FQN(u.ns, u.n))
	}
}
