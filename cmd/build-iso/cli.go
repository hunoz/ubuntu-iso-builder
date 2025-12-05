package buildiso

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"

	"github.com/hunoz/ubuntu-iso-builder/builder"
	"github.com/hunoz/ubuntu-iso-builder/cmd/utils"
	generate_cloud_config "github.com/hunoz/ubuntu-iso-builder/generate-cloud-config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v = viper.New()

var BuildIsoCmd = &cobra.Command{
	Use:     "build-iso",
	Aliases: []string{"build"},
	Short:   "Build a Ubuntu ISO",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed(FlagKey.CloudConfigFile.Long) {
			_ = utils.AddFlags(AlternateFlagKeys, cmd)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cloudConfigFilepath := FlagKey.CloudConfigFile.Retrieve(v)
		var typeKey string
		var cloudConfig string
		var outputPath string
		if cmd.Flags().Changed(FlagKey.CloudConfigFile.Long) {
			typeKey = FlagKey.Type.Retrieve(v)
			outputPath = FlagKey.OutputPath.Retrieve(v)
			file, err := os.Open(cloudConfigFilepath)
			if err != nil {
				if os.IsNotExist(err) {
					log.Fatalln("cloud-config file not found")
				} else {
					log.Fatalln("could not open cloud-config file: %w\n", err)
				}
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Fatalf("failed to close file %s: %v", cloudConfigFilepath, err)
				}
			}(file)
			if cfg, err := io.ReadAll(file); err != nil {
				log.Fatalf("failed to read cloud-config file: %v", err)
			} else {
				cloudConfig = string(cfg)
			}
		} else {
			hostname := AlternateFlagKeys.Hostname.Retrieve(v)
			adminUsername := AlternateFlagKeys.AdminUsername.Retrieve(v)
			adminPassword := AlternateFlagKeys.AdminPassword.Retrieve(v)
			rootPassword := AlternateFlagKeys.RootPassword.Retrieve(v)
			sshKeys := AlternateFlagKeys.SSHKey.Retrieve(v)
			typeKey = AlternateFlagKeys.Type.Retrieve(v)
			diskSerial := AlternateFlagKeys.DiskSerial.Retrieve(v)
			plexClaim := AlternateFlagKeys.PlexClaim.Retrieve(v)
			cloudflaredToken := AlternateFlagKeys.CloudflaredToken.Retrieve(v)
			outputPath = AlternateFlagKeys.OutputPath.Retrieve(v)

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

			cloudConfig = conf
		}

		// TODO: Implement the logic here
		isoBuilder := builder.NewISOBuilder(cloudConfig, typeKey, outputPath)
		if err := isoBuilder.Build(); err != nil {
			log.Fatalf("error building iso: %v", err)
		}

		log.Infoln("iso builder finished")
		log.Infof("ISO file located at %s", outputPath)

		return nil
	},
}

func init() {
	err := utils.AddFlags(FlagKey, BuildIsoCmd)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	_ = v.BindPFlags(BuildIsoCmd.Flags())
}
