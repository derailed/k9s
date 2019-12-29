package main

import (
	"os"

	"github.com/derailed/k9s/cmd"
	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"net/http"
	_ "net/http/pprof"
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

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: file})

	cmd.Execute()
}
