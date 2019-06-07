package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/printer"
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

	printLogo(printer.Cyan)
	printTuple(secFmt, "Version", version)
	printTuple(secFmt, "Commit", commit)
	printTuple(secFmt, "Date", date)
}

func printTuple(format, section, value string) {
	fmt.Printf(printer.Colorize(fmt.Sprintf(format, section+":"), printer.Cyan))
	fmt.Println(printer.Colorize(value, printer.White))
}
