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
// strategy: parse by character, setting flags for mode

// R0: shared stack pointer
// stack memory starting at 0x10 and growing down. R0 points to (empty) top of stack
// R1: memory pointer starting at 0x1000
// R2: shared arg value
// R4: used by drawChar
// R5: shared screen pointer
// R6: mode
// R7: dest
// R8: comp
// R9: jmp
var assembleStatement = []uint16{
    // init vars
    0x10,   // @0x10
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0 = 0x10
    0x4000, // @0x4000
    0xEC10, // D=A
    0x5,    // @R5
    0xE308, // M=D // R5 = 0x4000
    0x1000, // @0x1000
    0xEC10, // D=A
    0x1,    // @R1
    0xE308, // M=D // R1 = 0x1000

    // (PARSE)
    // read char
    0x1,    // @R1
    0xFC20, // A=M
    0xFC10, // D=M
    0x2,    // @R2
    0xE308, // M=D // R2=ascii from mem
    // write char
    // set MEM[R0] to next line, set R0 to R0+1
    0,  // @SWITCH
    0xEC10, // D=A
    0x0,    // @R0
    0xFDC8, // M=M+1
    0xFCA0, // A=M-1
    0xE308, // M=D
    0,  // @WRITECHAR
    0xEA87, // 0;JMP // goto WRITECHAR

    // (SWITCH) 25 -> 834
    // R2 should still contain last read char from mem
    0x2,    // @R2
    0xFC10, // D=M
    // if D==0 goto END 
    0,  // @END
    0xE302, // D;JEQ
    // if D==0x2F (/) goto COMMENT
    0x2F,   // ascii /
    0xE4D0, // D=D-A
    0,  // @COMMENT
    0xE302, // D;JEQ
    // if D==0x3B (;) goto JUMP
    0x0B,   // ascii ; - ascii /
    0xE4D0, // D=D-A
    0,  // @JUMP
    0xE302, // D;JEQ
    // if D==0x3D (=) goto EQUALS
    0x2,    // ascii = - ascii ;
    0xE4D0, // D=D-A
    0,  // @EQUALS
    0xE302, // D;JEQ
    // if D==0x40 (@) goto AINSTR
    0x3,    // ascii @ - ascii =
    0xE4D0, // D=D-A
    0,  // @AINSTR
    0xE302, // D;JEQ
    // if D=0x41 (A) goto PARSEA
    0x1,    // ascii A - ascii @
    0xE4D0, // D=D-A
    0,  // @PARSEA
    0xE302, // D;JEQ
    // if D=0x44 (D) goto PARSED
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0,  // @PARSED
    0xE302, // D;JEQ
    // if D=0x4D (M) goto PARSEM
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0,  // @PARSEM
    0xE302, // D;JEQ
    // TODO: if we reach here, we have a syntax error

    // (COMMENT) -> consume rest of the line

    // (JUMP) -> parse last 2 jump symbols

    // (EQUALS)

    // (AINSTR) -> only at start, parse rest as hex value

    // (PARSEA)

    // (PARSED)

    // (PARSEM)

    // end of parse loop
    // @PARSE
    // 0;JMP // goto PARSE

    // (WRITECHAR) 27 -> 836
    0,  // @JUMPBACK
    0xEC10, // D=A
    0x0,    // @R0
    0xFDC8, // M=M+1
    0xFCA0, // A=M-1
    0xE308, // M=D
    0x2,    // @2, start of drawChar
    0xEA87, // 0;JMP
    // (JUMPBACK) 35 -> 844
    0x5,    // @screen
    0xFDC8, // M=M+1  // @screen+=1
    // goto the address @R0 points to, decrementing stack pointer in the process
    0x0,    // @R0
    0xFCA8, // AM=M-1
    0xFC20, // A=M
    0xEA87, // 0;JMP // goto caller of WRITECHAR

    // (END) 41 -> 850
    0,  // @END
    0xEA87, // 0;JMP // goto END
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
    0xE308, // M=D // R2=ascii from mem
    // set MEM[R0] to SCRN, set R0 to R0+1
    0x350,  // @DECRI
    0xEC10, // D=A
    0x0,    // @R0
    0xFDC8, // M=M+1
    0xFCA0, // A=M-1
    0xE308, // M=D
    0x2,    // @2, start of drawChar
    0xEA87, // 0;JMP

    // we come back here after drawChar
    // (DECRI) 39 -> 848
    0x7,    // @R7
    0xFC98, // MD=M-1
    0x5,    // R5
    0xFDC8, // M=M+1
    0x33A,  // @LOOP
    0xE305, // D;JNE

    // (END) 45 -> 854
    0x356,  // @END
    0xEA87, // 0;JMP // goto END
}
