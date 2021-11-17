package main

// first pass assembler, no symbols. supported instructions:
// dest=comp;jump
// - dest can be A, M, D (or a combination but lets support that later)
// - = will remain even if no destination
// - comp is A/M/D operator A/M/D (cant deal with shifts yet)
// - ; will remain even if no jump
// - jump is JGT, JEQ, JGE, JLT, JNE, JLE, JMP or missing
// @hexvalue
// - will get translated to hexvalue instr setting A=hexvalue
// line starting with // which will get ignored

// translate statement in mem to machine binary
// read and print line, parse, print output in hex
// helloworld, actual assembler, writeHex
// statement starts at 0x1000 in memory

// R0: shared stack pointer
// stack memory starting at 0x10 and growing down. R0 points to (empty) top of stack
// R1: memory pointer starting at 0x1000
// R2: shared arg value
// R5: shared screen pointer
// used by others: R4, R6, R7 ?
var assembleStatement = []uint16{
    0x10,   // @0x10
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0 = 0x10
    0x1000, // @0x1000
    0xEC10, // D=A
    0x1,    // @R1
    0xE308, // M=D // R1 = 0x1000
    0x4000, // @0x4000
    0xEC10, // D=A
    0x5,    // @R5
    0xE308, // M=D // R5 = 0x4000

    0x1,    // @R1
    0xFC20, // A=M
    0xFC10, // D=M
    // write char
    // switch: start of line. AMD= / @ / //
    // TODO
}

// R0: shared instr pointer
// R1: memory pointer
// R2: char value passes to drawChar
// R5: screen pointer (shared)
// R6 and R7: temp
var writeHex = []uint16{
    0x1000, // @0x1000
    0xEC10, // D=A
    0x1,    // @R1
    0xE308, // M=D // R1 = 0x1000
    0x4000, // @0x4000
    0xEC10, // D=A
    0x5,    // @R5
    0xE308, // M=D // R5 = 0x4000
    // load value at MEM[R1] into R6
    0x1,    // @R1
    0xFC20, // A=M
    0xFC10, // D=M
    0x6,    // @R6
    0xE308, // M=D
    // set i
    0x4,    // @4
    0xEC10, // D=A
    0x7,    // R7
    0xE308, // M=D

    // 4x read value from mem, shift bits, mask last 4, compare and draw char
    // (LOOP) 17 -> 826
    0x6,    // @R6
    0xD218, // MD=M<<4
    0xF,    // @15
    0xE010, // D=D&A 
    // if D > 9 -> drawA-F else draw0-9 
    0x9,    // @9
    0xE4D0, // D=D-A
    0x344,  // @DRW091
    0xE306, // D;JLE // goto DRW091
    // (DRWAF1)
    // // add some more equal to difference between 9 and A in ASCII
    0x7,    // @7
    0xE090, // D=D+A
    // (DRW091) 27 -> 836
    0x39,   // @57 // we subtracted 9 and want to get to ascii value of digit
    0xE090, // D=D+A
    // now D is ASCII value of highest 4 bits of read value in memory

    // write char
    0x2,    // @R2
    0xE308, // M=D // R2=char to write, in ascii
    0x34E,  // @DECRI
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0=ref
    0x2,    // @2 (start of drawChar)
    0xEA87, // 0;JMP

    // we come back here after drawChar
    // (DECRI) 37 -> 846
    0x7,    // @R7
    0xFC98, // MD=M-1
    0x5,    // R5
    0xFDC8, // M=M+1
    0x33A,  // @LOOP
    0xE305, // D;JNE

    // (END) 43 -> 852
    0x354,  // @END
    0xEA87, // 0;JMP // goto END
}
