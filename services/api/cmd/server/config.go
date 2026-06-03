package main

import (
	"errors"
	"strings"
)

const minAdminTokenLength = 32

type serverConfig struct {
	Address         string
	AdminToken      string
	InsecureDevAuth bool
}

func loadServerConfig(getenv func(string) string) (serverConfig, error) {
	config := serverConfig{
		Address:         strings.TrimSpace(getenv("INORI_HTTP_ADDR")),
		AdminToken:      strings.TrimSpace(getenv("INORI_ADMIN_TOKEN")),
		InsecureDevAuth: strings.TrimSpace(getenv("INORI_INSECURE_DEV_AUTH")) == "1",
	}
	if config.Address == "" {
		config.Address = "127.0.0.1:8080"
	}

	if config.AdminToken == "" {
		if config.InsecureDevAuth {
			return config, nil
		}
		return serverConfig{}, errors.New("INORI_ADMIN_TOKEN is required unless INORI_INSECURE_DEV_AUTH=1 is set for local development")
	}
	if len(config.AdminToken) < minAdminTokenLength {
		return serverConfig{}, errors.New("INORI_ADMIN_TOKEN must be at least 32 characters")
	}
	return config, nil
}
