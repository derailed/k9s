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

const (
	shellCheck = `command -v bash >/dev/null && exec bash || exec sh`
	bannerFmt  = "<<K9s-Shell>> Pod: %s | Container: %s \n"
)

type shellOpts struct {
	clear, background bool
	binary            string
	banner            string
	args              []string
}

func runK(app *App, opts shellOpts) bool {
	bin, err := exec.LookPath("kubectl")
	if err != nil {
		log.Error().Msgf("Unable to find kubectl command in path %v", err)
		return false
	}
	opts.binary, opts.background = bin, false

	return run(app, opts)
}

func run(app *App, opts shellOpts) bool {
	app.Halt()
	defer app.Resume()

	return app.Suspend(func() {
		if err := execute(opts); err != nil {
			app.Flash().Errf("Command exited: %v", err)
		}
	})
}

func edit(app *App, opts shellOpts) bool {
	bin, err := exec.LookPath(os.Getenv("EDITOR"))
	if err != nil {
		log.Error().Msgf("Unable to find editor command in path %v", err)
		return false
	}
	opts.binary, opts.background = bin, false

	return run(app, opts)
}

func execute(opts shellOpts) error {
	if opts.clear {
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

	log.Debug().Msgf("Running command > %s %s", opts.binary, strings.Join(opts.args, " "))

	cmd := exec.Command(opts.binary, opts.args...)

	var err error
	if opts.background {
		err = cmd.Start()
	} else {
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

		_, _ = cmd.Stdout.Write([]byte(opts.banner))
		err = cmd.Run()
	}

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
