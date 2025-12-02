package utils

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type FlagKeyInterface interface {
	GetAdd() func(cmd *cobra.Command)
}

type FlagKey[T any] struct {
	Long        string
	Short       string
	Description string
	Add         func(cmd *cobra.Command)
	Retrieve    func(v *viper.Viper) T
}

func (f FlagKey[T]) GetAdd() func(cmd *cobra.Command) {
	return f.Add
}
