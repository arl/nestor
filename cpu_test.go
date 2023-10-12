package main

import (
	"fmt"
	"testing"
)

func TestPString(t *testing.T) {
	p := P(0x34)
	p.clear()
	fmt.Printf("0b%08b\n", p)
	fmt.Printf("%s\n", p)
}
