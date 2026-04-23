package main

import "testing"

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
        "XCH 1", // high/low store mode
        "LDX 2,{:WB_INIT}", // address to store word
        "LDX 3,{:WB_INIT}", // address to store length
		// consume leading delimiters
        "INP 1",
        "JAZ {END}",   // leading 0 is a parse error
        "SBM", "0x21",
        "JAM {WORD}",  // if A - 0x21 is minus, A was 0x20 or less (ignoring 0xFFFF)
        "ADM", "0x21",
        
        "START:", // because of leading parse, input starts in A already
        "JAZ {END}",
        "SBM", "0x21",
        "JAM {END}",    // found a delimiter, end. delimiter is consumed
        "ADM", "0x21",
        "INX 0",
        "JIZ 1,{:PREPHIGH}",

        // PREPLOW:
        "ORM",
        "TEMP:", "",
        "INX 1",
        "JMP {STORE}",

        "PREPHIGH:",
        "LDX 4,{:EIGHT}",
        "SHIFT:",
        "SHL", // this could be replaced by a single << 8 instr
        "DSZ 4",
        "JMP {SHIFT}",
        "STA {TEMP}",
        "INX 2",
        "DEX 1",

        "STORE:",
        "STN 2",
        "INP 1",    // read next token here!
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
