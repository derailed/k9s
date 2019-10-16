package cmd

import (
	"fmt"

	"github.com/derailed/k9s/internal/color"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	var silent bool

	command := cobra.Command{
		Use:   "version",
		Short: "Print version/build info",
		Long:  "Print version/build information",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion(silent)
		},
	}

	command.PersistentFlags().BoolVarP(&silent, "short", "s", false, "Simplified print version for resumed printing")

	return &command
}

func printVersion(silent bool) {
	const secFmt = "%-10s "
	var outputColor color.Paint

	if silent {
		outputColor = color.White
	} else {
		outputColor = color.Cyan
		printLogo(outputColor)
	}
	printTuple(secFmt, "Version", version, outputColor)
	printTuple(secFmt, "Commit", commit, outputColor)
	printTuple(secFmt, "Date", date, outputColor)
}

func printTuple(format, section, value string, outputColor color.Paint) {
	fmt.Printf(color.Colorize(fmt.Sprintf(format, section+":"), outputColor))
	fmt.Println(color.Colorize(value, color.White))
}
