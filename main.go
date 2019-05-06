package main

import (
	"os"
	"syscall"

	"github.com/derailed/k9s/cmd"
	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog"
)

func init() {
	config.EnsurePath(config.K9sLogs, config.DefaultDirMod)

	mod := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	if file, err := os.OpenFile(config.K9sLogs, mod, config.DefaultFileMod); err == nil {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: file})
		// Klogs (of course) want to print stuff to the screen ;(
		klog.SetOutput(file)
		syscall.Dup2(int(file.Fd()), 2)
	} else {
		panic(err)
	}
}

func main() {
	cmd.Execute()
}
