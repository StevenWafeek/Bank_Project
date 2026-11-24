package config

import "testing"

func TestLoadUsesPortForDefaultAddress(t *testing.T) {
	t.Setenv("ADDR", "")
	t.Setenv("PORT", "10000")

	if got := Load().Addr; got != ":10000" {
		t.Fatalf("Load().Addr = %q, want %q", got, ":10000")
	}
}

func TestLoadAddressOverridesPort(t *testing.T) {
	t.Setenv("ADDR", "127.0.0.1:9000")
	t.Setenv("PORT", "10000")

	if got := Load().Addr; got != "127.0.0.1:9000" {
		t.Fatalf("Load().Addr = %q, want %q", got, "127.0.0.1:9000")
	}
}
