package main

import (
	"fmt"
	"reflect"
	"testing"
)

// This file contains test harnesses written by Claude, implemented by me.
// In doing so I am learning about Forth by building a Forth on NewSAP3

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

// TestSAP3ForthFind scaffolds a hand-built dictionary and exercises a user-authored
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
func TestSAP3ForthFind(t *testing.T) {
	const (
		dictBase   = uint16(0x000)
		headAddr   = uint16(0xCCC)
		arg1Addr   = uint16(0xE00)
		cfaOutAddr = uint16(0xE01)
		immOutAddr = uint16(0xE02)
		lenAddr    = uint16(0xE03)
		brbEncoded = uint16(0xFC00)
	)

	// Harness + FIND as a single labeled block starting at 0x100.
	// Labels are in-line ("FOO:") and references use {FOO} (absolute) or
	// {:FOO} (same-page offset). No more hand-maintained address map.
	findAsm := []string{
		// Harness.
		"JMS {FIND}",
		"HLT",

		"FIND:",
		"LDA {ARG1}", // ARG1 = address of needle length (needle comes after)
		"XCH 4",      // STORE that in X4
		"LDA {HEAD}", // A stores address found in HEAD, end of dict

		"LOOP:", // (assumes current entry address in A)
		"STM",
		"SCRATCH:", "", // scratch cell; STM writes A here each loop
		"LDX 5,{:SCRATCH}", // COPY entry addr into X5
		"INX 5",            // X5 now points at flags+length word

		// check if lengths match
		"LDN 5",         // Load entry length word (may have IMM in bit 15)
		"ANM", "0x7FFF", // mask off IMM bit before comparing
		"SBN 4",       // SUBTRACT needle length (via X4)
		"JAZ {MATCH}", // if zero, we have a match!

		"FOLLOW:",
		"LDX 5,{:SCRATCH}",
		"JIZ 5,{:FAIL}", // if entry addr was 0, give up
		"LDN 5",         // follow link
		"JMP {LOOP}",

		"MATCH:",
		// check ceil(n/2) name words for equivalence
		"LDA {ARG1}",
		"XCH 6", // X6 = address of needle length
		"LDN 6", // A = needle length
		"ADM", "0x1",
		"SHR",   // A = (A+1) >> 1 = ceil(n/2)
		"XCH 7", // X7 = counter

		"INNER:",
		"INX 5", // X5++
		"INX 6", // X6++
		"LDN 5",
		"SBN 6",
		"JAZ {CONTINUE}",
		"JMP {FOLLOW}",

		"CONTINUE:",
		"DSZ 7", // X7-- and if 0, fall through
		"JMP {INNER}",

		// SUCCESS (fall-through from CONTINUE)
		"INX 5",
		"XCH 5",
		"STA {OUT1}",
		// extract IMM bit from entry's length word
		"LDX 5,{:SCRATCH}", // X5 = entry addr
		"INX 5",            // X5 = length word
		"LDN 5",            // A = length word
		"ANM", "0x8000",    // A = 0 or 0x8000
		"STA {OUT2}",
		"BRB",

		"FAIL:",
		"CLA", "STA {OUT1}",
		"CLA", "STA {OUT2}",
		"BRB",
	}

	externals := map[string]uint16{
		"HEAD": headAddr,
		"ARG1": arg1Addr,
		"OUT1": cfaOutAddr,
		"OUT2": immOutAddr,
	}

	code, _, err := assembleSAP3Labeled(0x100, findAsm, externals)
	if err != nil {
		t.Fatal(err)
	}

	var program [4096]uint16
	// JMP 100 at 0x0 — lost once the dictionary overwrites it, but
	// preserved as the "+"-entry's link so bootloader's JMP 0 still lands
	// us on the harness (see dict builder below).
	jmp, err := encodeASM3("JMP 100")
	if err != nil {
		t.Fatal(err)
	}
	copy(program[0x100:], code)

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
	prev := jmp
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

// TestSAP3ForthWord exercises a user-authored WORD subroutine — the tokenizer
// that reads whitespace-delimited tokens from the token tape and writes them
// to a fixed buffer in the layout FIND expects.
//
// WORD contract:
//   - No inputs via sysvars. Reads token tape via "INP 1". (Port 0 is
//     reserved for the bootloader's program tape.)
//   - Output buffer WORDBUF at 0xE10:
//     0xE10        : length (0 signals EOF; otherwise 1..N, no upper cap).
//     0xE11..      : packed name words, 2 chars per word, high byte first
//     (same scheme as packName).
//   - Delimiter: any char c where c <= 0x20 (space, tab, LF, CR, null).
//     The tape reader returns 0 past end-of-tape, which acts as a delimiter.
//     Custom delimiters (parsed words etc.) are out of scope for now.
//   - Behavior:
//   - Skip leading delimiters BEFORE starting to collect chars.
//   - Consume the terminating delimiter (do NOT leave it for next call).
//   - No length cap — length has its own 16-bit header word in the
//     dictionary entry. If the harness snapshot width ever becomes a
//     constraint, bump snapWords below.
//   - Input assumed uppercase; no case folding.
//   - Register discipline:
//   - Free to clobber A, B, X0..X7.
//   - MUST preserve X8 and X9 — the harness uses them across JMS WORD.
//   - Terminate with BRB.
//
// About the harness (at 0x100):
//
// In real use, WORD is called one-shot: caller invokes JMS WORD, reads
// WORDBUF, does its thing (e.g. hands it to FIND), then calls JMS WORD
// again — which overwrites WORDBUF with the next token. The buffer is
// a single shared slot; previous tokens are gone after the next call.
//
// That's fine for the interpreter but bad for testing: we want to assert
// about the whole tokenization of a multi-token tape in one go. So this
// harness wraps each JMS WORD with "copy WORDBUF into the next snapshot
// slot at SNAPBASE+i*snapWords, then advance i". After HLT the Go side
// reads all snapshot slots and verifies the full sequence.
//
// Everything snapshot-related (snapBase, X8 as the write pointer, the
// COPYLOOP, SNAP_INIT/SNAP_WIDTH data words) is TEST-ONLY scaffolding
// and has no counterpart in the real interpreter — a real caller just
// reads WORDBUF directly after JMS WORD.
//
// The harness bails when WORD returns length=0 (EOF) OR after maxSnaps
// iterations (safety cap — protects the test from runaway loops).
func TestSAP3ForthWord(t *testing.T) {
	const (
		wordBufAddr = uint16(0xE10) // length + name words
		snapBase    = uint16(0xA00) // maxSnaps slots × snapWords words
		maxSnaps    = 8
		snapWords   = 8 // 1 length + up to 7 name words (names up to 14 chars)
	)

	harnessAndWord := []string{
		// --- Harness ---
		// X8 = snapshot write pointer, X9 = safety counter.
		// Both preserved across JMS WORD by contract.
		"LDX 8,{:SNAP_INIT}",
		"LDX 9,{:MAX_ITERS}",

		"LOOP:",
		"JMS {WORD}", // fills WORDBUF (length at WORDBUF, name at WORDBUF+1..)

		// Copy snapWords words from WORDBUF into *X8..X8+snapWords-1.
		"LDX 6,{:WB_INIT}",    // X6 = source cursor (starts at 0xE10)
		"LDX 7,{:SNAP_WIDTH}", // X7 = snapWords (copy-loop counter)

		"COPYLOOP:",
		"LDN 6", // A = *X6
		"STN 8", // *X8 = A
		"INX 6",
		"INX 8",
		"DSZ 7",
		"JMP {COPYLOOP}",

		// Stop on length=0 (EOF).
		"LDA {WORDBUF}",
		"JAZ {DONE}",

		// Safety cap — prevents runaway if WORD never reports EOF.
		"DSZ 9",
		"JMP {LOOP}",

		"DONE:",
		"HLT",

		// Data words (same page, referenced via {:LABEL}).
		// Must match the Go-side consts above.
		"SNAP_INIT:", "0xA00",
		"MAX_ITERS:", "0x8",
		"WB_INIT:", "0xE10",
		"SNAP_WIDTH:", "0x8",

		// --- WORD subroutine — FILL IN ---
		// Reads from INP 1. Free to use X0..X7 and same-page scratch cells.
		// Must NOT touch X8, X9. Must end with BRB.
		"WORD:",
		// copy most of what ReadASCII is doing
		"CLA",
		"XCH 0", // count length in X0
		"CLA",
		"XCH 1",            // high/low store mode
		"LDX 2,{:WB_INIT}", // address to store word
		"LDX 3,{:WB_INIT}", // address to store length
		// consume leading delimiters
		"LEAD:",
		"INP 1",
		"JAZ {END}", // leading 0 is a parse error
		"SBM", "0x21",
		"JAM {LEAD}", // if A - 0x21 is minus, A was 0x20 or less (ignoring 0xFFFF)
		"ADM", "0x21",

		"START:", // because of leading parse, input starts in A already
		"JAZ {END}",
		"SBM", "0x21",
		"JAM {END}", // found a delimiter, end. delimiter is consumed
		"ADM", "0x21",
		"INX 0",
		"JIZ 1,{:PREPHIGH}",

		// PREPLOW:
		"ORM",
		"TEMP:", "",
		"DEX 1",
		"JMP {STORE}",

		"PREPHIGH:",
		"LDX 4,{:EIGHT}",
		"SHIFT:",
		"SHL", // this could be replaced by a single << 8 instr
		"DSZ 4",
		"JMP {SHIFT}",
		"STA {TEMP}",
		"INX 2",
		"INX 1",

		"STORE:",
		"STN 2",
		"INP 1", // read next token here!
		"JMP {START}",
		"EIGHT:", "0x8",

		// add the length value and return
		"END:",
		"XCH 0",
		"STN 3",
		"BRB",
	}

	externals := map[string]uint16{
		"WORDBUF": wordBufAddr,
	}

	code, _, err := assembleSAP3Labeled(0x100, harnessAndWord, externals)
	if err != nil {
		t.Fatal(err)
	}

	// Build the base program: JMP 100 at 0x0 (bootloader lands here),
	// harness+WORD at 0x100+.
	jmp100, err := encodeASM3("JMP 100")
	if err != nil {
		t.Fatal(err)
	}

	type exp struct {
		length uint16
		name   string
	}

	cases := []struct {
		name   string
		tape   string
		expect []exp
	}{
		{
			name:   "single token",
			tape:   "+",
			expect: []exp{{1, "+"}, {0, ""}},
		},
		{
			name:   "leading whitespace",
			tape:   "   DUP",
			expect: []exp{{3, "DUP"}, {0, ""}},
		},
		{
			name:   "multiple tokens",
			tape:   "+ DUP OR SWAP",
			expect: []exp{{1, "+"}, {3, "DUP"}, {2, "OR"}, {4, "SWAP"}, {0, ""}},
		},
		{
			name:   "mixed whitespace",
			tape:   " A\tB\nC ",
			expect: []exp{{1, "A"}, {1, "B"}, {1, "C"}, {0, ""}},
		},
		{
			name:   "empty input",
			tape:   "",
			expect: []exp{{0, ""}},
		},
		{
			name:   "whitespace only",
			tape:   "    ",
			expect: []exp{{0, ""}},
		},
		{
			name:   "seven chars",
			tape:   "LONGEST",
			expect: []exp{{7, "LONGEST"}, {0, ""}},
		},
		{
			// Length is a full 16-bit word in the dict entry; WORD should
			// not truncate. snapWords=8 → snapshots hold names up to 14 chars.
			name:   "long token",
			tape:   "ABCDEFGHIJ XYZ",
			expect: []exp{{10, "ABCDEFGHIJ"}, {3, "XYZ"}, {0, ""}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expect) > maxSnaps {
				t.Fatalf("test case expects %d snapshots, harness only captures %d", len(tc.expect), maxSnaps)
			}

			var program [4096]uint16
			program[0x0] = jmp100
			copy(program[0x100:], code)

			// Sentinel-fill the snapshot region so missed writes are visible.
			const sentinelVal = uint16(0xDEAD)
			for i := uint16(0); i < maxSnaps*snapWords; i++ {
				program[snapBase+i] = sentinelVal
			}
			// Also sentinel the WORDBUF so we can tell a silent BRB from
			// an "intentional length=0" write.
			for i := uint16(0); i < snapWords; i++ {
				program[wordBufAddr+i] = sentinelVal
			}

			// Two separate tapes: port 0 feeds the bootloader with the
			// program image, port 1 feeds WORD with ASCII tokens.
			programTape := &TapeReader{tape: program[:]}

			asciiTape := make([]uint16, len(tc.tape))
			for i, c := range []byte(tc.tape) {
				asciiTape[i] = uint16(c)
			}
			tokenTape := &TapeReader{tape: asciiTape}

			computer := NewSAP3()
			computer.LoadProgram(bootloader)
			computer.RegisterInputDevice(programTape, 0)
			computer.RegisterInputDevice(tokenTape, 1)
			td := &testDebugger{t: t}
			run(computer, td)

			// Verify each expected snapshot.
			for i, e := range tc.expect {
				base := snapBase + uint16(i*snapWords)
				gotLen := computer.RAM.mem[base]

				if gotLen == sentinelVal {
					t.Fatalf("snapshot %d: length still sentinel 0x%X — WORD did not run or did not write length", i, gotLen)
				}
				if gotLen != e.length {
					t.Errorf("snapshot %d: length got %d, want %d", i, gotLen, e.length)
				}

				// Only check name words that the length says are meaningful.
				if e.length > 0 {
					wantName := packName(e.name)
					for j, w := range wantName {
						got := computer.RAM.mem[base+1+uint16(j)]
						if got != w {
							t.Errorf("snapshot %d: name word %d got 0x%04X, want 0x%04X (%q)", i, j, got, w, e.name)
						}
					}
				}
			}

			// Verify the slot AFTER the last expected snapshot was never
			// written (i.e., the harness stopped when it should have).
			if len(tc.expect) < maxSnaps {
				beyond := snapBase + uint16(len(tc.expect)*snapWords)
				if computer.RAM.mem[beyond] != sentinelVal {
					t.Errorf("harness wrote past expected end: snapshot[%d].length = 0x%X, want sentinel 0x%X",
						len(tc.expect), computer.RAM.mem[beyond], sentinelVal)
				}
			}
		})
	}
}

