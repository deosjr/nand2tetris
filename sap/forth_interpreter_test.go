package main

import (
	"reflect"
	"testing"
)

// TestSAP3ForthFromFile is the file-based counterpart to TestSAP3ForthInterpreter.
// It assembles asm/interpreter.asm, loads it via the bootloader, and runs the
// same test cases. The assembly source is the authoritative implementation;
// the Go-side test is purely the harness.
func TestSAP3ForthFromFile(t *testing.T) {
	const (
		wordBufAddr = uint16(0xE10)
		cfaOutAddr  = uint16(0xE01)
		immOutAddr  = uint16(0xE02)
		numValAddr  = uint16(0xE30)
		numOkAddr   = uint16(0xE31)
		headAddr    = uint16(0xCCC)
		stackBase   = uint16(0xD00)
		errFlagAddr = uint16(0xE50)
		spXReg      = 8
	)

	src, err := readAsmFile("asm/interpreter.asm")
	if err != nil {
		t.Fatalf("reading asm/interpreter.asm: %v", err)
	}

	externals := map[string]uint16{
		"WORDBUF": wordBufAddr,
		"CFAOUT":  cfaOutAddr,
		"IMMOUT":  immOutAddr,
		"NUMVAL":  numValAddr,
		"NUMOK":   numOkAddr,
		"ERRFLAG": errFlagAddr,
	}

	code, _, err := assembleSAP3Labeled(0, src, externals)
	if err != nil {
		t.Fatalf("assembling asm/interpreter.asm: %v", err)
	}
	if len(code) > 256 {
		t.Fatalf("interpreter is %d words, exceeds one 256-word page", len(code))
	}
	t.Logf("interpreter: %d words (%d free on page)", len(code), 256-len(code))

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
		{"unknown token", "FOO", nil, 1}, // neither word nor number → ERRFLAG=1
		{"dup number", "42 DUP", []uint16{42, 42}, 0},
		{"dup latest", "1 2 3 DUP", []uint16{1, 2, 3, 3}, 0},
		{"drop twice", "1 2 3 DROP DROP", []uint16{1}, 0},
		// Hardcoded colon def in dictionary: `: DOUBLE DUP + ;`
		{"colon def DOUBLE", "7 DOUBLE", []uint16{14}, 0},
		{"nested colon", "3 DOUBLE DOUBLE", []uint16{12}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var program [4096]uint16
			copy(program[:], code)

			// Pre-sentinel ERRFLAG so "never reached HLT" is detectable.
			const errSentinel = uint16(0xDEAD)
			program[errFlagAddr] = errSentinel

			// Port 0: program tape (bootloader reads machine code from here).
			// Port 1: token tape (WORD reads ASCII bytes from here).
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
				t.Fatalf("ERRFLAG still sentinel 0x%X — outer loop never reached HLT "+
					"(infinite loop, pipeline stall, or all subroutines stubbed out)", gotErr)
			}
			if gotErr != tc.wantErr {
				t.Errorf("ERRFLAG: got %d, want %d", gotErr, tc.wantErr)
			}

			sp := computer.X[spXReg].Out()
			if sp < stackBase {
				t.Fatalf("SP (X%d = 0x%X) below stackBase 0x%X — stack underflow",
					spXReg, sp, stackBase)
			}
			depth := int(sp - stackBase)
			gotStack := make([]uint16, depth)
			for i := 0; i < depth; i++ {
				gotStack[i] = computer.RAM.mem[stackBase+uint16(i)]
			}

			// Stack contents are only defined on the success path.
			if tc.wantErr == 0 {
				if !reflect.DeepEqual(gotStack, tc.wantStack) {
					t.Errorf("stack: got %v (depth %d), want %v (depth %d)",
						gotStack, len(gotStack), tc.wantStack, len(tc.wantStack))
				}
			}
		})
	}
}
