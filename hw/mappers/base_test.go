package mappers

import (
	"bytes"
	"testing"
)

func Test_mirrorcopy(t *testing.T) {
	testCases := []struct {
		name string
		dst  []byte
		src  []byte
		want []byte
		n    int
	}{
		{
			name: "empty src",
			dst:  make([]byte, 8),
			src:  []byte{},
			want: make([]byte, 8),
			n:    0,
		},
		{
			name: "empty dst",
			dst:  []byte{},
			src:  []byte{1, 2, 3, 4},
			want: []byte{},
			n:    0,
		},
		{
			name: "same size",
			dst:  make([]byte, 4),
			src:  []byte{1, 2, 3, 4},
			want: []byte{1, 2, 3, 4},
			n:    4,
		},
		{
			name: "src larger than dst",
			dst:  make([]byte, 2),
			src:  []byte{1, 2, 3, 4},
			want: []byte{1, 2},
			n:    2,
		},
		{
			name: "dst larger than src (power of 2 ratio)",
			dst:  make([]byte, 8),
			src:  []byte{1, 2, 3, 4},
			want: []byte{1, 2, 3, 4, 1, 2, 3, 4},
			n:    8,
		},
		{
			name: "dst larger than src (power of 2 ratio, multiple levels)",
			dst:  make([]byte, 16),
			src:  []byte{1, 2, 3, 4},
			want: []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4},
			n:    16,
		},
		{
			name: "single byte source",
			dst:  make([]byte, 8),
			src:  []byte{42},
			want: []byte{42, 42, 42, 42, 42, 42, 42, 42},
			n:    8,
		},
		{
			name: "non-power of 2 pattern",
			dst:  make([]byte, 8),
			src:  []byte{1, 2, 3},
			want: []byte{1, 2, 3, 1, 2, 3, 1, 2},
			n:    8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if n := mirrorcopy(tc.dst, tc.src); n != tc.n {
				t.Errorf("got n = %d, want %d", n, tc.n)
			}
			if !bytes.Equal(tc.dst, tc.want) {
				t.Errorf("got dst = %v, want %v", tc.dst, tc.want)
			}
		})
	}
}
