package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version",
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd := cmd.Root()
		rootCmd.VersionTemplate()
		fmt.Println(fmt.Sprintf("%s version %s", rootCmd.Use, rootCmd.Version))
	},
}