// TestSAP3ForthNumber exercises a user-authored NUMBER subroutine — the
// "is this a literal?" parser the interpreter calls when FIND comes up
// empty on a token.
//
// NUMBER contract:
//   - Input: WORDBUF at 0xE10 (same layout WORD writes — length word
//     followed by packed name words, 2 chars per word, high byte first).
//     Read-only: NUMBER must not modify WORDBUF.
//   - Output:
//     NUMVAL at 0xE30 : parsed value on success (any uint16).
//     NUMOK  at 0xE31 : nonzero on success, 0 on failure. A separate
//     flag is required because NUMVAL=0 is a valid
//     result (the literal "0").
//   - Format (v1): unsigned decimal. Every char must be '0'..'9'.
//     Any non-digit → failure. length > 0 guaranteed by caller.
//     Hex prefix / signs / overflow detection are out of scope for now.
//   - Register discipline: free to clobber A, B, X0..X7.
//   - Terminate with BRB.
//
// Harness shape mirrors FIND: Go pre-fills WORDBUF with the packed test
// input, the assembly just does JMS NUMBER / HLT, we read the sysvars
// after the machine halts.
func TestSAP3ForthNumber(t *testing.T) {
	const (
		wordBufAddr = uint16(0xE10)
		numValAddr  = uint16(0xE30)
		numOkAddr   = uint16(0xE31)
	)

	numberAsm := []string{
		// --- Harness ---
		"JMS {NUMBER}",
		"HLT",

		// Data words (these should be globals probably)
		"WB_INIT:", "0xE10",
		"EIGHT:", "0x8",

		// --- NUMBER subroutine — FILL IN ---
		// Free to clobber A, B, X0..X7. Must end with BRB.
		// Must write NUMVAL and NUMOK on every exit path (success and
		// failure), otherwise the sentinel check below will flag it.
		"NUMBER:",
		"LDX 0,{:WB_INIT}", // X0 contains pointer to WORDBUF
		"CLA", "XCH 1",     // X1 for high/low read mode (quicker than checking length%2)
		"CLA", "XCH 2", // X2 will contain our number
		"LDN 0", "XCH 3", // X3 contains length (assumed nonzero!)

		"LOOP:",
		"JIZ 3,{:SUCCESS}",
		"DEX 3", // use DSZ somehow instead?
		"JIZ 1,{:READHIGH}",

		// READLOW:
		"DEX 1",
		"LDN 0",
		"ANM", "0xFF", // mask off the high bits
		"LOW:",
		// ascii->num(A): do we have a digit?
		// digits: 0-9 in ascii is 0x30 - 0x39
		"SBM", "0x30",
		"JAM {FAIL}",
		"SBM", "0xA",
		"JAM {DIGITFOUND}",
		"JMP {FAIL}",

		"DIGITFOUND:",
		"ADM", "0xA", // we subtracted 0x3A, so + 0xA completes ascii->num conversion
		// X2 = X2 * 10 + ascii->num(A)
		"XCH 2",            // new num now in X2, X2 value in A
		"STM", "TEMP:", "", // TEMP holds old X2
		"SHL",
		"SHL",
		"SHL", // X2 << 3, meaning X2 * 8
		"ADD {TEMP}",
		"ADD {TEMP}", // A = (X2 << 3) + X2 + X2
		"XCH 2",
		"STA {TEMP}",
		"XCH 2",
		"ADD {TEMP}",
		"XCH 2",
		"JMP {LOOP}",

		"READHIGH:",
		"INX 1",
		"LDX 4,{:EIGHT}",
		"INX 0", // advance word pointer
		"LDN 0",
		"SHIFT:",
		"SHR", // >> 8, in a loop
		"DSZ 4",
		"JMP {SHIFT}",
		// here A = *X0 >> 8, low bits have fallen off and high bits are zeroes
		"JMP {LOW}",

		"SUCCESS:",
		"XCH 2",
		"STA {NUMVAL}",
		"XCH 0",
		"STA {NUMOK}",
		"BRB",

		"FAIL:",
		"CLA",
		"STA {NUMVAL}",
		"STA {NUMOK}",
		"BRB",
	}

	externals := map[string]uint16{
		"WORDBUF": wordBufAddr,
		"NUMVAL":  numValAddr,
		"NUMOK":   numOkAddr,
	}

	code, _, err := assembleSAP3Labeled(0x100, numberAsm, externals)
	if err != nil {
		t.Fatal(err)
	}

	jmp100, err := encodeASM3("JMP 100")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name    string
		input   string
		wantVal uint16
		wantOK  bool
	}{
		// Success cases.
		{"zero", "0", 0, true},
		{"single digit", "7", 7, true},
		{"two digits", "42", 42, true},
		{"handling zero", "402", 402, true},
		{"spans two words", "1234", 1234, true}, // exercises unpacking across words
		{"five digits", "12345", 12345, true},
		{"max uint16", "65535", 65535, true},

		// Failure cases.
		{"forth word", "+", 0, false}, // interpreter path: FIND miss → NUMBER must reject
		{"letter", "A", 0, false},
		{"digit then letter", "12A3", 0, false}, // non-digit mid-token
		{"letter then digit", "A12", 0, false},  // non-digit at start
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var program [4096]uint16
			program[0x0] = jmp100
			copy(program[0x100:], code)

			// Sentinel the outputs so "NUMBER never ran" is distinguishable
			// from "NUMBER deliberately wrote 0 on failure".
			const sentinelVal = uint16(0xDEAD)
			program[numValAddr] = sentinelVal
			program[numOkAddr] = sentinelVal

			// Pre-fill WORDBUF with the packed test input.
			program[wordBufAddr] = uint16(len(tc.input))
			for i, w := range packName(tc.input) {
				program[wordBufAddr+1+uint16(i)] = w
			}

			computer := NewSAP3()
			computer.LoadProgram(bootloader)
			tapeReader := &TapeReader{tape: program[:]}
			computer.RegisterInputDevice(tapeReader, 0)
			td := &testDebugger{t: t}
			run(computer, td)

			gotVal := computer.RAM.mem[numValAddr]
			gotOK := computer.RAM.mem[numOkAddr]

			path := "failure"
			if tc.wantOK {
				path = "success"
			}

			if gotVal == sentinelVal {
				t.Fatalf("NUMBER did not write NUMVAL (still 0x%X) — stub or missing store on %s path?", gotVal, path)
			}
			if gotOK == sentinelVal {
				t.Fatalf("NUMBER did not write NUMOK (still 0x%X) — stub or missing store on %s path?", gotOK, path)
			}
			if (gotOK != 0) != tc.wantOK {
				t.Errorf("NUMOK: got 0x%X (success=%v), want success=%v", gotOK, gotOK != 0, tc.wantOK)
			}
			// Only assert NUMVAL on success — failure paths are free to
			// write any value (conventionally 0, but not required).
			if tc.wantOK && gotVal != tc.wantVal {
				t.Errorf("NUMVAL: got %d (0x%X), want %d (0x%X)", gotVal, gotVal, tc.wantVal, tc.wantVal)
			}
		})
	}
}

