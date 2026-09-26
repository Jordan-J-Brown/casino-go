package config

import (
	"errors"

	"github.com/spf13/viper"
)

const (
	keyOtlpEndpoint string = "otel.exporter_otlp_endpoint"
	keyOtlpInsecure string = "otel.exporter_otlp_insecure"
)

type Tracer struct {
	ExporterOtlpEndpoint string
	ExporterOtlpInsecure bool
}

func (cfg *Tracer) load(env Env) error {

	cfg.ExporterOtlpEndpoint = viper.GetString(keyOtlpEndpoint)
	if cfg.ExporterOtlpEndpoint == "" {
		return errors.New(keyOtlpEndpoint + " config is missing")
	}

	if !env.IsProd() {
		cfg.ExporterOtlpInsecure = GetOrDefault(keyOtlpInsecure, false)
	}

	return nil
}
