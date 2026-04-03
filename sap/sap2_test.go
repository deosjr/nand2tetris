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

// Example 9-6 and 9-7 use the interface circuit which is still TODO

func TestSAP2Example9_8(t *testing.T) {
	// This halts accidentally because 4095 is 0xFFF, which is also HLT
	// example moves everything further into mem by prefixing with F
	s := []string{"LDB 9", "AND", "JAZ 6", "LDA A", "JMP 7", "LDA B", "OUT", "BRB", "1", "4095", "0"}
	program, err := assembleSAP2FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP2()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	got := computer.O.Out()
	want := uint16(0x0000)
	if want != got {
		t.Errorf("want %v got %v", want, got)
	}
}

func TestSAP2JMSAndBRB(t *testing.T) {
    // Calls a subroutine at 0x10 that adds mem[0x21] to A, then returns.
    // After the call A should be 5+3=8, verified via OUT.
    s := make([]string, 0x22)
    s[0x00] = "LDA 20" // A = mem[0x20] = 5
    s[0x01] = "JMS 10" // call subroutine at 0x10; return addr = 0x02 stays in PC
    s[0x02] = "OUT"    // O = A (should be 8)
    s[0x03] = "HLT"
    s[0x10] = "ADD 21" // subroutine: A = 5 + 3 = 8
    s[0x11] = "BRB"    // return: clear JMS flag, resume from PC (0x02)
    s[0x20] = "5"      // data
    s[0x21] = "3"      // data

    program, err := assembleSAP2FromStrings(s)
    if err != nil {
        t.Fatal(err)
    }

    computer := NewSAP2()
    computer.LoadProgram(program)

    td := &testDebugger{t: t}
    run(computer, td)

    got := computer.O.Out()
    want := uint16(8)
    if want != got {
        t.Errorf("want %v got %v", want, got)
    }
}