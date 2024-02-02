//go:build ignore

package main

//go:generate bitfield -pkg hw -out ../status.go

type P struct {
	C bool `bitfield:"1"` // negative
	Z bool `bitfield:"1"` // zero
	I bool `bitfield:"1"` // interrupt disable
	D bool `bitfield:"1"` // decimal mode
	B bool `bitfield:"1"` // break
	U bool `bitfield:"1"` // unused
	V bool `bitfield:"1"` // overflow
	N bool `bitfield:"1"` // negative
}
