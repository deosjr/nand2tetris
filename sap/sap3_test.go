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
		"INX 1", // NEXT store high bits
		"JMP STORE",
		// prepare high bits
		"LDX 3,EIGHT", // use idx 3 to count down from 8
		"SHL",         // shift ascii bits up
		"DSZ 3",
		"JMP BACK2", // continue shifting
		"STA TEMP",  // save shifted high bits to fixed address for low path
		"INX 2",     // Go to next ADDRESS
		"DEX 1",     // NEXT store low bits
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

// packEntry builds a dictionary entry header: word 0 + packed name.
// Simplest packing first: [LINK 16 bits][FLAGS+LEN 16 bits][name...]
// TODO: Header word 0 layout: [LEN:3][IMM:1][LINK:12], high to low.
// link is the 12-bit address of the previous entry's header (0 = end of chain).
func packEntry(link uint16, name string, imm bool) []uint16 {
	length := uint16(len(name))
	if imm {
		length |= 1 << 15
	}
	return append([]uint16{link, length}, packName(name)...)
	/*
		if len(name) > 7 {
			panic("name exceeds 7 chars")
		}
		var w0 uint16
		w0 |= uint16(len(name)) << 13
		if imm {
			w0 |= 1 << 12
		}
		w0 |= link & 0xFFF
		return append([]uint16{w0}, packName(name)...)
	*/
}

