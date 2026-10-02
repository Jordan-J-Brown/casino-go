package config

import (
	"errors"

	"github.com/spf13/viper"
)

type API struct {
	Port string
}

func (api *API) load() error {
	api.Port = viper.GetString("api.port")
	if api.Port == "" {
		return errors.New("api.port is missing")
	}
	return nil
}
