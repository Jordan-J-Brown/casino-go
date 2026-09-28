package config

import (
	"errors"

	"github.com/Jordan-J-Brown/casino-go/pkg/config"
	"github.com/spf13/viper"
)

type Config struct {
	*config.Base
	CasinoWalletID string
	API            API
	Auth           Auth
	Games          Games
	Engines        Engines
}

func Load() (*Config, error) {
	base, err := config.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Base: base,
	}

	cfg.CasinoWalletID = viper.GetString("wallet.id")
	if cfg.CasinoWalletID == "" {
		return nil, errors.New("wallet.id is missing")
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
