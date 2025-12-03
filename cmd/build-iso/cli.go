package buildiso

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/hunoz/ubuntu-iso-builder/builder"
	"github.com/hunoz/ubuntu-iso-builder/cmd/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v = viper.New()

type Dependency struct {
	Name         string
	CheckCommand func() bool
	FixMessage   string
}

func commandIsInPath(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func checkDependencies() error {
	commands := []Dependency{
		{
			Name:         "mkosi",
			CheckCommand: func() bool { return commandIsInPath("mkosi") },
			FixMessage:   "Install mkosi with `apt|yum install mkosi`",
		},
		{
			Name:         "ukify",
			CheckCommand: func() bool { return commandIsInPath("ukify") },
			FixMessage:   "Install ukify with `apt|yum install systemd-ukify`",
		},
		{
			Name:         "systemd-repart",
			CheckCommand: func() bool { return commandIsInPath("systemd-repart") },
			FixMessage:   "Install systemd-repart with `apt|yum install systemd-repart`",
		},
		{
			Name: "systemd-boot",
			CheckCommand: func() bool {
				if _, err := os.Stat("/usr/lib/systemd/boot/efi"); err != nil {
					return false
				}
				return true
			},
			FixMessage: "Install systemd-boot with apt|yum install systemd-boot",
		},
	}
	for _, command := range commands {
		if command.CheckCommand() {
			continue
		}

		return fmt.Errorf("%s not installed: %s", command.Name, command.FixMessage)
	}

	return nil
}

var BuildIsoCmd = &cobra.Command{
	Use:     "build-iso",
	Aliases: []string{"build"},
	Short:   "Build a Ubuntu ISO",
	Long:    "Build a Ubuntu ISO using mkosi (requires a Linux OS)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS != "linux" {
			return fmt.Errorf("OS %s is not supported", runtime.GOOS)
		}

		if err := checkDependencies(); err != nil {
			log.Fatalf(err.Error())
			return nil
		}

		hostname := FlagKeys.Hostname.Retrieve(v)
		adminUsername := FlagKeys.AdminUsername.Retrieve(v)
		adminPassword := FlagKeys.AdminPassword.Retrieve(v)
		rootPassword := FlagKeys.RootPassword.Retrieve(v)
		sshKeys := FlagKeys.SSHKey.Retrieve(v)
		typeKey := FlagKeys.Type.Retrieve(v)
		diskSerial := FlagKeys.DiskSerial.Retrieve(v)
		plexClaim := FlagKeys.PlexClaim.Retrieve(v)
		cloudflaredToken := FlagKeys.CloudflaredToken.Retrieve(v)

		installer, err := builder.Build(&builder.BuildContext{
			Hostname:         hostname,
			AdminUsername:    adminUsername,
			AdminPassword:    adminPassword,
			RootPassword:     rootPassword,
			SSHKeys:          sshKeys,
			Type:             typeKey,
			DiskSerial:       diskSerial,
			PlexClaim:        plexClaim,
			CloudflaredToken: cloudflaredToken,
		})
		if err != nil {
			log.Fatalf("Error building ubuntu-iso: %s", err)
			return nil
		}

		fmt.Printf("Installer image is located at %s", installer)
		fmt.Printf("You can use it with the following command `sudo dd if=%s of=/dev/sdX bs=4M status=progress`", installer)

		return nil
	},
}

func init() {
	err := utils.AddFlags(FlagKeys, BuildIsoCmd)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	_ = v.BindPFlags(BuildIsoCmd.Flags())
}
