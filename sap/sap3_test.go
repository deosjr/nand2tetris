package main

import (
	"math/rand/v2"
	"reflect"
	"strings"
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

func TestSAP3Multiplication(t *testing.T) {
	// stores each cycle, seems very wasteful!
	s := []string{"LDX 1,7", "CLA", "ADD 8", "STA 9", "DSZ 1", "JMP 2", "HLT"}
	program, err := assembleSAP3FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	program[0x7] = 35
	program[0x8] = 15

	computer := NewSAP3()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	got := computer.RAM.mem[9]
	want := uint16(35 * 15)
	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
}

func TestSAP3Division(t *testing.T) {
	// terrible, terrible division algorithm.. and book had no HLT either
	s := []string{"CLA", "XCH 1", "LDA B", "SUB C", "INX 1", "JAM 8", "JAZ 8", "JMP 3", "XCH 1", "STA D", "HLT"}
	program, err := assembleSAP3FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	program[0xB] = 5280
	program[0xC] = 17

	computer := NewSAP3()
	computer.LoadProgram(program)

	td := &testDebugger{t: t}
	run(computer, td)

	got := computer.RAM.mem[0xD]
	want := uint16(5280/17) + 1 // +1 because its rounding up...
	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
}

// packName packs a name into uint16 words, 2 chars per word, high byte first.
// Odd-length names pad the low byte of the last word with 0.
func packName(name string) []uint16 {
	words := []uint16{}
	for i := 0; i < len(name); i += 2 {
		w := uint16(name[i]) << 8
		if i+1 < len(name) {
			w |= uint16(name[i+1])
		}
		words = append(words, w)
	}
	return words
}

// Read ASCII input (one char per 2-byte input read)
// and pack it into two chars per 2-byte memory slot
func TestSAP3ReadASCII(t *testing.T) {
	s := []string{
		// read ascii and store subroutine
		// uses X0-3, assumes starting ADDRESS in 0x18
		"CLA",
		"XCH 0", // X0 will contain length of input at the end
		"CLA",
		"XCH 1",         // 0 to start, 0 for store high, 1 for store low bits
		"LDX 2,ADDRESS", // load address in X2
		"DEX 2",         // start at ADDRESS-1 because we incr immediately in the loop
		// start of loop
		"INP 0", // into A
		// TODO: 32 (space) ends as well?
		"JAZ END",
		"INX 0",
		"JIZ 1,PREPHIGH", // jump to prep high bits
		// prepare low bits
		"ORM",   // immediate OR
		"",      // previous high bits stored here
		"DEX 1", // NEXT store high bits
		"JMP STORE",
		// prepare high bits
		"LDX 3,EIGHT", // use idx 3 to count down from 8
		"SHL",         // shift ascii bits up
		"DSZ 3",
		"JMP BACK2", // continue shifting
		"STA TEMP",  // save shifted high bits to fixed address for low path
		"INX 2",     // Go to next ADDRESS
		"INX 1",     // NEXT store low bits
		// store
		"STN 2",
		"JMP START", // jump to start of loop
		"0x8",       // EIGHT
		"0x100",     // ADDRESS
		"HLT",       // END, should be BRB when defined as actual subroutine
	}
	for i, raw := range s {
		for old, new := range map[string]string{
			"START":    "6",
			"PREPHIGH": "E",
			"STORE":    "15",
			"EIGHT":    "17",
			"ADDRESS":  "18",
			"END":      "19",
			"BACK2":    "F",
			"TEMP":     "B",
		} {
			raw = strings.ReplaceAll(raw, old, new)
		}
		s[i] = raw
	}
	program, err := assembleSAP3FromStrings(s)
	if err != nil {
		t.Fatal(err)
	}

	computer := NewSAP3()
	computer.LoadProgram(program)

	var input []uint16
	for _, c := range "Hello" {
		input = append(input, uint16(c))
	}

	tapeReader := &TapeReader{
		tape: input,
	}
	computer.RegisterInputDevice(tapeReader, 0)

	td := &testDebugger{t: t}
	run(computer, td)

	got := computer.RAM.mem[0x100:0x103]
	want := packName("Hello")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %d want %d", got, want)
	}

	gotX := int(computer.X[0].Out())
	wantX := len(input)
	if gotX != wantX {
		t.Errorf("len: got %d want %d", gotX, wantX)
	}
}
