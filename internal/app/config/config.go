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
		DatabaseDsn        string
	}

	Config struct {
		ServerAddress   string `env:"SERVER_ADDRESS"`
		BaseURL         string `env:"BASE_URL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		DatabaseDsn     string `env:"DATABASE_DSN"`
	}

	Once sync.Once
)

func ParseFlags() error {
	err := env.Parse(&Config)
	if err != nil {
		return err
	}

	Once.Do(func() {
		flag.StringVar(&Options.FlagServiceAddress, "a", ":8080", "address and port to run server")
		flag.StringVar(&Options.FlagBaseURL, "b", "http://localhost:8080", "base result url")
		flag.StringVar(&Options.StoragePath, "f", "", "storage path")
		flag.StringVar(&Options.DatabaseDsn, "d", "", "database connection")
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

	if Config.DatabaseDsn != "" {
		Options.DatabaseDsn = Config.DatabaseDsn
	}

	flag.Parse()

	return nil
}
