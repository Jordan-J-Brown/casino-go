package config

import (
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/config"
)

type Auth struct {
	SessionTTL time.Duration
}

func (auth *Auth) load() error {
	auth.SessionTTL = config.GetOrDefault("auth.session_ttl", 30*time.Minute)
	return nil
}
