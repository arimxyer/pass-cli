package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	GroupID: "utilities",
	Short:   "Print the version number of pass-cli",
	Long:    `Display version information including build date and commit hash.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pass-cli %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
