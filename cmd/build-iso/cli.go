package buildiso

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/hunoz/ubuntu-iso-builder/builder"
	"github.com/hunoz/ubuntu-iso-builder/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v = viper.New()

func checkDependencies() error {
	commands := []string{"mkosi", "ukify"}
	for _, command := range commands {
		_, err := exec.LookPath(command)
		if err == nil {
			continue
		}

		return fmt.Errorf("%s not installed", command)
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
			return err
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
			return err
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
