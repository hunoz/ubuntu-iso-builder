package generatecloudinit

import (
	"log"

	generate_cloud_config "github.com/hunoz/ubuntu-iso-builder/generate-cloud-config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v = viper.New()

var GenerateCloudInitCmd = &cobra.Command{
	Use:     "generate-cloud-config",
	Aliases: []string{"generate"},
	Short:   "Generate cloud-config files",
	RunE: func(cmd *cobra.Command, args []string) error {
		hostname := FlagKeys.Hostname.Retrieve(v)
		adminUsername := FlagKeys.AdminUsername.Retrieve(v)
		adminPassword := FlagKeys.AdminPassword.Retrieve(v)
		rootPassword := FlagKeys.RootPassword.Retrieve(v)
		sshKeys := FlagKeys.SSHKey.Retrieve(v)
		diskSerial := FlagKeys.DiskSerial.Retrieve(v)
		plexClaim := FlagKeys.PlexClaim.Retrieve(v)
		cloudflaredToken := FlagKeys.CloudflaredToken.Retrieve(v)
		outputPath := FlagKeys.OutputPath.Retrieve(v)

		conf, err := generate_cloud_config.GenerateCloudConfig(
			generate_cloud_config.CloudConfigContext{
				Hostname:         hostname,
				AdminUsername:    adminUsername,
				AdminPassword:    adminPassword,
				RootPassword:     rootPassword,
				SSHKeys:          sshKeys,
				DiskSerial:       diskSerial,
				PlexClaim:        plexClaim,
				CloudflaredToken: cloudflaredToken,
			},
		)
		if err != nil {
			log.Fatalf("error generating cloud-config: %v", err)
		}

		if err = generate_cloud_config.WriteCloudConfig(conf, outputPath); err != nil {
			log.Fatalf(err.Error())
		}

		return nil
	},
}
