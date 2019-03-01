package views

import (
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func run(app *appView, args ...string) bool {
	return app.Suspend(func() {
		if err := execute(args...); err != nil {
			log.Error("Command failed:", err, args)
			app.flash(flashErr, "Doh! command failed", err.Error())
		}
		log.Debug("Command exec sucessfully!")
	})
}

func execute(args ...string) error {
	bin, err := exec.LookPath("kubectl")
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
