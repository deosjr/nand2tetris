package main

import (
	"testing"
)

func TestSAP3Example10_5(t *testing.T) {
	s := []string{"LDX 7,4", "DSZ 7", "JMP 1", "HLT", "0x3"}
	program, err := assembleSAP3FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP3()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	// TODO: test amount of instructions taken?
}

func TestSAP3Page309(t *testing.T) {
	s := []string{"LDM", "0x28", "SBM", "0xF", "STM", "", "HLT"}
	program, err := assembleSAP3FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP3()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	got := computer.RAM.mem[5]
	want := uint16(25)
	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
}
