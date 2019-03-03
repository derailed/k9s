package views

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func runK(app *appView, args ...string) bool {
	bin, err := exec.LookPath("kubectl")
	if err != nil {
		log.Error("Unable to find kubeclt command in path")
		return false
	}

	return app.Suspend(func() {
		last := len(args) - 1
		if args[last] == "sh" {
			args[last] = "bash"
			if err := execute(bin, args...); err != nil {
				args[last] = "sh"
			} else {
				return
			}
		}
		if err := execute(bin, args...); err != nil {
			log.Errorf("Command exited: %T %v %v", err, err, args)
			app.flash(flashErr, "Command exited:", err.Error())
		}
	})
}

func run1(app *appView, bin string, args ...string) bool {
	return app.Suspend(func() {
		if err := execute(bin, args...); err != nil {
			log.Errorf("Command exited: %T %v %v", err, err, args)
			app.flash(flashErr, "Command exited: ", err.Error())
		}
	})
}

func execute(bin string, args ...string) error {
	clearScreen()
	log.Debugf("Running command > %s %s", bin, strings.Join(args, " "))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Debug("Command canceled with signal!")
		cancel()
	}()

	cmd := exec.Command(bin, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	log.Debug("Command return status ", err)
	select {
	case <-ctx.Done():
		return errors.New("canceled by operator")
	default:
		return err
	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
