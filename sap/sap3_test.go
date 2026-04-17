package main

import (
	"math/rand/v2"
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

func TestPageLoadSubroutine(t *testing.T) {
	s := make([]string, 0xFFF)
	s[0x000] = "JMS FF0"
	s[0x001] = "HLT"
	s[0xFF0] = "LDX 1,F8"
	s[0xFF1] = "LDX 2,F9"
	s[0xFF2] = "INP 0"
	s[0xFF3] = "STN 2"
	s[0xFF4] = "INX 2"
	s[0xFF5] = "DSZ 1"
	s[0xFF6] = "JMP FF2"
	s[0xFF7] = "BRB"
	s[0xFF8] = "0x100" // 256
	s[0xFF9] = "0x2"   // ADDRESS

	program, err := assembleSAP3FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP3()
	computer.LoadProgram(program)
	randomTape := make([]uint16, 256)
	for i := 0; i < 256; i++ {
		randomTape[i] = uint16(rand.UintN(16))
	}
	tapeReader := &TapeReader{
		tape: randomTape,
	}
	computer.RegisterInputDevice(tapeReader, 0)

	td := &testDebugger{t: t}
	run(computer, td)

	for i := 0; i < 256; i++ {
		got, want := computer.RAM.mem[i+2], randomTape[i]
		if got != want {
			t.Errorf("got %d want %d", got, want)
		}
	}
}

func TestBootstrapLoader(t *testing.T) {
	s := make([]string, 0xFFF)
	s[0x000] = "JMP FE0"
	s[0xFE0] = "CLA"
	s[0xFE1] = "STA FF9"
	s[0xFE2] = "LDX 3,E9"
	s[0xFE3] = "JMS FF0"
	s[0xFE4] = "XCH 2"   // ADDED WRT THE BOOK
	s[0xFE5] = "STA FF9" // DOESNT WORK WITHOUT IT
	s[0xFE6] = "DSZ 3"
	s[0xFE7] = "JMP FE3"
	s[0xFE8] = "HLT"
	s[0xFE9] = "0x3" // PAGES
	s[0xFF0] = "LDX 1,F8"
	s[0xFF1] = "LDX 2,F9"
	s[0xFF2] = "INP 0"
	s[0xFF3] = "STN 2"
	s[0xFF4] = "INX 2"
	s[0xFF5] = "DSZ 1"
	s[0xFF6] = "JMP FF2"
	s[0xFF7] = "BRB"
	s[0xFF8] = "0x100" // 256
	s[0xFF9] = "0x0"   // ADDRESS

	program, err := assembleSAP3FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP3()
	computer.LoadProgram(program)
	randomTape := make([]uint16, 3*256)
	for i := 0; i < 3*256; i++ {
		randomTape[i] = uint16(rand.UintN(16))
	}
	tapeReader := &TapeReader{
		tape: randomTape,
	}
	computer.RegisterInputDevice(tapeReader, 0)

	td := &testDebugger{t: t}
	run(computer, td)

	for i := 0; i < 3*256; i++ {
		got, want := computer.RAM.mem[i], randomTape[i]
		if got != want {
			t.Errorf("got %d want %d", got, want)
		}
	}
}
