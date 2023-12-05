package hwio

import (
	"testing"
)

func TestMemReadonly(t *testing.T) {
	var buf Mem
	buf.Data = make([]byte, 0x200)

	for i := 0; i < 8; i++ {
		if buf.Data[i] != 0 {
			t.Errorf("data written at offset %d", i)
		}
	}
}
