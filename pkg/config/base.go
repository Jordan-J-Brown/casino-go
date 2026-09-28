package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

type Base struct {
	Env            Env
	ServiceName    string
	ServiceVersion string
	Nats           Nats
	Tracer         Tracer
	Logging        Logging
}

func Load() (base *Base, err error) {
	defer func() {
		if err != nil {
			err = errors.New("config error: " + err.Error())
		}
	}()

	viper.SetConfigFile("./config.yaml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	base = &Base{}

	env := viper.GetString("env")
	switch Env(env) {
	case Production:
		base.Env = Production
	case Local:
		base.Env = Local
	default:
		return nil, fmt.Errorf(`env is missing or invalid: "%s"`, env)
	}

	base.ServiceName = viper.GetString("service_name")
	if base.ServiceName == "" {
		return nil, errors.New("service_name is missing")
	}

	base.ServiceVersion = viper.GetString("service_version")
	if base.ServiceVersion == "" {
		return nil, errors.New("service_version config is missing")
	}

	if err := base.Nats.load(); err != nil {
		return nil, err
	}
	if err := base.Tracer.load(base.Env); err != nil {
		return nil, err
	}
	if err := base.Logging.load(); err != nil {
		return nil, err
	}

	return base, nil
}
