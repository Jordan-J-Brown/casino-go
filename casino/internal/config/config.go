package config

import (
	"github.com/Jordan-J-Brown/casino-go/pkg/config"
)

type Config struct {
	*config.Base
	API     API
	Auth    Auth
	Games   Games
	Engines Engines
}

func Load() (*Config, error) {
	base, err := config.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Base: base,
	}

	if err := cfg.API.load(); err != nil {
		return nil, err
	}
	if err := cfg.Auth.load(); err != nil {
		return nil, err
	}
	if err := cfg.Games.load(); err != nil {
		return nil, err
	}
	if err := cfg.Engines.load(); err != nil {
		return nil, err
	}

	return cfg, nil
}
