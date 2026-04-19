package main

import (
	"fmt"
)

var debug = true

func main() {

	computer := NewSAP3()
	computer.LoadProgram(bootloader)

	// Multiplication of 35 and 15 should yield 525
	s := []string{"LDX 1,7", "CLA", "ADD 8", "DSZ 1", "JMP 2", "OUT 8", "HLT", "0x23", "0xF"}
	program, err := assembleSAP3FromStrings(s)
	if err != nil {
		panic(err)
	}

	tapeReader := &TapeReader{tape: program[:]}
	computer.RegisterInputDevice(tapeReader, 0)

	var debugger Debugger
	if debug {
		debugger = &standardDebugger{}
	}
	run(computer, debugger)

	fmt.Println(computer.P[8].Out())
}

func run(computer Computer, debugger Debugger) {
	//computer.SendReset(true)
	//computer.ClockTick()
	//computer.SendReset(false)
	fmt.Println("booting...")

	ticks := 0

	if debugger != nil {
		debugger.BeforeLoop()
	}
	for {
		if debugger != nil {
			debugger.BeforeTick(computer)
		}
		computer.ClockTick()
		ticks++
		if debugger != nil {
			debugger.AfterTick(computer)
		}
		if computer.Halted() {
			break
		}
	}
	fmt.Println("\nticks:", ticks)
}

type Debugger interface {
	BeforeLoop()
	BeforeTick(Computer)
	AfterTick(Computer)
}

type standardDebugger struct {
	i  int
	sp uint16
}

func (*standardDebugger) BeforeLoop() {
}

func (sd *standardDebugger) BeforeTick(c Computer) {
	sd.i++
	//fmt.Println("T:", c.(*SAP2).Ring.T())
}

func (sd *standardDebugger) AfterTick(c Computer) {
	//fmt.Println("IR:", c.IR.Out())
	//fmt.Println("PC:", c.PC.Out())
	//fmt.Println("MAR:", c.MAR.Out())
	//fmt.Println("A:", c.A.Out())
	//fmt.Println("X:", c.(*SAP3).X.Out())
}
