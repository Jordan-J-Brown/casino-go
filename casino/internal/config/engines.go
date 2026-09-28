package config

import (
	"time"

	"github.com/Jordan-J-Brown/casino-go/pkg/config"
)

type Engines struct {
	ClientTimeoutSeconds time.Duration
}

func (e *Engines) load() error {
	e.ClientTimeoutSeconds = config.GetOrDefault("engines.client_timeout_seconds", 3*time.Second)
	return nil
}
