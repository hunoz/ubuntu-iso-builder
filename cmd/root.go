package cmd

import (
	"os"

	buildiso "github.com/hunoz/ubuntu-iso-builder/cmd/build-iso"
	generatecloudinit "github.com/hunoz/ubuntu-iso-builder/cmd/generate-cloud-config"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var verbose bool

var RootCmd = &cobra.Command{
	Use:     "ubuntu-iso-builder",
	Short:   "This is a CLI to build a custom bootable Ubuntu ISO installer. Requires a Linux OS",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		configureLogging()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Usage()
			os.Exit(1)
		}
	},
}

func configureLogging() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       false,
		DisableColors:          false,
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
	})

	if verbose {
		log.SetLevel(log.DebugLevel)
		log.Debugln("Running in debug mode")
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	commands := []*cobra.Command{
		generatecloudinit.GenerateCloudConfigCmd,
		buildiso.BuildIsoCmd,
		versionCmd,
	}

	for _, command := range commands {
		RootCmd.AddCommand(command)
	}
}
