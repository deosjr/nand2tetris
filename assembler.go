package main

// first pass assembler, no symbols. supported instructions:
// (dest=)comp(;jump) , both dest and jump are optional
// - dest can be A, M, D or a combination
// - comp is A/M/D operator A/M/D (cant deal with shifts yet)
// - jump is JGT, JEQ, JGE, JLT, JNE, JLE, JMP or missing
// @hexvalue (for now expect always %4, i.e. 0x1 = 0001)
// - will get translated to hexvalue instr setting A=hexvalue
// line starting with // which will get ignored
// any failures are syntax errors, which for now just jump to END
// maybe this first assembler only deals with happy path
// smarter validation can be written in assembly later :)
// lots of duplication but we can golf it later

// translate statement in mem to machine binary
// statement starts at 0x1000 in memory
// output starts at 0x2000 in memory

// R0: shared stack pointer
// stack memory starting at 0x10 and growing down. R0 points to (empty) top of stack
// R1: memory pointer starting at 0x1000
// R2: memory pointer starting at 0x2000
// R4: used by drawChar
// R5: shared screen pointer
// R6, R7: temp vars
var assembleStatement = []uint16{
    // init vars
    0x10,   // @0x10
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0 = 0x10
    0x1000, // @0x1000
    0xEC10, // D=A
    0x1,    // @R1
    0xE308, // M=D // R1 = 0x1000
    0x2000, // @0x2000
    0xEC10, // D=A
    0x2,    // @R2
    0xE308, // M=D // R2 = 0x2000
    0x6,    // @R7
    0xEA88, // M=0 // R7 = 0

    // read char
    0x1,    // @R1
    0xFC20, // A=M
    0xFC10, // D=M

    // (SWITCH) 17
    // if D==0 goto END 
    0x12C,  // @END
    0xE302, // D;JEQ
    // if D==0x2F (/) goto COMMENT
    0x2F,   // ascii /
    0xE4D0, // D=D-A
    0x106,  // @COMMENT
    0xE302, // D;JEQ
    // if D==0x40 (@) goto AINSTR
    0x11,   // ascii @ - ascii /
    0xE4D0, // D=D-A
    0x10F,  // @AINSTR
    0xE302, // D;JEQ
    // else LOOKAHEAD for start C instr

    // R6 here to build up our instruction, setting individual bits
    0x7000, // 0111 0000 0000 0000
    0xEC10, // D=A
    0x6,    // @R6
    0xE308, // M=D // R6 = 0x7000
    0xD088, // M=M<<1 // R6 = 0xE000

    // (LOOKAHEAD) 32
    // idea here is to find out whether we need to parse a destination or not
    // consume A/M/D until we find something else, then switch on whether its an =
    // NOTE we already consumed a first token! therefore we start by incr R7 to 1

    0x7,    // @R7
    0xFDD8, // DM=M+1
    0x1,    // @R1
    0xF0A0, // A=M+D
    0xFC10, // D=M

    // if D==A goto LOOKAHEAD 
    0x41,   // ascii A
    0xE4D0, // D=D-A
    0x20,   // @LOOKAHEAD
    0xE302, // D;JEQ
    // if D==D goto LOOKAHEAD 
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0x20,   // @LOOKAHEAD
    0xE302, // D;JEQ
    // if D==M goto LOOKAHEAD 
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x20,   // @LOOKAHEAD
    0xE302, // D;JEQ
    // else if we read = goto DEST otherwise goto COMP
    0x10,   // ascii M - ascii =
    0xE090, // D=D+A
    0x60,   // @COMP
    0xE305, // D;JNE

    // DEST
    // if A/M/D set dest bits accordingly, loop
    0x1,    // @R1
    0xFC88, // M=M-1 // because we want to start the loop with incr
    // (DESTA) 55
    // We know that there is an = sign at R1+R6 but checking each character
    // again allows for syntax errors to be detected
    // TODO: this will allow duplicate dest, ie AAAM=D+1
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x41,   // ascii A
    0xE4D0, // D=D-A
    0x44,   // @DESTD
    0xE305, // D;JNE
    // D==A
    0x0020, // @A as dest bit
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0x37,   // DESTA
    0xEA87, // 0;JMP
    // (DESTD) 68
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0x4E,   // @DESTM
    0xE305, // D;JNE
    // D==D
    0x0010, // @D as dest bit
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0x37,   // DESTA
    0xEA87, // 0;JMP
    // (DESTM) 78
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x58,   // @DESTEQ
    0xE305, // D;JNE
    // D==M
    0x0008, // @M as dest bit
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0x37,   // DESTA
    0xEA87, // 0;JMP
    // (DESTEQ) 88
    // if equals sign, goto COMP
    0x10,   // ascii M - ascii =
    0xE090, // D=D+A
    0x1,    // @R1
    0xFDC8, // M=M+1
    0x60,   // COMP
    0xE302, // D;JEQ
    // TODO: else syntax error
    0x12C,  // @END
    0xEA87, // 0;JMP

    // (COMP) 96
    // lookahead one character for an operator: + - & | <
    // if operator, goto binary, otherwise parse unary comp
    0x1,    // @R1
    0xFDE0, // A=M+1
    0xFC10, // D=M
    0x26,   // ascii &
    0xE4D0, // D=D-A
    0,   // @BINARY
    0xE302, // D;JEQ
    0x5,    // ascii + - ascii &
    0xE4D0, // D=D-A
    0,   // @BINARY
    0xE302, // D;JEQ
    0x2,    // ascii - - ascii +
    0xE4D0, // D=D-A
    0,   // @BINARY
    0xE302, // D;JEQ
    0xF,    // ascii < - ascii -
    0xE4D0, // D=D-A
    0,   // @BINARY
    0xE302, // D;JEQ
    0x40,   // ascii | - ascii <
    0xE4D0, // D=D-A
    0,   // @BINARY
    0xE302, // D;JEQ

    // UNARY 119
    0x1,    // @R1
    0xFC20, // A=M
    0xFC10, // D=M
    0x21,   // ascii !
    0xE4D0, // D=D-A
    0xB3,   // @NOT
    0xE302, // D;JEQ
    0xC,    // ascii - - ascii !
    0xE4D0, // D=D-A
    0xD4,   // @NEG
    0xE302, // D;JEQ
    0x3,    // ascii 0 - ascii -
    0xE4D0, // D=D-A
    0x8C,   // @ONE
    0xE305, // D;JNE
    // ZERO
    0x0A80, // 0 comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP

    // (ONE) 140
    0xE390, // D=D-1 // ascii 1 - ascii 0
    0x95,   // @UNRYA
    0xE305, // D;JNE
    0x0FC0, // 1 comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP

    // (UNRYA) 149
    0x10,   // ascii A - ascii 1
    0xE4D0, // D=D-A
    0x9F,   // @UNRYD
    0xE305, // D;JNE
    0x0C00, // A comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP

    // (UNRYD) 159
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0xA9,   // @UNRYM
    0xE305, // D;JNE
    0x0300, // D comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP

    // (UNRYM) 169
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x12C,  // @END // TODO: syntax error
    0xE305, // D;JNE
    0x1C00, // M comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP

    // (NOT) 179
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x41,   // ascii A
    0xE4D0, // D=D-A
    0xC0,   // @NOTD
    0xE305, // D;JNE
    0x0C40, // !A comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP
    // (NOTD) 192
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0xCA,   // @NOTM
    0xE305, // D;JNE
    0x0340, // !D comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP
    // (NOTM) 202
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x12C,  // @END // TODO: syntax error
    0xE305, // D;JNE
    0x1C40, // !M comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP

    // (NEG) 212
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x31,   // ascii 1
    0xE4D0, // D=D-A
    0xE1,   // @NEGA
    0xE305, // D;JNE
    0x0E80, // -1 comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP
    // (NEGA) 225
    0x10,   // ascii A - ascii 1
    0xE4D0, // D=D-A
    0xEB,   // @NEGD
    0xE305, // D;JNE
    0x0CC0, // -A comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP
    // (NEGD) 235
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0xF5,   // @NEGM
    0xE305, // D;JNE
    0x03C0, // -D comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP
    // (NEGM) 245
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x12C,  // @END // TODO: syntax error
    0xE305, // D;JNE
    0x1CC0, // -M comp bits
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0xFF,   // @JUMP
    0xEA87, // 0;JMP ; goto JUMP

    // (BINARY)

    // (JUMP) 255
    // parse ENTER or ;J then two letter combo. set jump bits accordingly
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x80,   // ascii ENTER
    0xE4D0, // D=D-A
    0x126,  // @WRITE
    0xE302, // D;JEQ
    // TODO: parse the jump instruction part

    // (COMMENT) 262 -> consume rest of the line
    // TODO: syntax error if not followed by another / first
    // TODO: this only works for line comments, not inline after instr
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    // if D==0x80 (ENTER) goto END 
    0x80,   // ascii ENTER
    0xE4D0, // D=D-A
    0x12C,  // @END
    0xE302, // D;JEQ
    // else goto COMMENT
    0x106,  // @COMMENT
    0xEA87, // 0;JMP

    // (AINSTR) 271 -> parse rest as hex value
    // TODO: assumes max 4 valid hex chars follow!
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    // if D==0x80 (ENTER) goto WRITE 
    0x80,   // ascii ENTER
    0xE4D0, // D=D-A
    0x126,  // @WRITE
    0xE302, // D;JEQ
    // TODO: if not 0-9 or A-F, goto END
    0x80,   // otherwise set D back to read value
    0xE090, // D=D+A
    0x6,    // @R6
    0xD208, // M=M<<4

    0x41, // ascii A
    0xE4D0, // D=D-A
    0x120,  // @ALPHANUM 
    0xE303, // D;JGE // if D-65 >= 0, we have a A-F char
    // only for digits: add back to map 0-9A-F continuous
    0x7, // ascii A - ascii 9 - 1
    0xE090, // D=D+A
    // (ALPHANUM) 288
    // now [0,9] -> [-10,-1] and [A-F] -> [0,5]
    0xA,    // 10
    0xE090, // D=D+A
    0x6,    // @R6
    0xF548, // M=M|D
    0x10F,  // @AINSTR
    0xEA87, // 0;JMP // goto AINSTR

    // (WRITE) 294 write instruction to mem
    // R6 should contain the instruction now
    0x6, // @R6
    0xFC10, // D=M
    0x2,    // @R2
    0xFDC8, // M=M+1
    0xFCA0, // A=M-1
    0xE308, // M=D

    // (END) 300
    0x12C,  // @END
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
