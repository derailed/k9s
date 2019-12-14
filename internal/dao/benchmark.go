package dao

import (
	"os"

	"github.com/rs/zerolog/log"
)

// Benchmark represents a benchmark resource.
type Benchmark struct {
	Generic
}

var _ Accessor = &Benchmark{}
var _ Nuker = &Benchmark{}

// Delete a Benchmark.
func (d *Benchmark) Delete(path string, cascade, force bool) error {
	log.Debug().Msgf("Benchmark DELETE %q", path)
	return os.Remove(path)
}
