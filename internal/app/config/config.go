// Package config is responsible for parsing command-line flags and environment variables.
//
// It provides structures to handle application configuration, merging values from flags
// and environment variables into a unified configuration.
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"sync"

	"github.com/caarlos0/env/v6"
)

// ConfigStruct holds environment and file config values
type ConfigStruct struct {
	ServerAddress   string `env:"SERVER_ADDRESS" json:"server_address"`       // ServerAddress: The address and port for the server to run on (e.g., ":8080").
	BaseURL         string `env:"BASE_URL" json:"base_url"`                   // BaseURL: The base URL used for constructing short URLs (e.g., http://localhost:8080).
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"` // FileStoragePath: The file path for storing data (e.g., "/tmp/storage").
	DatabaseDsn     string `env:"DATABASE_DSN" json:"database_dsn"`           // DatabaseDsn: The connection string for the database (e.g., "postgres://user:password@localhost/db").
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`           // EnableHTTPS: Is https enable
	ConfigPath      string `env:"CONFIG"`                                     // ConfigPath: config file path
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`                             // TrustedSubnet: config trusted subnet
}

// Vars Options and Config
var (
	// Options contains the configuration values parsed from command-line flags.
	Options struct {
		FlagServiceAddress string // FlagServiceAddress: The address and port for the server to run on (e.g., ":8080").
		FlagBaseURL        string // FlagBaseURL: The base URL used for constructing short URLs (e.g., http://localhost:8080).
		StoragePath        string // StoragePath: The file path for storing data (e.g., "/tmp/storage").
		DatabaseDsn        string // DatabaseDsn: The connection string for the database (e.g., "postgres://user:password@localhost/db").
		EnableHTTPS        bool   // EnableHTTPS: Is https enable
		ConfigPath         string // ConfigPath: config file path
		TrustedSubnet      string // TrustedSubnet: config trusted subnet
	}

	// Config contains the configuration values parsed from environment variables.
	Config ConfigStruct

	// Once makes sure that flags parsing run once
	Once sync.Once
)

// ParseFlags parses configuration values from both command-line flags and environment variables.
//
// Priority is given to command-line flags if both flags and environment variables are set.
//
// Environment variables are loaded using the github.com/caarlos0/env/v6 package,
// and flags are defined and parsed using the `flag` package.
//
// Returns:
//   - error: If environment variables cannot be parsed.
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
		flag.BoolVar(&Options.EnableHTTPS, "s", false, "enable https")
		flag.StringVar(&Options.ConfigPath, "c", "", "config file path")
		flag.StringVar(&Options.TrustedSubnet, "t", "192.168.1.0/24", "config trusted subnet")

	})

	if Config.ConfigPath != "" {
		Options.ConfigPath = Config.ConfigPath
	}
	flag.Parse()

	if Options.ConfigPath != "" {
		cfg, err := loadConfig(Options.ConfigPath)
		if err == nil {
			Config = *cfg

			log.Printf("Config from file: %v", Config)

			return nil
		}
	}

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

	Options.EnableHTTPS = Config.EnableHTTPS

	if Config.TrustedSubnet != "" {
		Options.TrustedSubnet = Config.TrustedSubnet
	}

	flag.Parse()

	return nil
}

// loadConfig from JSON
func loadConfig(path string) (*ConfigStruct, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)

		return nil, err
	}
	defer file.Close()

	var config ConfigStruct
	d := json.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
