package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/views"
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
	printTuple(sectionFmt, "Configuration", config.K9sConfigFile)
	printTuple(sectionFmt, "Logs", config.K9sLogs)
	printTuple(sectionFmt, "Screen Dumps", config.K9sDumpDir)
}

func printLogo(c color.Paint) {
	for _, l := range views.LogoSmall {
		fmt.Println(color.Colorize(l, c))
	}
	fmt.Println()
}
