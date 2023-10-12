// package ines implements a Reader for roms inz the iNES file format, used for
// for the distribution of NES binary programs.
package ines

import (
	"fmt"
	"os"
)

type Rom struct {
	header
	Trainer []byte // Trainer, 512 bytes if present, or empty.
	PRG     []byte // PRG is PRG ROM data (length is multiples of 16k)
	CHR     []byte // CHR is PRG ROM data (length is multiples of 8k)
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
	if len(buf) < off+rom.prgsz {
		return nil, fmt.Errorf("incomplete PRG section")
	}
	rom.PRG = buf[off : off+rom.prgsz]
	off += rom.prgsz

	// CHR rom data
	if len(buf) < off+rom.chrsz {
		return nil, fmt.Errorf("incomplete CHR section")
	}
	rom.CHR = buf[off : off+rom.chrsz]
	off += rom.chrsz

	return rom, nil
}

const Magic = "NES\x1a"

func (hdr *header) decode(p []byte) error {
	if len(p) < 16 {
		return fmt.Errorf("too smaller, needs 16 bytes")
	}
	if string(p[:4]) != Magic {
		return fmt.Errorf("invalid magic number")
	}
	copy(hdr.raw[:], p[:16])

	hdr.prgsz = int(hdr.raw[4]) * 16384
	hdr.chrsz = int(hdr.raw[5]) * 8192
	return nil
}

type header struct {
	raw   [16]byte
	prgsz int
	chrsz int
}

// Has Trainer indicates the presence of a trainer section in the rom.
func (hdr *header) HasTrainer() bool {
	return hdr.raw[6]&0x04 != 0
}

// HasPersistent indicates the presence of persistent memory in the rom.
func (hdr *header) HasPersistent() bool {
	return hdr.raw[6]&0x02 != 0
}

// Mapper returns the mapper number (for now only the lower nibble is used)
func (hdr *header) Mapper() uint8 {
	return hdr.raw[6] >> 4
}
