// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"fmt"
	"os"

	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "List K9s configurations info",
		RunE:  printInfo,
	}
}

func printInfo(cmd *cobra.Command, args []string) error {
	if err := config.InitLocs(); err != nil {
		return err
	}

	const fmat = "%-27s %s\n"
	printLogo(color.Cyan)
	printTuple(fmat, "Version", version, color.Cyan)
	printTuple(fmat, "Config", config.AppConfigFile, color.Cyan)
	printTuple(fmat, "Custom Views", config.AppViewsFile, color.Cyan)
	printTuple(fmat, "Plugins", config.AppPluginsFile, color.Cyan)
	printTuple(fmat, "Hotkeys", config.AppHotKeysFile, color.Cyan)
	printTuple(fmat, "Aliases", config.AppAliasesFile, color.Cyan)
	printTuple(fmat, "Skins", config.AppSkinsDir, color.Cyan)
	printTuple(fmat, "Context Configs", config.AppContextsDir, color.Cyan)
	printTuple(fmat, "Logs", config.AppLogFile, color.Cyan)
	printTuple(fmat, "Benchmarks", config.AppBenchmarksDir, color.Cyan)
	printTuple(fmat, "ScreenDumps", getScreenDumpDirForInfo(), color.Cyan)

	return nil
}

func printLogo(c color.Paint) {
	for _, l := range ui.LogoSmall {
		fmt.Fprintln(out, color.Colorize(l, c))
	}
	fmt.Fprintln(out)
}

// getScreenDumpDirForInfo get default screen dump config dir or from config.K9sConfigFile configuration.
func getScreenDumpDirForInfo() string {
	if config.AppConfigFile == "" {
		return config.AppDumpsDir
	}

	f, err := os.ReadFile(config.AppConfigFile)
	if err != nil {
		log.Error().Err(err).Msgf("Reads k9s config file %v", err)
		return config.AppDumpsDir
	}

	var cfg config.Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		log.Error().Err(err).Msgf("Unmarshal k9s config %v", err)
		return config.AppDumpsDir
	}
	if cfg.K9s == nil {
		return config.AppDumpsDir
	}

	return cfg.K9s.AppScreenDumpDir()
}
