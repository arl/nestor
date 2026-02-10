package mappers

import (
	"github.com/tinylib/msgp/msgp"

	"github.com/arl/nestor/hw/snapshot"
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

func decodeState[T1 any, T2 interface {
	*T1
	msgp.Unmarshaler
}](s *snapshot.MapperState) T1 {
	var t T1
	if _, err := T2(&t).UnmarshalMsg(s.Data); err != nil {
		panic(err)
	}
	return t
}
