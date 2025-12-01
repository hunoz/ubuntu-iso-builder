package buildiso

import (
	"github.com/hunoz/ubuntu-iso-builder/cmd/utils"
	"github.com/spf13/cobra"
)

var FlagKey = struct {
	CloudConfigFile utils.FlagKey
}{
	CloudConfigFile: utils.FlagKey{
		Add: func(command *cobra.Command) {
			command.Flags().StringP("cloud-config-file", "f", "", "Path to the cloud-config file")
		},
	},
}
