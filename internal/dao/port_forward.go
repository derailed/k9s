// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*PortForward)(nil)
	_ Nuker    = (*PortForward)(nil)
)

// PortForward represents a port forward dao.
type PortForward struct {
	NonResource
}

// Delete deletes a portforward.
func (p *PortForward) Delete(_ context.Context, path string, _ *metav1.DeletionPropagation, _ Grace) error {
	p.getFactory().DeleteForwarder(path)

	return nil
}

// List returns a collection of port forwards.
func (p *PortForward) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	benchFile, ok := ctx.Value(internal.KeyBenchCfg).(string)
	if !ok || benchFile == "" {
		return nil, fmt.Errorf("no benchmark config file found in context")
	}
	path, _ := ctx.Value(internal.KeyPath).(string)

	config, err := config.NewBench(benchFile)
	if err != nil {
		log.Debug().Msgf("No custom benchmark config file found: %q", benchFile)
	}

	ff, cc := p.getFactory().Forwarders(), config.Benchmarks.Containers
	oo := make([]runtime.Object, 0, len(ff))
	for k, f := range ff {
		if !strings.HasPrefix(k, path) {
			continue
		}
		cfg := render.BenchCfg{
			C: config.Benchmarks.Defaults.C,
			N: config.Benchmarks.Defaults.N,
		}
		if cust, ok := cc[PodToKey(k)]; ok {
			cfg.C, cfg.N = cust.C, cust.N
			cfg.Host, cfg.Path = cust.HTTP.Host, cust.HTTP.Path
		}
		oo = append(oo, render.ForwardRes{
			Forwarder: f,
			Config:    cfg,
		})
	}

	return oo, nil
}

// ----------------------------------------------------------------------------
// Helpers...

var podNameRX = regexp.MustCompile(`\A(.+)\-(\w{10})\-(\w{5})\z`)

// PodToKey converts a pod path to a generic bench config key.
func PodToKey(path string) string {
	tokens := strings.Split(path, "|")
	ns, po := client.Namespaced(tokens[0])
	sections := podNameRX.FindStringSubmatch(po)
	if len(sections) >= 1 {
		po = sections[1]
	}
	return client.FQN(ns, po) + ":" + tokens[1]
}

// BenchConfigFor returns a custom bench spec if defined otherwise returns the default one.
func BenchConfigFor(benchFile, path string) config.BenchConfig {
	def := config.DefaultBenchSpec()
	cust, err := config.NewBench(benchFile)
	if err != nil {
		log.Debug().Msgf("No custom benchmark config file found. Using default: %q", benchFile)
		return def
	}
	if b, ok := cust.Benchmarks.Containers[PodToKey(path)]; ok {
		return b
	}

	def.C, def.N = cust.Benchmarks.Defaults.C, cust.Benchmarks.Defaults.N
	return def
}
