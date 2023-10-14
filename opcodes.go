package main

import "fmt"

var opCodes = [255]func(cpu *CPU){
	0x4C: JMPabs,
	0x6C: JMPind,
	0x78: SEI,
	0x8D: STAabs,
}

var disasmCodes = [255]func(cpu *CPU) string{
	0x4C: JMPabsDisasm,
	0x6C: JMPindDisasm,
	0x78: SEIDisasm,
	0x8D: STAabsDisasm,
}

// 78
func SEI(cpu *CPU) {
	cpu.P |= pI
	cpu.Clock += 2
	cpu.PC += 1
}

func SEIDisasm(cpu *CPU) string {
	return "SEI"
}

// 4C
func JMPabs(cpu *CPU) {
	dst := cpu.Read16(uint32(cpu.PC + 1))
	cpu.PC = dst
	cpu.Clock += 3
}

func JMPabsDisasm(cpu *CPU) string {
	dst := cpu.Read16(uint32(cpu.PC + 1))
	return fmt.Sprintf("JMP $%04X", dst)
}

// 6C
func JMPind(cpu *CPU) {
	ptr := cpu.Read16(uint32(cpu.PC + 1))
	dst := cpu.Read16(uint32(ptr))
	cpu.PC = dst
	cpu.Clock += 5
}

func JMPindDisasm(cpu *CPU) string {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	return fmt.Sprintf("JMP ($%04X)", oper)
}

// 8D
func STAabs(cpu *CPU) {
	addr := cpu.Read16(uint32(cpu.PC + 1))
	cpu.bus.Write8(uint32(addr), cpu.A)
	cpu.PC += 3
	cpu.Clock += 5
}

func STAabsDisasm(cpu *CPU) string {
	oper := cpu.Read16(uint32(cpu.PC + 1))
	return fmt.Sprintf("STA $%04X", oper)
}
