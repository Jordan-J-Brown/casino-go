package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type Nats struct {
	URL                 string
	ConnectTimeout      time.Duration
	ReconnectWait       time.Duration
	MaxReconnects       int
	DrainTimeout        time.Duration
	PingInterval        time.Duration
	MaxPingsOutstanding int
}

func (n *Nats) load() error {
	natsURL := viper.GetString("nats.url")
	if natsURL == "" {
		return errors.New("nats.url is missing")
	}
	n.URL = natsURL

	n.ConnectTimeout = viper.GetDuration("nats.connect_timeout")
	if n.ConnectTimeout == 0 {
		n.ConnectTimeout = 10 * time.Second
	}

	n.ReconnectWait = viper.GetDuration("nats.reconnect_wait")
	if n.ReconnectWait == 0 {
		n.ReconnectWait = 2 * time.Second
	}
	n.MaxReconnects = viper.GetInt("nats.max_reconnects")
	if n.MaxReconnects == 0 {
		n.MaxReconnects = -1
	}

	n.DrainTimeout = viper.GetDuration("nats.drain_timeout")
	if n.DrainTimeout == 0 {
		n.DrainTimeout = 5 * time.Second
	}

	n.PingInterval = viper.GetDuration("nats.ping_interval")
	if n.PingInterval == 0 {
		n.PingInterval = 20 * time.Second
	}
	n.MaxPingsOutstanding = viper.GetInt("nats.max_pings_outstanding")
	if n.MaxPingsOutstanding == 0 {
		n.MaxPingsOutstanding = 3
	}

	return nil
}
