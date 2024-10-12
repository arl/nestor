// package ines implements a Reader for roms inz the iNES file format, used for
// for the distribution of NES binary programs.
package ines

import (
	"fmt"
	"io"
	"os"
)

type Rom struct {
	header
	Trainer []uint8 // Trainer, 512 bytes if present, or empty.
	PRGROM  []uint8 // PRG is PRG ROM data (size is a multiple of 16k)
	CHRROM  []uint8 // CHR is PRG ROM data (size is a multiple of 8k)
}

func yesno(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func (rom *Rom) PrintInfos(w io.Writer) {
	fmt.Fprintf(w, "iNES2.0 : %s\n", yesno(rom.IsNES20()))
	fmt.Fprintf(w, "PRG ROM: %dx16k\n", rom.header.PRGROMSlots())
	fmt.Fprintf(w, "CHR ROM: %dx8k\n", rom.header.CHRROMSlots())
	fmt.Fprintf(w, "Nametable mirroring: %s\n", rom.Mirroring())
	fmt.Fprintf(w, "Mapper: %d\n", rom.Mapper())
	fmt.Fprintf(w, "Trainer: %s\n", yesno(rom.HasTrainer()))
	fmt.Fprintf(w, "Persistent: %s\n", yesno(rom.HasPersistence()))
}

// ReadRom loads a rom from an iNES file.
func ReadRom(path string) (*Rom, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rom, err := Decode(buf)
	if err != nil {
		return nil, err
	}
	return rom, nil
}

// Decode the give buffer into a rom file.
func Decode(buf []byte) (*Rom, error) {
	rom := new(Rom)

	// header
	var off int
	if err := rom.header.decode(buf); err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}
	off += 16

	// trainer
	if rom.HasTrainer() {
		if len(buf) < off+512 {
			return nil, fmt.Errorf("incomplete TRAINER section")
		}
		rom.Trainer = buf[off : off+512]
		off += 512
	}

	// PRG rom data
	prgRomSize := 0x4000 * rom.prgsz
	if len(buf) < off+prgRomSize {
		return nil, fmt.Errorf("incomplete PRG section")
	}
	rom.PRGROM = buf[off : off+prgRomSize]
	off += prgRomSize

	// CHR rom data
	chrRomSize := 0x2000 * rom.chrsz
	if len(buf) < off+chrRomSize {
		return nil, fmt.Errorf("incomplete CHR section")
	}
	rom.CHRROM = buf[off : off+chrRomSize]
	off += chrRomSize

	return rom, nil
}

const Magic = "NES\x1a"

func (hdr *header) decode(p []byte) error {
	if len(p) < 16 {
		return fmt.Errorf("too small, needs 16 bytes")
	}
	if string(p[:4]) != Magic {
		return fmt.Errorf("invalid magic number")
	}
	copy(hdr.raw[:], p[:16])

	hdr.prgsz = int(hdr.raw[4])
	hdr.chrsz = int(hdr.raw[5])
	return nil
}

type header struct {
	raw   [16]byte
	prgsz int
	chrsz int
}

// PRGROMSlots returns the number of 16kB slots of PRGROM.
func (hdr *header) PRGROMSlots() int {
	return hdr.prgsz
}

// CHRROMSlots returns the number of 8kB slots of CHRROM.
func (hdr *header) CHRROMSlots() int {
	return hdr.chrsz
}

// Mapper returns the mapper number.
func (hdr *header) Mapper() uint16 {
	return uint16(hdr.raw[7]&0xF0) | uint16(hdr.raw[6]>>4)
}

// IsNES20 reports whether the rom is in the NES2.0 format.
func (hdr *header) IsNES20() bool {
	return hdr.raw[7]&0x0c == 0x08
}

type Mirroring int

const (
	HorzMirroring Mirroring = iota
	VertMirroring
)

func (m Mirroring) String() string {
	switch m {
	case HorzMirroring:
		return "horizontal"
	case VertMirroring:
		return "vertical"
	}
	return "unknown"
}

func (hdr *header) Mirroring() Mirroring {
	if hdr.raw[6]&(1<<0) != 0 {
		return VertMirroring
	}
	return HorzMirroring
}

// HasPersistence indicates the presence of persistent saved memory in the rom.
func (hdr *header) HasPersistence() bool {
	return hdr.raw[6]&(1<<1) != 0
}

// Has Trainer indicates the presence of a trainer section in the rom.
func (hdr *header) HasTrainer() bool {
	return hdr.raw[6]&(1<<2) != 0
}
