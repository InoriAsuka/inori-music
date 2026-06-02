package main

import (
	"testing"
	"time"
)

func TestStorageRefreshInterval(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "unset"},
		{name: "valid", value: "15m", want: 15 * time.Minute},
		{name: "invalid", value: "later"},
		{name: "non positive", value: "0s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("INORI_STORAGE_REFRESH_INTERVAL", tt.value)
			if got := storageRefreshInterval(); got != tt.want {
				t.Fatalf("storageRefreshInterval() = %s, want %s", got, tt.want)
			}
		})
	}
}
