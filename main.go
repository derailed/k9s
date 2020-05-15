package main

import (
	"os"

	_ "net/http/pprof"

	"github.com/derailed/k9s/cmd"
	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func init() {
	config.EnsurePath(config.K9sLogs, config.DefaultDirMod)
}

func main() {
	mod := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	file, err := os.OpenFile(config.K9sLogs, mod, config.DefaultFileMod)
	if err != nil {
		panic(err)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: file})

	cmd.Execute()
}
