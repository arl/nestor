package mappers

import (
	"nestor/emu"
	log "nestor/emu/logger"
)

var modMapper = log.NewModule("mapper")

var All = map[uint16]emu.MapperDesc{
	0: NROM,
}
