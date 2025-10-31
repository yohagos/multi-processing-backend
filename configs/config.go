package configs

import (
	"time"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPAddr       string        `env:"HTTP_ADDR" envDefault:":8080"`
	AllowedOrigins []string      `env:"ALLOWED_ORIGINS" envDefault:"*"`
	ReadTimeout    time.Duration `env:"READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout    time.Duration `env:"IDLE_TIMEOUT" envDefault:"15s"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
	return cfg
}
