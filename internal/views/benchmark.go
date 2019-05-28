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

const benchFmat = "%s_%s_%d.txt"

// K9sBenchDir directory to store K9s benchmark files.
var K9sBenchDir = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-bench-%s", config.MustK9sUser()))

type (
	benchmark struct {
		canceled bool
		config   benchConfig
		worker   *requester.Work
	}

	benchConfig struct {
		Method, Path, URL string
		C, N              int
	}
)

func newBenchmark(cfg benchConfig) (*benchmark, error) {
	b := benchmark{config: cfg}
	return &b, b.init()
}

func (b *benchmark) init() error {
	req, err := http.NewRequest(b.config.Method, b.config.URL, nil)
	if err != nil {
		return err
	}

	b.worker = &requester.Work{
		Request: req,
		N:       b.config.N,
		C:       b.config.C,
		Output:  "",
	}

	return nil
}

func (b *benchmark) annuled() bool {
	return b.canceled
}

func (b *benchmark) cancel() {
	b.canceled = true
	b.worker.Stop()
}

func (b *benchmark) run(done func()) {
	buff := new(bytes.Buffer)
	b.worker.Writer = buff
	b.worker.Run()
	if !b.canceled {
		if err := b.save(buff); err != nil {
			log.Error().Err(err).Msg("Saving benchmark")
		}
	}
	done()
}

func (b *benchmark) save(r io.Reader) error {
	if err := os.MkdirAll(K9sBenchDir, 0777); err != nil {
		return err
	}

	ns, n := namespaced(b.config.Path)
	file := filepath.Join(K9sBenchDir, fmt.Sprintf(benchFmat, ns, n, time.Now().UnixNano()))
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
