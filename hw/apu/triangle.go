package apu

import (
	"nestor/emu/log"
	"nestor/hw/hwio"
)

type TriangleChannel struct {
	Linear hwio.Reg8 `hwio:"offset=0x08,writeonly,wcb"`
	Timer  hwio.Reg8 `hwio:"offset=0x0A,writeonly,wcb"`
	Length hwio.Reg8 `hwio:"offset=0x0B,writeonly,wcb"`
}

func (tc *TriangleChannel) WriteLINEAR(old, val uint8) {
	log.ModSound.InfoZ("write triangle linear").Uint8("val", val).End()
}
func (tc *TriangleChannel) WriteTIMER(old, val uint8) {
	log.ModSound.InfoZ("write triangle timer").Uint8("val", val).End()
}
func (tc *TriangleChannel) WriteLENGTH(old, val uint8) {
	log.ModSound.InfoZ("write triangle length").Uint8("val", val).End()
}