// TestSAP3Find scaffolds a hand-built dictionary and exercises a user-authored
// FIND subroutine. Claude wrote the test harnass, I did the assembly.
//
// FIND contract:
//   - On entry: ARG1 = length of needle in bytes (not words!).
//     The needle follows immediately after in memory.
//   - Walk the chain starting from the sentinel pointed to by the HEAD sysvar
//     at 0xCCC. Sentinels have LEN=0 and must be skipped.
//   - Terminate the walk when the cursor reaches 0 (root).
//   - On exit: write the CFA to OUT1 (0xE01), or 0 if the needle is not
//     found. Write a nonzero value to OUT2 (0xE02) iff the found entry has
//     the IMMEDIATE bit set (0 otherwise, including on miss).
//   - End with BRB. Entry point is at address 0x102.
//
// Replace findAsm below with your implementation.
func TestSAP3Find(t *testing.T) {
	const (
		dictBase   = uint16(0x000)
		headAddr   = uint16(0xCCC)
		arg1Addr   = uint16(0xE00)
		cfaOutAddr = uint16(0xE01)
		immOutAddr = uint16(0xE02)
		lenAddr    = uint16(0xE03)
		brbEncoded = uint16(0xFC00)
	)

	findAsm := []string{
		// FIND:
		"LDA ARG1", // ARG1 = address of needle length (needle comes after)
		"XCH 4",    // STORE that in X4
		"LDA HEAD", // A stores address found in HEAD, end of dict
		// LOOP: (assumes HEAD in A)
		"STM",
		"",              // SCRATCH
		"LDX 5,SCRATCH", // COPY HEAD into X5, SCRATCH will be used to restore later
		"INX 5",         // X5 now pointing at flags+length word for entry

		// check if lengths match
		"LDN 5", // Load entry length word (may have IMM in bit 15)
		"ANM",
		"0x7FFF",    // mask off IMM bit before comparing
		"SBN 4",     // SUBTRACT needle length (address stored in X4)
		"JAZ MATCH", // if zero, we have a match!
		// FOLLOW:
		// continue search
		"LDX 5,SCRATCH",
		"JIZ 5,REL", // if HEAD was 0, give up. REL=FAIL, but page-relative
		"LDN 5",     // follow link
		"JMP LOOP",
		// MATCH:
		// check n/2 (rounded up!!!) name words for equivalence
		"LDA ARG1",
		"XCH 6", // X6 = address of needle length
		"LDN 6", // A = needle length
		"ADM",
		"0x1",
		"SHR",   // A = (A+1) >> 1, i.e. ceil(n/2)
		"XCH 7", // X7 = A
		// INNER:
		"INX 5", // X5++
		"INX 6", // X6++
		"LDN 5",
		"SBN 6",
		"JAZ CONTINUE",
		"JMP FOLLOW",
		// CONTINUE:
		"DSZ 7", // X7-- and if 0, end loop
		"JMP INNER",
		// SUCCESS:
		"INX 5",
		"XCH 5",
		"STA OUT1",
		// extract IMM bit from entry's length word
		"LDX 5,SCRATCH", // X5 = entry addr
		"INX 5",         // X5 = length word
		"LDN 5",         // A = length word
		"ANM",
		"0x8000", // A = 0 or 0x8000
		"STA OUT2",
		"BRB",
		// FAIL:
		"CLA",
		"STA OUT1",
		"CLA",
		"STA OUT2",
		"BRB",
	}

	// Harness at 0x100: call FIND, halt.
	src := make([]string, 4096)
	src[0x0] = "JMP 100"    // Jump over the dictionary
	src[0x100] = "JMS FIND" // call FIND
	src[0x101] = "HLT"
	for i, ln := range findAsm {
		src[0x102+i] = ln
	}

	// Label substitution — expand labels in every source line.
	labels := map[string]string{
		"FIND":     "102",
		"LOOP":     "105",
		"SCRATCH":  "06", // 0x106, in same-page notation
		"FOLLOW":   "10E",
		"MATCH":    "112",
		"INNER":    "119",
		"CONTINUE": "11F",
		"FAIL":     "12B",
		"REL":      "2B",
		"HEAD":     "CCC",
		"ARG1":     "E00",
		"OUT1":     "E01",
		"OUT2":     "E02",
	}
	for i, raw := range src {
		for old, repl := range labels {
			raw = strings.ReplaceAll(raw, old, repl)
		}
		src[i] = raw
	}

	program, err := assembleSAP3FromStrings(src)
	if err != nil {
		t.Fatal(err)
	}
	jump := program[0x0]

	// Build the dictionary: walk a list of entries, chain them via LINK,
	// place each at consecutive addresses starting at dictBase. Each entry
	// gets a 1-word body (BRB) so its CFA is a legal (unused) primitive.
	entries := []struct {
		name string
		imm  bool
	}{
		{"+", false},
		{"DUP", false},
		{"OR", false},
		{"SWAP", true},
	}

	cfaOf := map[string]uint16{}
	prev := jump
	addr := dictBase
	for _, e := range entries {
		hdr := packEntry(prev, e.name, e.imm)
		for i, w := range hdr {
			program[int(addr)+i] = w
		}
		cfa := addr + uint16(len(hdr))
		cfaOf[e.name] = cfa
		program[cfa] = brbEncoded
		prev = addr
		addr = cfa + 1
	}

	// Sentinel: LEN=0, IMM=0, LINK=most-recent-entry.
	sentinel := addr // <-- HEAD will point here
	program[sentinel] = prev & 0xFFF
	program[sentinel+1] = 0x0 // length=0

	// Sysvars.
	program[headAddr] = sentinel
	program[arg1Addr] = lenAddr

	cases := []struct {
		name    string
		needle  string
		wantCFA uint16
		wantIMM bool
	}{
		{"most recent (immediate)", "SWAP", cfaOf["SWAP"], true},
		{"oldest (1-char)", "+", cfaOf["+"], false},
		{"middle odd-length", "DUP", cfaOf["DUP"], false},
		{"middle even-length", "OR", cfaOf["OR"], false},
		{"not found", "BOGUS", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			// Sentinel output values so we can detect if FIND didn't run.
			const sentinelVal = 0xDEAD
			program[cfaOutAddr] = sentinelVal
			program[immOutAddr] = sentinelVal

			computer := NewSAP3()
			computer.LoadProgram(bootloader)
			tapeReader := &TapeReader{tape: program[:]}
			computer.RegisterInputDevice(tapeReader, 0)
			td := &testDebugger{t: t}

			// create the needle in memory, the word we are looking for
			length := len(tc.needle)
			program[lenAddr] = uint16(length)
			for i, v := range packName(tc.needle) {
				program[lenAddr+1+uint16(i)] = v
			}

			run(computer, td)

			gotCFA := computer.RAM.mem[cfaOutAddr]
			gotIMM := computer.RAM.mem[immOutAddr]

			if gotCFA == sentinelVal {
				t.Fatalf("FIND did not write CFA_OUT (still 0x%X) — stub or missing store?", gotCFA)
			}
			if gotIMM == sentinelVal {
				t.Fatalf("FIND did not write IMM_OUT (still 0x%X) — stub or missing store?", gotIMM)
			}
			if gotCFA != tc.wantCFA {
				t.Errorf("CFA: got 0x%X, want 0x%X", gotCFA, tc.wantCFA)
			}
			if (gotIMM != 0) != tc.wantIMM {
				t.Errorf("IMM: got 0x%X (nonzero=%v), want nonzero=%v", gotIMM, gotIMM != 0, tc.wantIMM)
			}
		})
	}
}
