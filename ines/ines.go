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
	PRGROM  []uint8 // PRG is PRG ROM data (size is a multiple of 16k)
	CHRROM  []uint8 // CHR is PRG ROM data (size is a multiple of 8k)

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
		fmt.Fprintf(w, "|PRG RAM                | % 13dk |\n", rom.header.PRGRAMSize()/1024)
		fmt.Fprintf(w, "|PRG NVRAM              | % 13dk |\n", rom.header.PRGNVRAMSize()/1024)
		fmt.Fprintf(w, "|CHR RAM                | % 13dk |\n", rom.header.CHRRAMSize()/1024)
		fmt.Fprintf(w, "|CHR NVRAM              | % 13dk |\n", rom.header.CHRNVRAMSize()/1024)
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

type header struct {
	raw   [16]byte
	prgsz int
	chrsz int
}

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
	if hdr.IsNES20() {
		hdr.prgsz |= int(hdr.raw[9]&0x0F) << 8
		hdr.chrsz |= int(hdr.raw[9] & 0xF0)
	}
	return nil
}

// nslotsPRGROM returns the number of 16kB slots of PRGROM.
func (hdr *header) nslotsPRGROM() int {
	return hdr.prgsz
}

// nslotsCHRROM returns the number of 8kB slots of CHRROM.
func (hdr *header) nslotsCHRROM() int {
	return hdr.chrsz
}

// PRGRAMSize returns the size of the PRG-RAM (volatile).
func (hdr *header) PRGRAMSize() int {
	if hdr.IsNES20() {
		return 64 << int(hdr.raw[10]&0x0F)
	}
	return 0
}

// PRGNVRAMSize returns the size of the PRG-NVRAM/	EEPROM (non-volatile).
func (hdr *header) PRGNVRAMSize() int {
	if hdr.IsNES20() {
		return 64 << int(hdr.raw[10]>>4)
	}
	return 0
}

// CHRRAMSize returns the size of the CHR-RAM (volatile).
func (hdr *header) CHRRAMSize() int {
	if hdr.IsNES20() {
		return 64 << int(hdr.raw[11]&0x0F)
	}
	return 0
}

// CHRNVRAMSize returns the size of the CHR-NVRAM (non-volatile).
func (hdr *header) CHRNVRAMSize() int {
	if hdr.IsNES20() {
		return 64 << int(hdr.raw[11]>>4)
	}
	return 0
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
func (hdr *header) HasPersistence() bool {
	return hdr.raw[6]&0x02 == 0x02
}

// HasTrainer indicates the presence of a trainer section in the rom.
func (hdr *header) HasTrainer() bool {
	return hdr.raw[6]&0x04 == 0x04
}

// HasAltNametables indicates a mapper-specific alternative nametable layout
func (hdr *header) HasAltNametables() bool {
	return hdr.raw[6]&0x08 == 0x08
}
