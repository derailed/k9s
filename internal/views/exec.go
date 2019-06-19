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

	"github.com/rs/zerolog/log"
)

func runK(clear bool, app *appView, args ...string) bool {
	bin, err := exec.LookPath("kubectl")
	if err != nil {
		log.Error().Msgf("Unable to find kubeclt command in path %v", err)
		return false
	}

	return run(clear, app, bin, args...)
}

func run(clear bool, app *appView, bin string, args ...string) bool {
	return app.Suspend(func() {
		if err := execute(clear, bin, args...); err != nil {
			app.flash().errf("Command exited: %v", err)
		}
	})
}

func edit(clear bool, app *appView, args ...string) bool {
	bin, err := exec.LookPath(os.Getenv("EDITOR"))
	if err != nil {
		log.Error().Msgf("Unable to find editor command in path %v", err)
		return false
	}

	return run(clear, app, bin, args...)
}

func execute(clear bool, bin string, args ...string) error {
	if clear {
		clearScreen()
	}
	log.Debug().Msgf("Running command > %s %s", bin, strings.Join(args, " "))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Debug().Msg("Command canceled with signal!")
		cancel()
	}()

	cmd := exec.Command(bin, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	log.Debug().Msgf("Command return status %v", err)
	select {
	case <-ctx.Done():
		return errors.New("canceled by operator")
	default:
		return err
	}
}

func clearScreen() {
	log.Debug().Msg("Clearing screen...")
	fmt.Print("\033[H\033[2J")
}
