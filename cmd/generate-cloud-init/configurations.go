package generatecloudinit

import (
	"github.com/hunoz/ubuntu-iso-builder/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var FlagKey = struct {
	Hostname         utils.FlagKey[string]
	AdminUsername    utils.FlagKey[string]
	AdminPassword    utils.FlagKey[string]
	RootPassword     utils.FlagKey[string]
	SSHKey           utils.FlagKey[[]string]
	Type             utils.FlagKey[string]
	DiskSerial       utils.FlagKey[string]
	PlexClaim        utils.FlagKey[string]
	CloudflaredToken utils.FlagKey[string]
}{
	Hostname: utils.FlagKey[string]{
		Long:        "hostname",
		Short:       "n",
		Description: "Hostname that the machine using the ISO will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("hostname", "n", "", "Hostname that the machine using the ISO will have")
			_ = cmd.MarkFlagRequired("hostname")
		},
		Retrieve: func(v viper.Viper) string {
			return v.GetString("hostname")
		},
	},
	AdminUsername: utils.FlagKey[string]{
		Long:        "admin-username",
		Short:       "u",
		Description: "Username that the machine using the ISO will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("admin-username", "u", "localadmin", "Username that the machine using the ISO will have")
		},
		Retrieve: func(v viper.Viper) string {
			return v.GetString("admin-username")
		},
	},
	AdminPassword: utils.FlagKey[string]{
		Long:        "admin-password",
		Short:       "p",
		Description: "Password that the admin user in the OS will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("admin-password", "p", "password", "Password that the admin user in the OS will have")
		},
		Retrieve: func(v viper.Viper) string {
			return v.GetString("admin-password")
		},
	},
	RootPassword: utils.FlagKey[string]{
		Long:        "root-password",
		Short:       "r",
		Description: "Password that the root user in the OS will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("root-password", "r", "password", "Password that the root user in the OS will have")
		},
		Retrieve: func(v viper.Viper) string {
			return v.GetString("root-password")
		},
	},
	SSHKey: utils.FlagKey[[]string]{
		Long:        "ssh-key",
		Short:       "k",
		Description: "SSH key that the users configured in the OS will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringArrayP("ssh-key", "k", []string{}, "SSH key that the users configured in the OS will have")
		},
		Retrieve: func(v viper.Viper) []string {
			return v.GetStringSlice("ssh-key")
		},
	},
	Type: utils.FlagKey[string]{
		Long:        "type",
		Short:       "t",
		Description: "Type of the machine that the machine using the ISO will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("type", "t", "server", "Type of the machine that the machine using the ISO will have [server, desktop]")
		},
		Retrieve: func(v viper.Viper) string {
			return v.GetString("type")
		},
	},
	DiskSerial: utils.FlagKey[string]{
		Long:        "disk-serial",
		Short:       "s",
		Description: "Serial of the disk where the OS will be installed",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("disk-serial", "d", "", "Serial of the disk where the OS will be installed")
		},
		Retrieve: func(v viper.Viper) string {
			return v.GetString("disk-serial")
		},
	},
	PlexClaim: utils.FlagKey[string]{
		Long:        "plex-claim",
		Short:       "c",
		Description: "Plex claim that will be used to activate Plex",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("plex-claim", "c", "", "Plex claim that will be used to activate Plex")
		},
		Retrieve: func(v viper.Viper) string {
			return v.GetString("plex-claim")
		},
	},
	CloudflaredToken: utils.FlagKey[string]{
		Long:        "cloudflared-token",
		Short:       "t",
		Description: "Cloudflared token that will be used to activate Cloudflared",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("cloudflared-token", "t", "", "Cloudflared token that will be used to activate Cloudflared")
		},
		Retrieve: func(v viper.Viper) string {
			return v.GetString("cloudflared-token")
		},
	},
}
