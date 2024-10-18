package emu

import (
	"io"

	"nestor/hw"
)

type Config struct {
	Input hw.InputConfig `toml:"input"`
	Video VideoConfig    `toml:"video"`

	TraceOut io.WriteCloser `toml:"-"`
}

type VideoConfig struct {
	DisableVSync bool `toml:"disable_vsync"`
}
