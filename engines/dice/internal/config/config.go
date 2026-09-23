package config

import baseconfig "github.com/Jordan-J-Brown/casino-go/pkg/config"

type Config struct {
	*baseconfig.Base
}

func Load() (*Config, error) {
	base, err := baseconfig.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Base: base,
	}

	return cfg, nil
}
