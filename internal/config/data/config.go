// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"fmt"
	"io"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Config tracks a context configuration.
type Config struct {
	Context *Context `yaml:"k9s"`
	mx      sync.RWMutex
}

// NewConfig returns a new config.
func NewConfig(ct *api.Context) *Config {
	return &Config{
		Context: NewContextFromConfig(ct),
	}
}

func (c *Config) Merge(c1 *Config) {
	if c1 == nil {
		return
	}
	if c.Context != nil && c1.Context != nil {
		c.Context.merge(c1.Context)
	}
}

// Validate ensures config is in norms.
func (c *Config) Validate(conn client.Connection, ks KubeSettings) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.Context == nil {
		c.Context = NewContext()
	}
	c.Context.Validate(conn, ks)
}

// Dump used for debugging.
func (c *Config) Dump(w io.Writer) {
	bb, _ := yaml.Marshal(&c)

	fmt.Fprintf(w, "%s\n", string(bb))
}
