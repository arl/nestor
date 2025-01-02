package mappers

import (
	"nestor/emu/log"
	"nestor/hw"
)

var modMapper = log.NewModule("mapper")

var All = map[uint16]hw.MapperDesc{
	0: NROM,
	2: UxROM,
	3: CNROM,
}
