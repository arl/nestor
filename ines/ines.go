// package ines implements a Reader for roms in the iNES or NES 2.0 file format,
// used for for the distribution of NES binary programs.
package ines

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Rom struct {
	header
	Trainer []uint8 // Trainer, 512 bytes if present, or empty.
	PRGROM  []uint8 // PRGROM data (size is a multiple of 16k)
	CHRROM  []uint8 // CHRROM data (size is a multiple of 8k)

	Name string
}

func yn(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func (rom *Rom) PrintInfos(w io.Writer) {
	fmt.Fprintf(w, "%s\n", rom.Name)
	fmt.Fprintf(w, "|iNES2.0                | % 14s |\n", yn(rom.IsNES20()))
	if rom.IsNES20() {
		fmt.Fprintf(w, "|Region                 | % 14s |\n", rom.Region())
	}
	fmt.Fprintf(w, "|Mapper                 | % 14d |\n", rom.Mapper())
	if rom.IsNES20() {
		fmt.Fprintf(w, "|Submapper              | % 14d |\n", rom.SubMapper())
	}
	fmt.Fprintf(w, "|PRG ROM                | % 8d x 16k |\n", rom.nslotsPRGROM())
	fmt.Fprintf(w, "|CHR ROM                | % 9d x 8k |\n", rom.nslotsCHRROM())
	if rom.IsNES20() {
		fmt.Fprintf(w, "|PRG RAM                | % 13dk |\n", rom.PRGRAMSize()/1024)
		fmt.Fprintf(w, "|PRG NVRAM              | % 13dk |\n", rom.PRGNVRAMSize()/1024)
		fmt.Fprintf(w, "|CHR RAM                | % 13dk |\n", rom.CHRRAMSize()/1024)
		fmt.Fprintf(w, "|CHR NVRAM              | % 13dk |\n", rom.CHRNVRAMSize()/1024)
	}
	if rom.IsNES20() {
		fmt.Fprintf(w, "|Bus conflicts          | % 14s |\n", yn(rom.HasBusConflicts()))
	}

	fmt.Fprintf(w, "|Nametable mirroring    | % 14s |\n", rom.Mirroring())
	fmt.Fprintf(w, "|Alternative nametable  | % 14s |\n", yn(rom.HasAltNametables()))
	fmt.Fprintf(w, "|Trainer                | % 14s |\n", yn(rom.HasTrainer()))
	fmt.Fprintf(w, "|Persistent             | % 14s |\n", yn(rom.HasPersistence()))
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
	rom.Name = filepath.Base(path)
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
	prgRomSize := 0x4000 * rom.prgromsz
	if len(buf) < off+prgRomSize {
		return nil, fmt.Errorf("incomplete PRG section")
	}
	rom.PRGROM = buf[off : off+prgRomSize]
	off += prgRomSize

	// CHR rom data
	chrRomSize := 0x2000 * rom.chrromsz
	if len(buf) < off+chrRomSize {
		return nil, fmt.Errorf("incomplete CHR section")
	}
	rom.CHRROM = buf[off : off+chrRomSize]
	off += chrRomSize

	return rom, nil
}

const Magic = "NES\x1a"

type header struct {
	raw        [16]byte
	prgromsz   int
	chrromsz   int
	prgramsz   int
	prgnvramsz int
	chrramsz   int
	chrnvramsz int
}

func (hdr *header) decode(p []byte) error {
	if len(p) < 16 {
		return fmt.Errorf("too small, needs 16 bytes")
	}
	if string(p[:4]) != Magic {
		return fmt.Errorf("invalid magic number")
	}
	copy(hdr.raw[:], p[:16])

	hdr.prgromsz = int(hdr.raw[4])
	hdr.chrromsz = int(hdr.raw[5])
	if hdr.IsNES20() {
		hdr.prgromsz |= int(hdr.raw[9]&0x0F) << 8
		hdr.chrromsz |= int(hdr.raw[9] & 0xF0)
		hdr.prgramsz = 64 << int(hdr.raw[10]&0x0F)
		hdr.prgnvramsz = 64 << int(hdr.raw[10]>>4)
		hdr.chrramsz = 64 << int(hdr.raw[11]&0x0F)
		hdr.chrnvramsz = 64 << int(hdr.raw[11]>>4)
	}
	return nil
}

// nslotsPRGROM returns the number of 16kB slots of PRGROM.
func (hdr *header) nslotsPRGROM() int {
	return hdr.prgromsz
}

// nslotsCHRROM returns the number of 8kB slots of CHRROM.
func (hdr *header) nslotsCHRROM() int {
	return hdr.chrromsz
}

// PRGRAMSize returns the size of the PRG-RAM (volatile).
func (hdr *header) PRGRAMSize() int {
	return hdr.prgramsz
}

// PRGNVRAMSize returns the size of the PRG-NVRAM/EEPROM (non-volatile). alias WRAM.
func (hdr *header) PRGNVRAMSize() int {
	return hdr.prgnvramsz
}

// CHRRAMSize returns the size of the CHR-RAM (volatile). alias VRAM.
func (hdr *header) CHRRAMSize() int {
	return hdr.chrramsz
}

// CHRNVRAMSize returns the size of the CHR-NVRAM (non-volatile).
func (hdr *header) CHRNVRAMSize() int {
	return hdr.chrnvramsz
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=Region

// Region indicates the region where the rom was released, for non-homebrew games.
type Region byte

const (
	NTSC Region = iota
	PAL
	Multiple
	Dendy

	Unspecified = 0xFF
)

func (hdr *header) Region() Region {
	if hdr.IsNES20() {
		return Region(hdr.raw[10] & 0x03)
	}
	return Unspecified
}

// Mapper returns the mapper number.
func (hdr *header) Mapper() uint16 {
	base := uint16(hdr.raw[7]&0xF0) | uint16(hdr.raw[6]>>4)
	if hdr.IsNES20() {
		return uint16(hdr.raw[8]&0x0F) | base
	}
	return base
}

// SubMapper returns the submapper number.
func (hdr *header) SubMapper() uint8 {
	if hdr.IsNES20() {
		return hdr.raw[8] >> 4
	}
	return 0
}

// IsNES20 reports whether the rom is in the NES2.0 format.
func (hdr *header) IsNES20() bool {
	return hdr.raw[7]&0x0c == 0x08
}

// NTMirroring describes the layout of the NES 2x2 background nametable
// graphics.
type NTMirroring int

//go:generate go run golang.org/x/tools/cmd/stringer -type=NTMirroring
const (
	HorzMirroring NTMirroring = 1 + iota
	VertMirroring
	OnlyAScreen
	OnlyBScreen
)

func (hdr *header) Mirroring() NTMirroring {
	if hdr.raw[6]&0x01 == 0x01 {
		return VertMirroring
	}
	return HorzMirroring
}

// HasPersistence indicates the presence of persistent saved memory in the rom.
// The original cartridge contained battery-backed PRG RAM ($6000-7FFF) or other
// persistent memory.
func (hdr *header) HasPersistence() bool {
	return hdr.raw[6]&0x02 == 0x02
}

// HasTrainer indicates the presence of a 512 bytes trainer section at
// $7000-$71FF (stored before PRG data).
func (hdr *header) HasTrainer() bool {
	return hdr.raw[6]&0x04 == 0x04
}

// HasAltNametables indicates the presence of a mapper-specific alternative
// nametable layout.
func (hdr *header) HasAltNametables() bool {
	return hdr.raw[6]&0x08 == 0x08
}

// HasBusConflicts reports whether the board is subject to bus conflicts.
func (hdr *header) HasBusConflicts() bool {
	return hdr.raw[6]&0x08 == 0x08
}
