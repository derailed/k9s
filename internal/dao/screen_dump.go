package dao

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type ScreenDump struct {
	Resource
}

var _ Accessor = &ScreenDump{}
var _ Nuker = &ScreenDump{}

// Delete a ScreenDump.
func (d *ScreenDump) Delete(dir, sel string, cascade, force bool) error {
	log.Debug().Msgf("ScreenDump DELETE %q:%q", dir, sel)
	return os.Remove(filepath.Join("/"+dir, sel))
}
