package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
)

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print configuration info",
		Long:  "Print configuration information",
		Run: func(cmd *cobra.Command, args []string) {
			printInfo()
		},
	}
}

func printInfo() {
	const fmat = "%-25s %s\n"

	printLogo(color.Cyan)
	printTuple(fmat, "Configuration", config.K9sConfigFile, color.Cyan)
	printTuple(fmat, "Logs", config.DefaultLogFile, color.Cyan)
	printTuple(fmat, "Screen Dumps", getScreenDumpDirForInfo(), color.Cyan)
}

func printLogo(c color.Paint) {
	for _, l := range ui.LogoSmall {
		fmt.Fprintln(out, color.Colorize(l, c))
	}
	fmt.Fprintln(out)
}

// getScreenDumpDirForInfo get default screen dump config dir or from config.K9sConfigFile configuration.
func getScreenDumpDirForInfo() string {
	if config.K9sConfigFile == "" {
		return config.K9sDefaultScreenDumpDir
	}

	f, err := os.ReadFile(config.K9sConfigFile)
	if err != nil {
		log.Error().Err(err).Msgf("Reads k9s config file %v", err)
		return config.K9sDefaultScreenDumpDir
	}

	var cfg config.Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		log.Error().Err(err).Msgf("Unmarshal k9s config %v", err)
		return config.K9sDefaultScreenDumpDir
	}
	return cfg.K9s.GetScreenDumpDir()
}
