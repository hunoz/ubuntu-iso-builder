package builder

import (
	"fmt"
	"os/exec"
	"strings"
)

type Dependency struct {
	Name         string
	CheckCommand func() bool
	FixMessage   string
}

func commandIsInPath(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

type ISOBuilder struct {
	cloudConfig string
	osType      string
	outputPath  string
}

func (b *ISOBuilder) checkDependencies() error {
	commands := []Dependency{
		{
			Name:         "xorisso",
			CheckCommand: func() bool { return commandIsInPath("xorriso") },
			FixMessage:   "xorisso needs to be installed",
		},
		{
			Name:         "7z",
			CheckCommand: func() bool { return commandIsInPath("7z") },
			FixMessage:   "7z needs to be installed",
		},
	}

	for _, command := range commands {
		if command.CheckCommand() {
			continue
		}

		return fmt.Errorf("%s missing: %s", command.Name, command.FixMessage)
	}
	return nil
}

func (b *ISOBuilder) modifyGrubForAutoinstall(content string) string {
	modified := strings.Replace(content, "---", "autoinstall ds=nocloud;s=/cdrom/nocloud/ ---", -1)

	return strings.Replace(modified, "set timeout=30", "set timeout=5", -1)
}

func (b *ISOBuilder) downloadIso()              {}
func (b *ISOBuilder) extractIso()               {}
func (b *ISOBuilder) createAutoinstallConfigs() {}
func (b *ISOBuilder) modifyGrubConfig()         {}

func (b *ISOBuilder) Build() error {
	if err := b.checkDependencies(); err != nil {
		return err
	}

	return nil
}

func NewISOBuilder(cloudConfig, osType, outputPath string) *ISOBuilder {
	return &ISOBuilder{
		cloudConfig: cloudConfig,
		osType:      osType,
		outputPath:  outputPath,
	}
}
