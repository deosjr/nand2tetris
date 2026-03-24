package main

import (
	"fmt"
)

var debug = true

func main() {

	//asm := [16]string{"LDA R9", "ADD RA", "OUT", "HLT", "", "", "", "", "", "15", "20", "", "", "", "", ""}
	asm := [16]string{"LDA R9", "ADD RA", "ADD RB", "ADD RC", "SUB RD", "OUT", "HLT", "", "", "1", "2", "3", "4", "5", "", ""}

	program, err := assembleSAP1FromStrings(asm)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, a := range program {
		fmt.Printf("%08b\n", a)
	}

	computer := NewSAP1()
	fmt.Println("loading ROM")
	computer.LoadProgram(program)

	var debugger Debugger
	if debug {
		debugger = &standardDebugger{}
	}
	run(computer, debugger)

}

func run(computer *SAP1, debugger Debugger) {
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
		if computer.Halt {
			break
		}
	}
	fmt.Println("\nticks:", ticks)
}

type Debugger interface {
	BeforeLoop()
	BeforeTick(*SAP1)
	AfterTick(*SAP1)
}

type standardDebugger struct {
	i  int
	sp uint16
}

func (*standardDebugger) BeforeLoop() {
}

func (sd *standardDebugger) BeforeTick(c *SAP1) {
	sd.i++
	fmt.Println("T:", c.Ring.T())
}

func (sd *standardDebugger) AfterTick(c *SAP1) {
	fmt.Println("IR:", c.IR.Out())
	fmt.Println("PC:", c.PC.Out())
	fmt.Println("MAR:", c.MAR.Out())
	fmt.Println("A:", c.A.Out())
}
