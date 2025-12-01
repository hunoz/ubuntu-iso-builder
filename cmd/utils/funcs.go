package utils

import (
	"errors"
	"reflect"

	"github.com/spf13/cobra"
)

func AddFlags(obj interface{}, cmd *cobra.Command) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("object must be a struct")
	}

	var flagKeys []FlagKey

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		flagKey, ok := value.Interface().(FlagKey)
		if !ok {
			return errors.New("object must implement FlagKey")
		}

		if flagKey.Add == nil {
			return errors.New("object must implement FlagKey.Add")
		}

		flagKeys = append(flagKeys, flagKey)
	}

	for _, flagKey := range flagKeys {
		flagKey.Add(cmd)
	}

	return nil
}
