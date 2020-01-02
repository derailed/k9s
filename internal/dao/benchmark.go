package dao

import (
	"os"
)

// Benchmark represents a benchmark resource.
type Benchmark struct {
	Generic
}

var _ Accessor = (*Benchmark)(nil)
var _ Nuker = (*Benchmark)(nil)

// Delete a Benchmark.
func (d *Benchmark) Delete(path string, cascade, force bool) error {
	return os.Remove(path)
}
