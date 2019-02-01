package main

import (
	"os"
	"path"

	"github.com/derailed/k9s/cmd"
	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func init() {
	file, err := os.OpenFile(path.Join("/tmp", "k9s.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(file)
}

func main() {
	cmd.Execute()
}
