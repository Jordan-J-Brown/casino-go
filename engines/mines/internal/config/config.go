package config

import (
	"errors"

	baseconfig "github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/spf13/viper"
)

type Config struct {
	*baseconfig.Base
	EngineID      string
	EngineVersion string
}

func Load() (*Config, error) {
	base, err := baseconfig.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Base: base,
	}

	cfg.EngineID = viper.GetString("engine_id")
	if cfg.EngineID == "" {
		return nil, errors.New("engine_id is missing")
	}

	cfg.EngineVersion = viper.GetString("engine_version")
	if cfg.EngineVersion == "" {
		return nil, errors.New("engine_version is missing")
	}

	return cfg, nil
}
