package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/spf13/cobra"
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
	const sectionFmt = "%-15s "

	printLogo(color.Cyan)
	printTuple(sectionFmt, "Configuration", config.K9sConfigFile, color.Cyan)
	printTuple(sectionFmt, "Logs", config.K9sLogs, color.Cyan)
	printTuple(sectionFmt, "Screen Dumps", config.K9sDumpDir, color.Cyan)
}

func printLogo(c color.Paint) {
	for _, l := range ui.LogoSmall {
		fmt.Println(color.Colorize(l, c))
	}
	fmt.Println()
}
