package generatecloudinit

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v = viper.New()

var GenerateCloudInitCmd = &cobra.Command{
	Use:     "generate-cloud-init",
	Aliases: []string{"generate"},
	Short:   "Generate cloud-init files",
	RunE: func(cmd *cobra.Command, args []string) error {
		hostname := FlagKey.Hostname.Retrieve(v)
		adminUsername := FlagKey.AdminUsername.Retrieve(v)
		adminPassword := FlagKey.AdminPassword.Retrieve(v)
		rootPassword := FlagKey.RootPassword.Retrieve(v)
		sshKey := FlagKey.SSHKey.Retrieve(v)
		typeKey := FlagKey.Type.Retrieve(v)
		diskSerial := FlagKey.DiskSerial.Retrieve(v)
		plexClaim := FlagKey.PlexClaim.Retrieve(v)
		cloudflaredToken := FlagKey.CloudflaredToken.Retrieve(v)

		return nil
	},
}
