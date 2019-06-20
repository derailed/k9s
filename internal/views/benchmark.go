package views

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/rakyll/hey/requester"
	"github.com/rs/zerolog/log"
)

const (
	benchFmat = "%s_%s_%d.txt"
	k9sUA     = "k9s/0.0.7"
)

// K9sBenchDir directory to store K9s benchmark files.
var K9sBenchDir = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-bench-%s", config.MustK9sUser()))

type benchmark struct {
	canceled bool
	config   config.BenchConfig
	worker   *requester.Work
}

func newBenchmark(base string, cfg config.BenchConfig) (*benchmark, error) {
	b := benchmark{config: cfg}

	return &b, b.init(base)
}

func (b *benchmark) init(base string) error {
	req, err := http.NewRequest(b.config.HTTP.Method, base, nil)
	if err != nil {
		return err
	}

	if b.config.Auth.User != "" || b.config.Auth.Password != "" {
		req.SetBasicAuth(b.config.Auth.User, b.config.Auth.Password)
	}

	req.Header = b.config.HTTP.Headers
	ua := req.UserAgent()
	if ua == "" {
		ua = k9sUA
	} else {
		ua += " " + k9sUA
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("User-Agent", ua)

	b.worker = &requester.Work{
		Request:     req,
		RequestBody: []byte(b.config.HTTP.Body),
		N:           b.config.N,
		C:           b.config.C,
		H2:          b.config.HTTP.HTTP2,
		Output:      "",
	}

	return nil
}

func (b *benchmark) annulled() bool {
	return b.canceled
}

func (b *benchmark) cancel() {
	if b == nil {
		return
	}
	b.canceled = true
	b.worker.Stop()
}

func (b *benchmark) run(cluster string, done func()) {
	buff := new(bytes.Buffer)
	b.worker.Writer = buff
	b.worker.Run()
	if !b.canceled {
		if err := b.save(cluster, buff); err != nil {
			log.Error().Err(err).Msg("Saving benchmark")
		}
	}
	done()
}

func (b *benchmark) save(cluster string, r io.Reader) error {
	dir := filepath.Join(K9sBenchDir, cluster)
	if err := os.MkdirAll(dir, 0744); err != nil {
		return err
	}

	ns, n := namespaced(b.config.Name)
	file := filepath.Join(dir, fmt.Sprintf(benchFmat, ns, n, time.Now().UnixNano()))
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	bb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f.Write(bb)

	return nil
}
