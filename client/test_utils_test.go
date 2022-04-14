package client

import (
	"os"
	"testing"
)

func setEnv(t *testing.T, key, value string) {
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set environment variable %s to value %s; err: %v", key, value, err)
	}
}

func unsetEnv(t *testing.T, key string) {
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset environment variable %s; err: %v", key, err)
	}
}
