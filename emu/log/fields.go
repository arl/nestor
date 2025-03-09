package log

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

type FieldType int

const (
	FieldTypeUnknown FieldType = iota
	FieldTypeBool
	FieldTypeString
	FieldTypeHex8
	FieldTypeHex16
	FieldTypeHex32
	FieldTypeHex64
	FieldTypeInt
	FieldTypeUint
	FieldTypeError
	FieldTypeDuration
	FieldTypeStringer
	FieldTypeBlob
)

type ZField struct {
	Type FieldType
	Key  string

	// Possible values. Only one of these is populated, depedning on Type
	String    string
	Integer   uint64
	Duration  time.Duration
	Error     error
	Interface any
	Boolean   bool
	Blob      []byte
}

func (f *ZField) Value() string {
	switch f.Type {
	case FieldTypeBool:
		if f.Boolean {
			return "true"
		}
		return "false"
	case FieldTypeString:
		return f.String
	case FieldTypeUint:
		return strconv.FormatUint(f.Integer, 10)
	case FieldTypeInt:
		return strconv.FormatInt(int64(f.Integer), 10)
	case FieldTypeHex8:
		return fmt.Sprintf("%02x", uint(f.Integer))
	case FieldTypeHex16:
		return fmt.Sprintf("%04x", uint(f.Integer))
	case FieldTypeHex32:
		return fmt.Sprintf("%08x", uint(f.Integer))
	case FieldTypeHex64:
		return fmt.Sprintf("%016x", uint(f.Integer))
	case FieldTypeError:
		if f.Error == nil {
			return "<nil>"
		}
		return f.Error.Error()
	case FieldTypeDuration:
		return f.Duration.String()
	case FieldTypeStringer:
		return f.Interface.(fmt.Stringer).String()
	case FieldTypeBlob:
		return hex.Dump(f.Blob)
	}
	return ""
}
