package main

import (
	"flag"
	"os"

	"github.com/derailed/k9s/cmd"
	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
)

func init() {
	config.EnsurePath(config.K9sLogs, config.DefaultDirMod)
}

func init() {
	klog.InitFlags(nil)

	if err := flag.Set("logtostderr", "false"); err != nil {
		panic(err)
	}
	if err := flag.Set("alsologtostderr", "false"); err != nil {
		panic(err)
	}
	if err := flag.Set("stderrthreshold", "fatal"); err != nil {
		panic(err)
	}
	if err := flag.Set("v", "0"); err != nil {
		panic(err)
	}
	if err := flag.Set("log_file", config.K9sLogs); err != nil {
		panic(err)
	}
}

func main() {
	mod := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	file, err := os.OpenFile(config.K9sLogs, mod, config.DefaultFileMod)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		panic(err)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: file})

	cmd.Execute()
}
