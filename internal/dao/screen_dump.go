package dao

import (
	"os"

	"github.com/rs/zerolog/log"
)

type ScreenDump struct {
	Generic
}

var _ Accessor = &ScreenDump{}
var _ Nuker = &ScreenDump{}

// Delete a ScreenDump.
func (d *ScreenDump) Delete(path string, cascade, force bool) error {
	log.Debug().Msgf("ScreenDump DELETE %q", path)
	return os.Remove(path)
}
