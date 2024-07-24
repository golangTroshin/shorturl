package config

import (
	"flag"
	"sync"

	"github.com/caarlos0/env/v6"
)

var (
	Options struct {
		FlagServiceAddress string
		FlagBaseURL        string
		StoragePath        string
	}

	Config struct {
		ServerAddress   string `env:"SERVER_ADDRESS"`
		BaseURL         string `env:"BASE_URL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
	}

	once sync.Once
)

func ParseFlags() error {
	err := env.Parse(&Config)
	if err != nil {
		return err
	}

	once.Do(func() {
		flag.StringVar(&Options.FlagServiceAddress, "a", ":8080", "address and port to run server")
		flag.StringVar(&Options.FlagBaseURL, "b", "http://localhost:8080", "base result url")
		flag.StringVar(&Options.StoragePath, "f", "storage", "storage path")
	})

	if Config.ServerAddress != "" {
		Options.FlagServiceAddress = Config.ServerAddress
	}

	if Config.BaseURL != "" {
		Options.FlagBaseURL = Config.BaseURL
	}

	if Config.FileStoragePath != "" {
		Options.StoragePath = Config.FileStoragePath
	}

	flag.Parse()

	return nil
}
