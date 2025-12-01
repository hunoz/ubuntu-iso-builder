package utils

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FlagKey[T any] struct {
	Long        string
	Short       string
	Description string
	Add         AddFlagFunc
	Retrieve    RetrieveValueFunc[T]
}

type RetrieveValueFunc[T any] = func(v *viper.Viper) T

type AddFlagFunc = func(cmd *cobra.Command)
