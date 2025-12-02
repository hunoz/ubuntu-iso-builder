package main

import (
	"github.com/hunoz/ubuntu-iso-builder/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.WithError(err).Fatalln("Error running command")
	}
}
