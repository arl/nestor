package mappers

import (
	"nestor/emu"
	"nestor/emu/log"
)

var modMapper = log.NewModule("mapper")

var All = map[uint16]emu.MapperDesc{
	0: NROM,
}
