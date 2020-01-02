package dao

import (
	"os"
)

// ScreenDump represents a scraped resources.
type ScreenDump struct {
	Generic
}

var _ Accessor = (*ScreenDump)(nil)
var _ Nuker = (*ScreenDump)(nil)

// Delete a ScreenDump.
func (d *ScreenDump) Delete(path string, cascade, force bool) error {
	return os.Remove(path)
}
