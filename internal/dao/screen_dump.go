package dao

import (
	"os"
)

type ScreenDump struct {
	Generic
}

var _ Accessor = &ScreenDump{}
var _ Nuker = &ScreenDump{}

// Delete a ScreenDump.
func (d *ScreenDump) Delete(path string, cascade, force bool) error {
	return os.Remove(path)
}
