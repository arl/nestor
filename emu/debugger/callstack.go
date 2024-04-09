package debugger

import (
	"fmt"
	"slices"
)

type stackFrameFlag uint8

const (
	sffNone stackFrameFlag = iota
	sffNMI
	sffIRQ
)

type stackFrame struct {
	src    uint16
	target uint16
	ret    uint16
	flag   stackFrameFlag
}

type callStack []stackFrame

func (cs *callStack) push(src, dst, ret uint16, flag stackFrameFlag) {
	*cs = append(*cs, stackFrame{
		src:    src,
		target: dst,
		ret:    ret,
		flag:   flag,
	})
}

func (cs *callStack) len() int {
	return len(*cs)
}

func (cs *callStack) pop() {
	if cs.len() == 0 {
		return
	}
	*cs = (*cs)[:cs.len()-1]
}

func (cs *callStack) reset() {
	*cs = (*cs)[:0]
}

type frameInfo [2]string

func (cs *callStack) build(pc uint16) []frameInfo {
	nfos := make([]frameInfo, 0, cs.len()+1)
	var curf *stackFrame
	for i, f := range *cs {
		if i > 0 {
			curf = &((*cs)[i-1])
		}
		src := fmt.Sprintf("$%04X", f.src)
		nfos = slices.Insert(nfos, 0, frameInfo{
			cs.entryPoint(curf),
			src,
		})
	}

	// Current frame
	curf = nil
	if cs.len() > 0 {
		curf = &((*cs)[cs.len()-1])
	}

	return slices.Insert(nfos, 0, frameInfo{
		cs.entryPoint(curf),
		fmt.Sprintf("$%04X", pc),
	})
}

func (callStack) entryPoint(f *stackFrame) string {
	if f == nil {
		return "[bottom of stack]"
	}

	str := fmt.Sprintf("%04X", f.target)
	switch f.flag {
	case sffNMI:
		return "[nmi] $" + str
	case sffIRQ:
		return "[irq] $" + str
	default:
		return str
	}
}
