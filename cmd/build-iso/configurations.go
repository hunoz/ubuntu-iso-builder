package buildiso

import (
	"os"
	"path/filepath"

	"github.com/hunoz/ubuntu-iso-builder/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var FlagKey = struct {
	CloudConfigFile utils.FlagKey[string]
	Type            utils.FlagKey[string]
	Version         utils.FlagKey[string]
	OutputPath      utils.FlagKey[string]
}{
	CloudConfigFile: utils.FlagKey[string]{
		Long:        "cloud-config-file",
		Short:       "f",
		Description: "The fully rendered cloud-config file",
		Add: func(command *cobra.Command) {
			command.Flags().StringP("cloud-config-file", "f", "", "Path to the cloud-config file")
			_ = command.MarkFlagRequired("cloud-config")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("cloud-config")
		},
	},
	Type: utils.FlagKey[string]{
		Long:        "type",
		Short:       "t",
		Description: "Type of the machine that the machine using the ISO will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("type", "t", "server", "Type of the machine that the machine using the ISO will have [server, desktop]")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("type")
		},
	},
	OutputPath: utils.FlagKey[string]{
		Long:        "output-path",
		Short:       "o",
		Description: "Output path where the ISO will be written to",
		Add: func(cmd *cobra.Command) {
			tmpDir := os.TempDir()
			outputPath := filepath.Join(tmpDir, "ubuntu.iso")
			cmd.Flags().StringP("output-path", "o", outputPath, "Output path where the cloud-config file will be written to")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("output-path")
		},
	},
	Version: utils.FlagKey[string]{
		Long:        "version",
		Short:       "",
		Description: "Version of Ubuntu that will be used. Example: 24.04",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().String("version", "24.04", "Version of Ubuntu that will be used. Example: 24.04")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("version")
		},
	},
}

var AlternateFlagKeys = struct {
	Hostname         utils.FlagKey[string]
	AdminUsername    utils.FlagKey[string]
	AdminPassword    utils.FlagKey[string]
	RootPassword     utils.FlagKey[string]
	SSHKeys          utils.FlagKey[[]string]
	DiskSerial       utils.FlagKey[string]
	PlexClaim        utils.FlagKey[string]
	CloudflaredToken utils.FlagKey[string]
}{
	Hostname: utils.FlagKey[string]{
		Long:        "hostname",
		Short:       "n",
		Description: "Hostname that the machine will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("hostname", "n", "", "Hostname that the machine will have")
			_ = cmd.MarkFlagRequired("hostname")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("hostname")
		},
	},
	AdminUsername: utils.FlagKey[string]{
		Long:        "admin-username",
		Short:       "u",
		Description: "Username of the admin user in the OS. Example: localadmin",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("admin-username", "u", "localadmin", "Username that the machine will have")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("admin-username")
		},
	},
	AdminPassword: utils.FlagKey[string]{
		Long:        "admin-password",
		Short:       "p",
		Description: "Hashed password (e.g. with mkpasswd sha-512) that the admin user will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("admin-password", "p", "password", "Hashed password (e.g. with mkpasswd sha-512) that the admin user will have")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("admin-password")
		},
	},
	RootPassword: utils.FlagKey[string]{
		Long:        "root-password",
		Short:       "r",
		Description: "Password that the root user will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("root-password", "r", "password", "Password that the root user will have")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("root-password")
		},
	},
	SSHKeys: utils.FlagKey[[]string]{
		Long:        "ssh-key",
		Short:       "k",
		Description: "SSH key that the admin user and root will have",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringArrayP("ssh-key", "k", []string{}, "SSH key that the admin user and root will have")
		},
		Retrieve: func(v *viper.Viper) []string {
			return v.GetStringSlice("ssh-key")
		},
	},
	DiskSerial: utils.FlagKey[string]{
		Long:        "disk-serial",
		Short:       "s",
		Description: "Serial of the disk where the OS will be installed",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("disk-serial", "s", "", "Serial of the disk where the OS will be installed")
			_ = cmd.MarkFlagRequired("disk-serial")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("disk-serial")
		},
	},
	PlexClaim: utils.FlagKey[string]{
		Long:        "plex-claim",
		Short:       "c",
		Description: "Plex claim that will be used to activate Plex",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("plex-claim", "c", "", "Plex claim that will be used to activate Plex")
			_ = cmd.MarkFlagRequired("plex-claim")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("plex-claim")
		},
	},
	CloudflaredToken: utils.FlagKey[string]{
		Long:        "cloudflared-token",
		Short:       "d",
		Description: "Cloudflared token that will be used to activate Cloudflared",
		Add: func(cmd *cobra.Command) {
			cmd.Flags().StringP("cloudflared-token", "d", "", "Cloudflared token that will be used to activate Cloudflared")
			_ = cmd.MarkFlagRequired("cloudflared-token")
		},
		Retrieve: func(v *viper.Viper) string {
			return v.GetString("cloudflared-token")
		},
	},
}
