package dao

import (
	"os"
)

// ScreenDump represents a scraped resources.
type ScreenDump struct {
	Generic
}

var _ Accessor = &ScreenDump{}
var _ Nuker = &ScreenDump{}

// Delete a ScreenDump.
func (d *ScreenDump) Delete(path string, cascade, force bool) error {
	return os.Remove(path)
}
