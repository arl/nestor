//go:build ignore

package main

//go:generate bitfield -pkg hw -out ../status.go

type P struct {
	carry      bool `bitfield:"1"` // carry
	zero       bool `bitfield:"1"` // zero
	intDisable bool `bitfield:"1"` // interrupt disable
	decimal    bool `bitfield:"1"` // decimal mode
	brk        bool `bitfield:"1"` // break
	unused     bool `bitfield:"1"` // unused
	overflow   bool `bitfield:"1"` // overflow
	negative   bool `bitfield:"1"` // negative
}
