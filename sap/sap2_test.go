package main

import (
	"math/rand/v2"
	"testing"
)

// TODO: fuzz test?
func TestSAP2Example9_1(t *testing.T) {
	s := []string{"LDA 20", "ADD 21", "ADD 22", "STA 23", "LDB 23", "LDX 23", "HLT"}
	program, err := assembleSAP2FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	a, b, c := uint16(rand.IntN(80)), uint16(rand.IntN(80)), uint16(rand.IntN(80))
	program[0x20], program[0x21], program[0x22] = a, b, c

	computer := NewSAP2()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	sum := computer.RAM.mem[0x23]
	want := a+b+c
	if want != sum {
		t.Errorf("(%d+%d+%d): want %v got %v", a, b, c, want, sum)
	}
}

func TestSAP2Example9_2(t *testing.T) {
	s := []string{"LDA 6", "SUB 7", "JAM 5", "JAZ 5", "JMP 1", "HLT", "25", "9"}
	program, err := assembleSAP2FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP2()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	got := computer.A.Out()
	want := (uint16(0xFFFF) - 1) & 0xFFF // -2 in 12 bits
	if want != got {
		t.Errorf("want %v got %v", want, got)
	}
}

// Example 9-3 relies on built-in square-root and log subroutines

func TestSAP2Example9_4(t *testing.T) {
	s := []string{"LDX 5", "DEX", "JIZ 4", "JMP 1", "HLT", "3"}
	program, err := assembleSAP2FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP2()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	got := computer.X.Out()
	want := uint16(0)
	if want != got {
		t.Errorf("want %v got %v", want, got)
	}
}

func TestSAP2Example9_5(t *testing.T) {
	s := []string{"NOP", "LDX A", "CLA", "DEX", "ADD 9", "JIZ 7", "JMP 3", "OUT", "HLT", "12", "8"}
	program, err := assembleSAP2FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP2()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	got := computer.O.Out()
	want := uint16(12 * 8)
	if want != got {
		t.Errorf("want %v got %v", want, got)
	}
}