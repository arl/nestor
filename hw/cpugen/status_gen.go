//go:build ignore

package main

//go:generate bitfield -pkg hw -out ../status.go

type P struct {
	Carry      bool `bitfield:"1"` // carry
	Zero       bool `bitfield:"1"` // zero
	IntDisable bool `bitfield:"1"` // interrupt disable
	Decimal    bool `bitfield:"1"` // decimal mode
	Break      bool `bitfield:"1"` // break
	Unused     bool `bitfield:"1"` // unused
	Overflow   bool `bitfield:"1"` // overflow
	Negative   bool `bitfield:"1"` // negative
}
