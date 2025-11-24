package config

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestDefaultConfigRoundTrip(t *testing.T) {
	var cfg = defaultConfig
	buf, err := toml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to encode default config: %s", err)
	}

	if _, err := toml.Decode(string(buf), &cfg); err != nil {
		t.Fatalf("config round trip failed: %s", err)
	}
}
