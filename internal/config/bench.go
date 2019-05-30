package config

import (
	"io/ioutil"
	"net/http"

	"gopkg.in/yaml.v2"
)

var (
	// K9sBench the name of the benchmarks config file.
	K9sBench = "bench"
)

type (
	// Bench tracks K9s styling options.
	Bench struct {
		Benchmarks *Benchmarks `yaml:"benchmarks"`
	}

	// Benchmarks tracks K9s benchmarks configuration.
	Benchmarks struct {
		Defaults   Benchmark              `yaml:"defaults"`
		Services   map[string]BenchConfig `yam':"services"`
		Containers map[string]BenchConfig `yam':"containers"`
	}

	// Auth basic auth creds
	Auth struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	}

	// Benchmark represents a generic benchmark.
	Benchmark struct {
		C int `yaml:"concurrency"`
		N int `yaml:"requests"`
	}

	// BenchConfig represents a service benchmark.
	BenchConfig struct {
		C       int         `yaml:"concurrency"`
		N       int         `yaml:"requests"`
		Method  string      `yaml:"method"`
		Host    string      `yaml:"host"`
		Path    string      `yaml:"path"`
		HTTP2   bool        `yaml:"http2"`
		Body    string      `yaml:"body"`
		Auth    Auth        `yaml:"auth"`
		Headers http.Header `yaml:"headers"`
		Name    string
	}
)

const (
	// DefaultC default concurrency.
	DefaultC = 1
	// DefaultN default number of requests.
	DefaultN = 200
	// DefaultMethod default http verb.
	DefaultMethod = "GET"
)

func newBenchmark() Benchmark {
	return Benchmark{
		C: DefaultC,
		N: DefaultN,
	}
}

func (b Benchmark) empty() bool {
	return b.C == 0 && b.N == 0
}

func newBenchmarks() *Benchmarks {
	return &Benchmarks{
		Defaults: newBenchmark(),
	}
}

func newBench() *Bench {
	return &Bench{
		Benchmarks: newBenchmarks(),
	}
}

// NewBench creates a new default config.
func NewBench(path string) (*Bench, error) {
	s := &Bench{Benchmarks: newBenchmarks()}
	err := s.load(path)
	return s, err
}

// Reload update the configuration from disk.
func (s *Bench) Reload(path string) error {
	return s.load(path)
}

// Load K9s benchmark configs from file
func (s *Bench) load(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(f, &s)
}
