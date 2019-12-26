package dao

import (
	"os"
)

// Benchmark represents a benchmark resource.
type Benchmark struct {
	Generic
}

var _ Accessor = &Benchmark{}
var _ Nuker = &Benchmark{}

// Delete a Benchmark.
func (d *Benchmark) Delete(path string, cascade, force bool) error {
	return os.Remove(path)
}
