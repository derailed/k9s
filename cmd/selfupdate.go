// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func selfUpdate() *cobra.Command {
	command := cobra.Command{
		Use:   "selfupdate",
		Short: "Update k9s binary",
		Long:  "Update k9s binary",
		Run: func(cmd *cobra.Command, args []string) {
			checkAndUpdateApp()
		},
	}

	return &command
}

func checkAndUpdateApp() error {
	latest, err := checkVersion()
	if err != nil {
		return err
	}

	if latest.LessOrEqual(version) {
		log.Info().Msgf("Current version (%s) is the latest", version)
		return nil
	}

	return updateApp(latest)
}

func updateApp(latest *selfupdate.Release) error {
	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}

	return nil
}

func checkVersion() (*selfupdate.Release, error) {
	latest, found, err := selfupdate.DetectLatest(context.Background(), selfupdate.ParseSlug("derailed/k9s"))
	if err != nil {
		return nil, fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}
	return latest, nil
}
