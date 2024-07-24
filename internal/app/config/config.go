package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

var Options struct {
	FlagServiceAddress string
	FlagBaseURL        string
}

var config struct {
	serverAddress string `env:"SERVER_ADDRESS"`
	baseURL       string `env:"BASE_URL"`
}

func ParseFlags() error {
	err := env.Parse(&config)
	if err != nil {
		return err
	}

	if config.serverAddress != "" {
		Options.FlagServiceAddress = config.serverAddress
	} else {
		flag.StringVar(&Options.FlagServiceAddress, "a", "localhost:8080", "address and port to run server")
	}

	if config.baseURL != "" {
		Options.FlagBaseURL = config.baseURL
	} else {
		flag.StringVar(&Options.FlagBaseURL, "b", "http://localhost:8080", "base result url")
	}

	flag.Parse()
	return nil
}
