package mappers

import (
	"github.com/tinylib/msgp/msgp"

	"nestor/hw/snapshot"
)

func encodeState(num uint16, m msgp.Marshaler) *snapshot.MapperState {
	// TODO: use a scratch buffer and amortize allocations?
	buf, err := m.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}

	return &snapshot.MapperState{
		Num:  num,
		Data: buf,
	}
}

func decodeState(m msgp.Unmarshaler, s *snapshot.MapperState) {
	if _, err := m.UnmarshalMsg(s.Data); err != nil {
		panic(err)
	}
}
