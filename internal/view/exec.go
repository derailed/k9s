package view

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

func runK(clear bool, app *App, args ...string) bool {
	bin, err := exec.LookPath("kubectl")
	if err != nil {
		log.Error().Msgf("Unable to find kubectl command in path %v", err)
		return false
	}

	return run(clear, app, bin, false, args...)
}

func run(clear bool, app *App, bin string, bg bool, args ...string) bool {
	app.Halt()
	defer app.Resume()

	return app.Suspend(func() {
		if err := execute(clear, bin, bg, args...); err != nil {
			app.Flash().Errf("Command exited: %v", err)
		}
	})
}

func edit(clear bool, app *App, args ...string) bool {
	bin, err := exec.LookPath(os.Getenv("EDITOR"))
	if err != nil {
		log.Error().Msgf("Unable to find editor command in path %v", err)
		return false
	}

	return run(clear, app, bin, false, args...)
}

func execute(clear bool, bin string, bg bool, args ...string) error {
	if clear {
		clearScreen()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		clearScreen()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Debug().Msg("Command canceled with signal!")
		cancel()
	}()

	log.Debug().Msgf("Running command > %s %s", bin, strings.Join(args, " "))

	cmd := exec.Command(bin, args...)

	var err error
	if bg {
		err = cmd.Start()
	} else {
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		err = cmd.Run()
	}
	log.Debug().Msgf("Command returned error?? %v", err)
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
