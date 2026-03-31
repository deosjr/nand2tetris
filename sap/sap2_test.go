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
	program[20], program[21], program[22] = a, b, c

	computer := NewSAP2()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	sum := computer.RAM.mem[23]
	want := a+b+c
	if want != sum {
		t.Errorf("(%d+%d+%d): want %v got %v", a, b, c, want, sum)
	}
}