package main

import "testing"

func TestLoadServerConfigRequiresAdminToken(t *testing.T) {
	_, err := loadServerConfig(func(string) string { return "" })
	if err == nil {
		t.Fatal("loadServerConfig() error = nil, want missing token error")
	}
}

func TestLoadServerConfigRejectsWeakAdminToken(t *testing.T) {
	_, err := loadServerConfig(func(key string) string {
		if key == "INORI_ADMIN_TOKEN" {
			return "short"
		}
		return ""
	})
	if err == nil {
		t.Fatal("loadServerConfig() error = nil, want weak token error")
	}
}

func TestLoadServerConfigAcceptsAdminTokenAndDefaultAddress(t *testing.T) {
	config, err := loadServerConfig(func(key string) string {
		if key == "INORI_ADMIN_TOKEN" {
			return "0123456789abcdef0123456789abcdef"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("loadServerConfig() error = %v", err)
	}
	if config.Address != "127.0.0.1:8080" {
		t.Fatalf("address = %q, want default loopback", config.Address)
	}
	if config.AdminToken == "" || config.InsecureDevAuth {
		t.Fatalf("unexpected config = %+v", config)
	}
}

func TestLoadServerConfigAllowsExplicitInsecureDevelopmentMode(t *testing.T) {
	config, err := loadServerConfig(func(key string) string {
		if key == "INORI_INSECURE_DEV_AUTH" {
			return "1"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("loadServerConfig() error = %v", err)
	}
	if !config.InsecureDevAuth || config.AdminToken != "" {
		t.Fatalf("unexpected config = %+v", config)
	}
}

func TestLoadServerConfigAcceptsAddressOverride(t *testing.T) {
	config, err := loadServerConfig(func(key string) string {
		switch key {
		case "INORI_ADMIN_TOKEN":
			return "0123456789abcdef0123456789abcdef"
		case "INORI_HTTP_ADDR":
			return "127.0.0.1:18080"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("loadServerConfig() error = %v", err)
	}
	if config.Address != "127.0.0.1:18080" {
		t.Fatalf("address = %q, want override", config.Address)
	}
}
