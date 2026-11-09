// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigLogoConcurrentValidation(t *testing.T) {
	var cfg Config
	assert.Empty(t, cfg.Logo())
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 100 {
			cfg.Validate(nil, "context", "cluster")
		}
	})
	wg.Go(func() {
		for range 100 {
			assert.Empty(t, cfg.Logo())
		}
	})
	wg.Wait()
}

func TestConfigLogoConcurrentContextUpdates(t *testing.T) {
	cfg := Config{Context: NewContext()}
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 100 {
			cfg.Context.mx.Lock()
			cfg.Context.Logo = "CONTEXT"
			cfg.Context.mx.Unlock()
		}
	})
	wg.Go(func() {
		for range 100 {
			assert.Contains(t, []string{"", "CONTEXT"}, cfg.Logo())
		}
	})
	wg.Wait()
}
