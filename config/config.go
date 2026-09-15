package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	AppID             string
	AppPassword        string
	ChannelAuthTenant string
	Port              string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3978"
	}

	return &Config{
		AppID:             os.Getenv("APP_ID"),
		AppPassword:        os.Getenv("APP_PASSWORD"),
		ChannelAuthTenant: os.Getenv("TENANT_ID"),
		Port:              port,
	}
}
