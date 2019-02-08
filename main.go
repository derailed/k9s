package main

import (
	"os"

	"github.com/derailed/k9s/cmd"
	"github.com/derailed/k9s/config"
	log "github.com/sirupsen/logrus"
)

func init() {
	mod := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	if file, err := os.OpenFile(config.K9sLogs, mod, 0644); err == nil {
		log.SetOutput(file)
	} else {
		panic(err)
	}
}

func main() {
	cmd.Execute()
}
