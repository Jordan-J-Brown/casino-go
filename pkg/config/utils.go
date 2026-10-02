package config

import (
	"reflect"

	"github.com/spf13/viper"
)

func GetOrDefault[T any](key string, def T) T {
	value, ok := viper.Get(key).(T)
	if !ok || reflect.ValueOf(value).IsZero() {
		return def
	}

	return value
}
