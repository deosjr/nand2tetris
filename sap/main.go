package main

import (
	"fmt"
)

var debug = true

func main() {

	s := []string{"LDX 5", "DEX", "JIZ 4", "JMP 1", "HLT", "3"}
	program, err := assembleSAP2FromStrings(s)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, a := range program[:10] {
		fmt.Printf("%03x\n", a)
	}

	computer := NewSAP2()
	fmt.Println("loading ROM")
	computer.LoadProgram(program)

	var debugger Debugger
	if debug {
		debugger = &standardDebugger{}
	}
	run(computer, debugger)

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
	fmt.Println("X:", c.(*SAP2).X.Out())
}