// TestSAP3ForthInterpreter wires WORD + FIND + NUMBER together into a
// minimal outer loop and runs small Forth programs like "3 5 +".
//
// Interpreter loop (one pass per iteration):
//
//  1. WORD                         — next token into WORDBUF; length=0 → clean HLT
//  2. FIND                         — look it up
//     hit  → EXECUTE (JSN through an X register, see below)
//     miss → fall to NUMBER
//  3. NUMBER ok → push value onto data stack
//  4. NUMBER fail → set ERRFLAG=1, HLT
//
// Data stack: grows up from STACK_BASE (0xD00). SP lives in X8 (chosen
// because WORD's contract guarantees X8 is preserved across calls — the
// SP must survive every WORD / FIND / NUMBER invocation). SP points at
// the next empty slot, so depth = X8 - STACK_BASE and contents live in
// RAM[STACK_BASE : X8].
//
// EXECUTE — calling a CFA held in A: SAP3 has no "JMS indirect through
// register", but JSN is exactly that — it pushes the return address to
// the hardware stack and jumps to an X register's full 12-bit value. So
// the dispatch is just XCH 0 / JSN 0 (swap CFA into X0, indirect-call).
//
// Primitive dictionary: one entry, "+". The body is emitted inline at the
// CFA and pops two stack items, adds, pushes the sum. Enough to exercise
// the whole pipeline on "3 5 +" and small chains.
//
// Why one big assembleSAP3Labeled block rather than four? Measuring the
// sizes (FIND ~46, WORD ~36, NUMBER ~56, outer loop ~20) the whole Forth
// runtime fits comfortably in a single 256-word page with ~100 words of
// headroom for more primitives. Shoving everything onto one page means
// {:LOCAL} page-relative refs work across the entire runtime, and lets
// us dedupe the two shared read-only constants (WB_INIT=0xE10, EIGHT=0x8)
// that both WORD and NUMBER define identically.
//
// Cost: label collisions across the three subroutines have to be resolved
// by renaming. The minimum-viable prefix scheme applied here:
//   - LOOP  (in outer, FIND, NUMBER) → I_LOOP, F_LOOP, N_LOOP
//   - FAIL  (in FIND, NUMBER)        → F_FAIL, N_FAIL
//   - SHIFT (in WORD, NUMBER)        → W_SHIFT, N_SHIFT
//   - TEMP  (in WORD, NUMBER)        → W_TEMP, N_TEMP
//
// TEMP cannot be shared even though the cells look identical: both WORD
// and NUMBER use the "next-word-immediate" self-modifying pattern (ORM
// / STM with the immediate cell as the target of a later STA), so each
// subroutine's TEMP is a distinct cell inside its own code stream.
//
// Why dictBase = 0x000? FIND's end-of-chain check is "current entry
// address is 0" (JIZ on the scratch cell in FOLLOW), which hardwires the
// oldest entry to live at address 0. That conflicts with the bootloader,
// which enters at JMP 0. Resolution: the oldest entry's LINK word is
// FIND's "address of the previous entry" — but FIND stops before ever
// reading it (the JIZ triggers one step earlier). So we're free to stuff
// arbitrary bits in that word, and we stuff "JMP 100" — the bootloader
// lands on the entry's LINK, decodes it as an instruction, and jumps to
// the outer loop. Dict structure stays valid; bootloader trampoline is
// invisible to FIND.
//
// Layout:
//
//	0x000       : dictionary (oldest entry; LINK word = JMP 100 encoding)
//	            : grows up toward 0x100
//	0x100       : Forth runtime (outer loop + WORD + FIND + NUMBER)
//	            : one 256-word page, ~100 words free for future primitives
//	0x200-0xCBF : future code pages (more primitives, DOCOL, COMPILE, ...)
//	0xCCC       : HEAD sysvar
//	0xD00       : data stack (grows up; 256 cells)
//	0xE01       : CFAOUT, IMMOUT         (FIND outputs; 0xE00 free)
//	0xE10       : WORDBUF                (WORD output / FIND needle / NUMBER input)
//	0xE30       : NUMVAL, NUMOK          (NUMBER sysvars)
//	0xE50       : ERRFLAG                (0=ok, 1=unknown token)
func TestSAP3ForthInterpreter(t *testing.T) {
	const (
		wordBufAddr = uint16(0xE10)
		cfaOutAddr  = uint16(0xE01)
		immOutAddr  = uint16(0xE02)
		numValAddr  = uint16(0xE30)
		numOkAddr   = uint16(0xE31)
		headAddr    = uint16(0xCCC)
		stackBase   = uint16(0xD00)
		errFlagAddr = uint16(0xE50)
		dictBase    = uint16(0x000) // FIND requires oldest entry at 0; see docblock
		codeBase    = uint16(0x100)
		spXReg      = 8
	)

	asm := []string{
		// ============================================================
		// Outer loop — entry point at codeBase.
		// ============================================================
		// X8 = stackBase (empty stack).
		"LDX 8,{:I_SP_INIT}",

		"I_LOOP:",
		"JMS {WORD}",
		"LDA {WORDBUF}", // length word
		"JAZ {I_DONE}",  // length=0 → clean exit

		"JMS {FIND}",
		"LDA {CFAOUT}",
		"JAZ {I_TRY_NUMBER}", // miss → try NUMBER

		// EXECUTE: A holds CFA. JSN indirects through an X register.
		"XCH 0", // X0 = CFA, A = old X0 (don't care)
		"JSN 0", // call CFA; returns here via BRB
		"JMP {I_LOOP}",

		"I_TRY_NUMBER:",
		"JMS {NUMBER}",
		"LDA {NUMOK}",
		"JAZ {I_ERROR}", // not a word, not a number
		"LDA {NUMVAL}",
		"STN 8", // *SP = value
		"INX 8", // SP++
		"JMP {I_LOOP}",

		"I_DONE:",
		"CLA", "STA {ERRFLAG}", // success: ERRFLAG = 0
		"HLT",

		"I_ERROR:",
		"LDM", "0x1",
		"STA {ERRFLAG}", // failure: ERRFLAG = 1
		"HLT",

		"I_SP_INIT:", fmt.Sprintf("0x%X", stackBase),

		// ============================================================
		// WORD — reads INP 1, writes WORDBUF. Preserves X8.
		// Lifted verbatim from TestSAP3ForthWord; renames: all internal
		// labels get W_ prefix. WB_INIT and EIGHT dedup to shared block.
		// ============================================================
		"WORD:",
		"CLA",
		"XCH 0", // count length in X0
		"CLA",
		"XCH 1",            // high/low store mode
		"LDX 2,{:WB_INIT}", // address to store word
		"LDX 3,{:WB_INIT}", // address to store length
		// consume leading delimiters
		"W_LEAD:",
		"INP 1",
		"JAZ {W_END}", // leading 0 is a parse error
		"SBM", "0x21",
		"JAM {W_LEAD}",
		"ADM", "0x21",

		"W_START:",
		"JAZ {W_END}",
		"SBM", "0x21",
		"JAM {W_END}", // delimiter ends token; consumed
		"ADM", "0x21",
		"INX 0",
		"JIZ 1,{:W_PREPHIGH}",

		// W_PREPLOW:
		"ORM",
		"W_TEMP:", "",
		"DEX 1",
		"JMP {W_STORE}",

		"W_PREPHIGH:",
		"LDX 4,{:EIGHT}",
		"W_SHIFT:",
		"SHL",
		"DSZ 4",
		"JMP {W_SHIFT}",
		"STA {W_TEMP}",
		"INX 2",
		"INX 1",

		"W_STORE:",
		"STN 2",
		"INP 1", // read next char
		"JMP {W_START}",

		"W_END:",
		"XCH 0",
		"STN 3",
		"BRB",

		// ============================================================
		// FIND — reads WORDBUF+HEAD, writes CFAOUT+IMMOUT. Preserves X8.
		// Lifted from TestSAP3ForthFind; all internal labels get F_ prefix.
		// ARG1 indirection dropped — FIND reads WORDBUF directly via WB_INIT.
		// ============================================================
		"FIND:",
		"LDX 4,{:WB_INIT}", // X4 = WORDBUF address (needle length at *X4)
		"LDA {HEAD}",

		"F_LOOP:", // (assumes current entry address in A)
		"STM",
		"F_SCRATCH:", "", // scratch cell; STM writes A here each loop
		"LDX 5,{:F_SCRATCH}",
		"INX 5", // X5 now points at flags+length word

		"LDN 5",
		"ANM", "0x7FFF", // mask off IMM bit
		"SBN 4",
		"JAZ {F_MATCH}",

		"F_FOLLOW:",
		"LDX 5,{:F_SCRATCH}",
		"JIZ 5,{:F_FAIL}", // if entry addr was 0, give up
		"LDN 5",           // follow link
		"JMP {F_LOOP}",

		"F_MATCH:",
		"LDX 6,{:WB_INIT}", // X6 = WORDBUF address (reload for name comparison)
		"LDN 6",
		"ADM", "0x1",
		"SHR",
		"XCH 7", // X7 = ceil(n/2) counter

		"F_INNER:",
		"INX 5",
		"INX 6",
		"LDN 5",
		"SBN 6",
		"JAZ {F_CONTINUE}",
		"JMP {F_FOLLOW}",

		"F_CONTINUE:",
		"DSZ 7",
		"JMP {F_INNER}",

		// F_SUCCESS (fall-through).
		"INX 5",
		"XCH 5",
		"STA {CFAOUT}",
		"LDX 5,{:F_SCRATCH}",
		"INX 5",
		"LDN 5",
		"ANM", "0x8000",
		"STA {IMMOUT}",
		"BRB",

		"F_FAIL:",
		"CLA", "STA {CFAOUT}",
		"CLA", "STA {IMMOUT}",
		"BRB",

		// ============================================================
		// NUMBER — reads WORDBUF, writes NUMVAL+NUMOK. Preserves X8.
		// Lifted verbatim from TestSAP3ForthNumber; all internal labels
		// get N_ prefix.
		// ============================================================
		"NUMBER:",
		"LDX 0,{:WB_INIT}", // pointer to WORDBUF
		"CLA", "XCH 1",     // high/low read mode
		"CLA", "XCH 2", // running number
		"LDN 0", "XCH 3", // length (assumed nonzero)

		"N_LOOP:",
		"JIZ 3,{:N_SUCCESS}",
		"DEX 3",
		"JIZ 1,{:N_READHIGH}",

		// N_READLOW:
		"DEX 1",
		"LDN 0",
		"ANM", "0xFF",
		"N_LOW:",
		"SBM", "0x30",
		"JAM {N_FAIL}",
		"SBM", "0xA",
		"JAM {N_DIGITFOUND}",
		"JMP {N_FAIL}",

		"N_DIGITFOUND:",
		"ADM", "0xA", // complete ascii→num conversion
		"XCH 2",
		"STM", "N_TEMP:", "", // TEMP holds old X2
		"SHL", "SHL", "SHL", // X2 << 3
		"ADD {N_TEMP}",
		"ADD {N_TEMP}", // A = (X2 << 3) + X2 + X2 = X2 * 10
		"XCH 2",
		"STA {N_TEMP}",
		"XCH 2",
		"ADD {N_TEMP}",
		"XCH 2",
		"JMP {N_LOOP}",

		"N_READHIGH:",
		"INX 1",
		"LDX 4,{:EIGHT}",
		"INX 0",
		"LDN 0",
		"N_SHIFT:",
		"SHR",
		"DSZ 4",
		"JMP {N_SHIFT}",
		"JMP {N_LOW}",

		"N_SUCCESS:",
		"XCH 2",
		"STA {NUMVAL}",
		"XCH 0",
		"STA {NUMOK}",
		"BRB",

		"N_FAIL:",
		"CLA",
		"STA {NUMVAL}",
		"STA {NUMOK}",
		"BRB",

		// ============================================================
		// Shared read-only constants — used by both WORD and NUMBER.
		// ============================================================
		"WB_INIT:", fmt.Sprintf("0x%X", wordBufAddr),
		"EIGHT:", "0x8",
	}

	externals := map[string]uint16{
		"WORDBUF": wordBufAddr,
		"CFAOUT":  cfaOutAddr,
		"IMMOUT":  immOutAddr,
		"NUMVAL":  numValAddr,
		"NUMOK":   numOkAddr,
		"HEAD":    headAddr,
		"ERRFLAG": errFlagAddr,
	}

	code, _, err := assembleSAP3Labeled(codeBase, asm, externals)
	if err != nil {
		t.Fatalf("assembling Forth runtime: %v", err)
	}

	// Everything must fit on one 256-word page so {:LOCAL} refs resolve.
	// (The assembler already rejects cross-page {:LOCAL} — this check
	// just gives a nicer error than the raw assembler message.)
	if len(code) > 256 {
		t.Fatalf("Forth runtime is %d words, exceeds 256-word page", len(code))
	}
	t.Logf("Forth runtime: %d words (%d free on page)", len(code), 256-len(code))

	jmp100, err := encodeASM3("JMP 100")
	if err != nil {
		t.Fatal(err)
	}

	// "+" primitive body — placed inline in the dictionary as the CFA.
	// Pop two, add, push sum. Uses X8 (the SP) directly.
	plusBody := []string{
		"DEX 8", // SP-- (now points at TOS)
		"LDN 8", // A = *SP (TOS)
		"DEX 8", // SP-- (now points at NOS)
		"ADN 8", // A += *SP (NOS)
		"STN 8", // *SP = sum (overwrites NOS slot)
		"INX 8", // SP++ (one cell lighter than on entry)
		"BRB",
	}

	dupBody := []string{
		"DEX 8", // SP--
		"LDN 8", // A = *SP
		"INX 8", // SP++
		"STN 8", // *SP = A
		"INX 8", // SP++
		"BRB",
	}

	builtins := map[string][]string{
		"+":   plusBody,
		"DUP": dupBody,
	}

	cases := []struct {
		name      string
		input     string
		wantStack []uint16
		wantErr   uint16
	}{
		{"single push", "42", []uint16{42}, 0},
		{"two pushes", "3 5", []uint16{3, 5}, 0},
		{"simple add", "3 5 +", []uint16{8}, 0},
		{"three pushes", "1 2 3", []uint16{1, 2, 3}, 0},
		{"chained add", "1 2 3 + +", []uint16{6}, 0},
		{"add zero", "0 7 +", []uint16{7}, 0},
		{"unknown token", "FOO", nil, 1}, // neither word nor number
		{"dup number", "42 DUP", []uint16{42, 42}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var program [4096]uint16
			copy(program[codeBase:], code)

			// Dictionary at dictBase=0x000. The oldest entry's LINK word
			// sits at RAM[0] — also where the bootloader's JMP 0 lands.
			// We encode it as "JMP 100" so the bootloader trampolines to
			// the outer loop; FIND's walk never follows this LINK (its
			// JIZ-on-entry-addr termination fires one step earlier).
			prev := jmp100
			base := dictBase
			for key, body := range builtins {
				hdr := packEntry(prev, key, false)
				for i, w := range hdr {
					program[int(base)+i] = w
				}
				cfa := base + uint16(len(hdr))
				for i, s := range body {
					w, err := encodeASM3(s)
					if err != nil {
						t.Fatalf("encoding + primitive %q: %v", s, err)
					}
					program[int(cfa)+i] = w
				}
				prev = base
				base = cfa + uint16(len(body))
			}

			// Sentinel entry (LEN=0, LINK=most-recent). HEAD points here.
			sentinel := base
			program[sentinel] = prev
			program[sentinel+1] = 0
			program[headAddr] = sentinel

			// Sentinel ERRFLAG so "never reached HLT" is distinguishable
			// from "HLT ran and wrote 0".
			const errSentinel = uint16(0xDEAD)
			program[errFlagAddr] = errSentinel

			// Two tapes: port 0 feeds the bootloader, port 1 feeds WORD.
			programTape := &TapeReader{tape: program[:]}
			ascii := make([]uint16, len(tc.input))
			for i, c := range []byte(tc.input) {
				ascii[i] = uint16(c)
			}
			tokenTape := &TapeReader{tape: ascii}

			computer := NewSAP3()
			computer.LoadProgram(bootloader)
			computer.RegisterInputDevice(programTape, 0)
			computer.RegisterInputDevice(tokenTape, 1)
			td := &testDebugger{t: t}
			run(computer, td)

			gotErr := computer.RAM.mem[errFlagAddr]
			if gotErr == errSentinel {
				t.Fatalf("ERRFLAG still sentinel 0x%X — outer loop never reached HLT. Pipeline stalled (infinite loop, or WORD/FIND/NUMBER stubbed out).", gotErr)
			}
			if gotErr != tc.wantErr {
				t.Errorf("ERRFLAG: got %d, want %d", gotErr, tc.wantErr)
			}

			sp := computer.X[spXReg].Out()
			if sp < stackBase {
				t.Fatalf("SP (X%d = 0x%X) below stackBase 0x%X — stack underflow (pop without push?)",
					spXReg, sp, stackBase)
			}
			depth := int(sp - stackBase)
			gotStack := make([]uint16, depth)
			for i := 0; i < depth; i++ {
				gotStack[i] = computer.RAM.mem[stackBase+uint16(i)]
			}

			// On error paths the stack state is undefined (depends on
			// what was pushed before the bad token), so only assert on
			// successful runs.
			if tc.wantErr == 0 {
				if !reflect.DeepEqual(gotStack, tc.wantStack) {
					t.Errorf("stack: got %v (depth %d), want %v (depth %d)",
						gotStack, len(gotStack), tc.wantStack, len(tc.wantStack))
				}
			}
		})
	}
}
