// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package perf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/dao"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/rakyll/hey/requester"
	"github.com/rs/zerolog/log"
)

const (
	// BOZO!! Revisit bench and when we should timeout.
	benchTimeout = 2 * time.Minute
	benchFmat    = "%s_%s_%d.txt"
	k9sUA        = "k9s/"
)

var (
	// K9sBenchDir directory to store K9s Benchmark files.
	K9sBenchDir = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-bench-%s", config.MustK9sUser()))
)

// Benchmark puts a workload under load.
type Benchmark struct {
	canceled bool
	config   config.BenchConfig
	worker   *requester.Work
	cancelFn context.CancelFunc
	mx       sync.RWMutex
}

// NewBenchmark returns a new benchmark.
func NewBenchmark(base, version string, cfg config.BenchConfig) (*Benchmark, error) {
	b := Benchmark{config: cfg}
	if err := b.init(base, version); err != nil {
		return nil, err
	}
	return &b, nil
}

func (b *Benchmark) init(base, version string) error {
	var ctx context.Context
	ctx, b.cancelFn = context.WithTimeout(context.Background(), benchTimeout)
	req, err := http.NewRequestWithContext(ctx, b.config.HTTP.Method, base, nil)
	if err != nil {
		return err
	}
	if b.config.Auth.User != "" || b.config.Auth.Password != "" {
		req.SetBasicAuth(b.config.Auth.User, b.config.Auth.Password)
	}
	req.Header = b.config.HTTP.Headers
	log.Debug().Msgf("Benchmarking Request %s", req.URL.String())

	ua := req.UserAgent()
	if ua == "" {
		ua = k9sUA
	} else {
		ua += " " + k9sUA
	}
	ua += version
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("User-Agent", ua)

	log.Debug().Msgf("Using bench config N:%d--C:%d", b.config.N, b.config.C)

	b.worker = &requester.Work{
		Request:     req,
		RequestBody: []byte(b.config.HTTP.Body),
		N:           b.config.N,
		C:           b.config.C,
		H2:          b.config.HTTP.HTTP2,
	}

	return nil
}

// Cancel kills the benchmark in progress.
func (b *Benchmark) Cancel() {
	if b == nil {
		return
	}

	b.mx.Lock()
	defer b.mx.Unlock()
	b.canceled = true
	if b.cancelFn != nil {
		b.cancelFn()
		b.cancelFn = nil
	}
}

// Canceled checks if the benchmark was canceled.
func (b *Benchmark) Canceled() bool {
	return b.canceled
}

// Run starts a benchmark.
func (b *Benchmark) Run(cluster string, done func()) {
	log.Debug().Msgf("Running benchmark on cluster %s", cluster)
	buff := new(bytes.Buffer)
	b.worker.Writer = buff
	// this call will block until the benchmark is complete or times out.
	b.worker.Run()
	b.worker.Stop()
	if buff.Len() > 0 {
		if err := b.save(cluster, buff); err != nil {
			log.Error().Err(err).Msg("Saving Benchmark")
		}
	}
	done()
}

func (b *Benchmark) save(cluster string, r io.Reader) error {
	dir := filepath.Join(K9sBenchDir, cluster)
	if err := os.MkdirAll(dir, 0744); err != nil {
		return err
	}

	ns, n := client.Namespaced(b.config.Name)
	file := filepath.Join(dir, fmt.Sprintf(benchFmat, ns, dao.BenchRx.ReplaceAllString(n, "_"), time.Now().UnixNano()))
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer func() {
		if e := f.Close(); e != nil {
			log.Fatal().Err(e).Msg("Bench save")
		}
	}()

	if _, err = io.Copy(f, r); err != nil {
		return err
	}

	return nil
}
