package config

import (
	"flag"
)

type Config struct {
	Port int
}

func Load() *Config {
	cfg := &Config{}
	flag.IntVar(&cfg.Port, "port", 8080, "Server port")
	flag.Parse()
	return cfg
}
