package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/color"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version/build info",
		Long:  "Print version/build information",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
}

func printVersion() {
	const secFmt = "%-10s "

	printLogo(color.Cyan)
	printTuple(secFmt, "Version", version)
	printTuple(secFmt, "Commit", commit)
	printTuple(secFmt, "Date", date)
}

func printTuple(format, section, value string) {
	fmt.Printf(color.Colorize(fmt.Sprintf(format, section+":"), color.Cyan))
	fmt.Println(color.Colorize(value, color.White))
}
