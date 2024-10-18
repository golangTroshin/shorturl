package config_test

import (
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/golangTroshin/shorturl/internal/app/config"
	"github.com/stretchr/testify/assert"
)

func resetOnce() {
	config.Once = sync.Once{}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		args     []string
		expected struct {
			FlagServiceAddress string
			FlagBaseURL        string
			StoragePath        string
			DatabaseDsn        string
		}
	}{
		{
			name: "Default values",
			env:  map[string]string{},
			args: []string{},
			expected: struct {
				FlagServiceAddress string
				FlagBaseURL        string
				StoragePath        string
				DatabaseDsn        string
			}{
				FlagServiceAddress: ":8080",
				FlagBaseURL:        "http://localhost:8080",
				StoragePath:        "",
				DatabaseDsn:        "",
			},
		},
		{
			name: "Flags override environment variables",
			env: map[string]string{
				"SERVER_ADDRESS":    ":9090",
				"BASE_URL":          "http://example.com",
				"FILE_STORAGE_PATH": "",
				"DATABASE_DSN":      "",
			},
			args: []string{
				"-a", ":7070",
				"-b", "http://override.com",
				"-f", "/custom/storage",
				"-d", "postgres://custom:custom@localhost:5433/customdb?sslmode=disable",
			},
			expected: struct {
				FlagServiceAddress string
				FlagBaseURL        string
				StoragePath        string
				DatabaseDsn        string
			}{
				FlagServiceAddress: ":7070",
				FlagBaseURL:        "http://override.com",
				StoragePath:        "/custom/storage",
				DatabaseDsn:        "postgres://custom:custom@localhost:5433/customdb?sslmode=disable",
			},
		},
		{
			name: "Flags override environment variables",
			env: map[string]string{
				"SERVER_ADDRESS": ":9090",
				"BASE_URL":       "http://example.com",
			},
			args: []string{
				"-a", ":7070",
				"-b", "http://override.com",
			},
			expected: struct {
				FlagServiceAddress string
				FlagBaseURL        string
				StoragePath        string
				DatabaseDsn        string
			}{
				FlagServiceAddress: ":7070",
				FlagBaseURL:        "http://override.com",
				StoragePath:        "",
				DatabaseDsn:        "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetOnce()

			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = append([]string{os.Args[0]}, tt.args...)

			err := config.ParseFlags()
			flag.Parse()
			assert.NoError(t, err)

			assert.Equal(t, tt.expected.FlagServiceAddress, config.Options.FlagServiceAddress, "FlagServiceAddress does not match")
			assert.Equal(t, tt.expected.FlagBaseURL, config.Options.FlagBaseURL, "FlagBaseURL does not match")
			assert.Equal(t, tt.expected.StoragePath, config.Options.StoragePath, "StoragePath does not match")
			assert.Equal(t, tt.expected.DatabaseDsn, config.Options.DatabaseDsn, "DatabaseDsn does not match")
		})
	}
}
