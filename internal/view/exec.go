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

func runK(a *App, opts shellOpts) bool {
	bin, err := exec.LookPath("kubectl")
	if err != nil {
		log.Error().Msgf("Unable to find kubectl command in path %v", err)
		return false
	}
	var args []string
	if u, err := a.Conn().Config().CurrentUserName(); err == nil {
		args = append(args, "--as", u)
	}
	if g, err := a.Conn().Config().CurrentGroupNames(); err == nil {
		args = append(args, "--as-group", strings.Join(g, ","))
	}
	args = append(args, "--context", a.Config.K9s.CurrentContext)
	if cfg := a.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
		args = append(args, "--kubeconfig", *cfg)
	}
	if len(args) > 0 {
		opts.args = append(opts.args, args...)
	}

	opts.binary, opts.background = bin, false

	return run(a, opts)
}

func run(a *App, opts shellOpts) bool {
	a.Halt()
	defer a.Resume()

	return a.Suspend(func() {
		if err := execute(opts); err != nil {
			a.Flash().Errf("Command exited: %v", err)
		}
	})
}

func edit(a *App, opts shellOpts) bool {
	bin, err := exec.LookPath(os.Getenv("EDITOR"))
	if err != nil {
		log.Error().Msgf("Unable to find editor command in path %v", err)
		return false
	}
	opts.binary, opts.background = bin, false

	return run(a, opts)
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

	log.Debug().Msgf("Running command> %s %s", opts.binary, strings.Join(opts.args, " "))

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
