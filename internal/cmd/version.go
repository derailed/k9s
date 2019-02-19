package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Long:  "Prints version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version:%s GitCommit:%s On %s\n", version, commit, date)
		},
	}
}
