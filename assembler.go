package main

// uses: 
// R0: shared instr pointer // I already want a callstack + stackptr...
// R1: memory pointer
// R2: char value passes to drawChar
// R5: screen pointer (shared)
var writeHex = []uint16{
    0x1000, // @0x1000
    0xEC10, // D=A
    0x1,    // @R1
    0xE308, // M=D // R1 = 0x1000
    0x4000, // @0x4000
    0xEC10, // D=A
    0x5,    // @R5
    0xE308, // M=D // R5 = 0x4000

    // 4x read value from mem, shift bits, mask last 4, compare and draw char
    // not a loop since we cannot pass the shift distance as a variable..
    // drawAF/draw09 should be a subroutine but i dont have a callstack yet
    0x1,    // @R1
    0xFC20, // A=M
    0xD210, // D=M<<4
    0xF,    // @15
    0xE010, // D=D&A 
    // if D > 9 -> drawA-F else draw0-9 
    0x9,    // @9
    0xE4D0, // D=D-A
    0x13,   // @DRW091
    0xE306, // D;JLE // goto DRW091
    // (DRWAF1)
    // // add some more equal to difference between 9 and A in ASCII
    0x7,    // @7
    0xE090, // D=D+A
    // (DRW091) 19
    0x39,   // @57 // we subtracted 9 and want to get to ascii value of digit
    0xE090, // D=D+A
    // now D is ASCII value of highest 4 bits of read value in memory

    // write char
    0x2,    // @R2
    0xE308, // M=D // R2=char to write, in ascii
    0x1D,   // @END
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0=ref
    0x36,   // @54, first instr after this func (assumed start of drawChar)
    0xEA87, // 0;JMP

    // (END) 29
    0x1D,   // @END
    0xEA87, // 0;JMP // goto END

    // fill with noop until we hit 54 instructions here as well
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
    0xEA80,
}
