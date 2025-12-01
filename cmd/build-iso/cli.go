package buildiso

import (
	"errors"
	"fmt"
	"os"

	"github.com/hunoz/ubuntu-iso-builder/cmd/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var v = viper.New()

var BuildIsoCmd = &cobra.Command{
	Use:     "build-iso",
	Aliases: []string{"build"},
	Short:   "Build a Ubuntu ISO",
	RunE: func(cmd *cobra.Command, args []string) error {
		cloudConfigFilepath := viper.GetString(FlagKey.CloudConfigFile.Long)

		file, err := os.Open(cloudConfigFilepath)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.New("cloud-config file not found")
			} else {
				return fmt.Errorf("could not open cloud-config file: %w", err)
			}
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				panic(err)
			}
		}(file)

		// TODO: Implement the logic here

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
