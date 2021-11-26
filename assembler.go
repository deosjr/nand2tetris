package main

// first pass assembler, no symbols. supported instructions:
// (dest=)comp(;jump) , both dest and jump are optional
// - dest can be A, M, D or a combination
// - comp as defined in the Hack spec
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
// stack memory starting at 0x10 and growing down. R0 points to top of stack
// R1: memory pointer starting at 0x1000
// R2: memory pointer starting at 0x2000
// R4: used by drawChar
// R5: shared screen pointer
// R6, R7: temp vars
var assembleFirstPass = []uint16{
    // init vars
    0x10,   // @0x10
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0 = 0x10
    0x1000, // @0x1000
    0xEC10, // D=A
    0x1,    // @R1
    0xE308, // M=D // R1 = 0x1000
    0x1000, // @0x1000 // TODO not enough space! need to overwrite input
    0xEC10, // D=A
    0x2,    // @R2
    0xE308, // M=D // R2 = 0x1000
    0x6,    // @R6
    0xEA88, // M=0 // R6 = 0
    0x7,    // @R7
    0xEA88, // M=0 // R7 = 0

    // (START) 16
    // read char
    0x1,    // @R1
    0xFC20, // A=M
    0xFC10, // D=M
    // if D==0 goto END 
    0x245,  // @END
    0xE302, // D;JEQ
    // if D==0x20 (SPACE) goto START, allowing leading spaces (indents)
    0x20,   // ascii SPACE
    0xE4D0, // D=D-A
    0x1D,   // @STARTCOMMENT
    0xE305, // D;JNE
    0x1,    // @R1
    0xFDC8, // M=M+1
    0x10,   // @START
    0xEA87, // 0;JMP
    // (STARTCOMMENT) 29
    // if D==0x2F (/) goto COMMENT
    0xF,    // ascii / - ascii SPACE
    0xE4D0, // D=D-A
    0x1EA,  // @COMMENT
    0xE302, // D;JEQ
    // if D==0x40 (@) goto AINSTR
    0x11,   // ascii @ - ascii /
    0xE4D0, // D=D-A
    0x1FA,  // @AINSTR
    0xE302, // D;JEQ
    // else LOOKAHEAD for start C instr

    // R6 here to build up our instruction, setting individual bits
    0x7000, // 0111 0000 0000 0000
    0xEC10, // D=A
    0x6,    // @R6
    0xE308, // M=D // R6 = 0x7000
    0xD088, // M=M<<1 // R6 = 0xE000

    // (LOOKAHEAD) 42
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
    0x2A,   // @LOOKAHEAD
    0xE302, // D;JEQ
    // if D==D goto LOOKAHEAD 
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0x2A,   // @LOOKAHEAD
    0xE302, // D;JEQ
    // if D==M goto LOOKAHEAD 
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x2A,   // @LOOKAHEAD
    0xE302, // D;JEQ
    // else if we read = goto DEST otherwise goto COMP
    0x10,   // ascii M - ascii =
    0xE090, // D=D+A
    0x6A,   // @COMP
    0xE305, // D;JNE

    // DEST
    // if A/M/D set dest bits accordingly, loop
    0x1,    // @R1
    0xFC88, // M=M-1 // because we want to start the loop with incr
    // (DESTA) 65
    // We know that there is an = sign at R1+R7 but checking each character
    // again allows for syntax errors to be detected
    // TODO: this will allow duplicate dest, ie AAAM=D+1
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x41,   // ascii A
    0xE4D0, // D=D-A
    0x4E,   // @DESTD
    0xE305, // D;JNE
    // D==A
    0x0020, // @A as dest bit
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0x41,   // @DESTA
    0xEA87, // 0;JMP
    // (DESTD) 78
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0x58,   // @DESTM
    0xE305, // D;JNE
    // D==D
    0x0010, // @D as dest bit
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0x41,   // @DESTA
    0xEA87, // 0;JMP
    // (DESTM) 88
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x62,   // @DESTEQ
    0xE305, // D;JNE
    // D==M
    0x0008, // @M as dest bit
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    0x41,   // @DESTA
    0xEA87, // 0;JMP
    // (DESTEQ) 98
    // if equals sign, goto COMP
    0x10,   // ascii M - ascii =
    0xE090, // D=D+A
    0x1,    // @R1
    0xFDC8, // M=M+1
    0x6A,   // @COMP
    0xE302, // D;JEQ
    0x245,  // @END // TODO syntax error
    0xEA87, // 0;JMP

    // (COMP) 106
    // lookahead one character for an operator: + - & | <
    // if operator, goto binary, otherwise parse unary comp
    0x1,    // @R1
    0xFDE0, // A=M+1
    0xFC10, // D=M
    0x26,   // ascii &
    0xE4D0, // D=D-A
    0xF1,   // @BINARY
    0xE302, // D;JEQ
    0x5,    // ascii + - ascii &
    0xE4D0, // D=D-A
    0xF1,   // @BINARY
    0xE302, // D;JEQ
    0x2,    // ascii - - ascii +
    0xE4D0, // D=D-A
    0xF1,   // @BINARY
    0xE302, // D;JEQ
    0xF,    // ascii < - ascii -
    0xE4D0, // D=D-A
    0xF1,   // @BINARY
    0xE302, // D;JEQ
    0x40,   // ascii | - ascii <
    0xE4D0, // D=D-A
    0xF1,   // @BINARY
    0xE302, // D;JEQ

    // UNARY
    0x1,    // @R1
    0xFC20, // A=M
    0xFC10, // D=M
    0x21,   // ascii !
    0xE4D0, // D=D-A
    0xB3,   // @NOT
    0xE302, // D;JEQ
    0xC,    // ascii - - ascii !
    0xE4D0, // D=D-A
    0xCE,   // @NEG
    0xE302, // D;JEQ
    0x3,    // ascii 0 - ascii -
    0xE4D0, // D=D-A
    0x94,   // @ONE
    0xE305, // D;JNE
    // ZERO
    0x0A80, // 0 comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP

    // (ONE) 148
    0xE390, // D=D-1 // ascii 1 - ascii 0
    0x9B,   // @UNRYA
    0xE305, // D;JNE
    0x0FC0, // 1 comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP

    // (UNRYA) 155
    0x10,   // ascii A - ascii 1
    0xE4D0, // D=D-A
    0xA3,   // @UNRYD
    0xE305, // D;JNE
    0x0C00, // A comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP

    // (UNRYD) 163
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0xAB,   // @UNRYM
    0xE305, // D;JNE
    0x0300, // D comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP

    // (UNRYM) 171
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x245,  // @END // TODO: syntax error
    0xE305, // D;JNE
    0x1C00, // M comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP

    // (NOT) 179
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x41,   // ascii A
    0xE4D0, // D=D-A
    0xBE,   // @NOTD
    0xE305, // D;JNE
    0x0C40, // !A comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (NOTD) 190
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0xC6,   // @NOTM
    0xE305, // D;JNE
    0x0340, // !D comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (NOTM) 198
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x245,  // @END // TODO: syntax error
    0xE305, // D;JNE
    0x1C40, // !M comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP

    // (NEG) 206
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x31,   // ascii 1
    0xE4D0, // D=D-A
    0xD9,   // @NEGA
    0xE305, // D;JNE
    0x0E80, // -1 comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (NEGA) 217
    0x10,   // ascii A - ascii 1
    0xE4D0, // D=D-A
    0xE1,   // @NEGD
    0xE305, // D;JNE
    0x0CC0, // -A comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (NEGD) 225
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0xE9,   // @NEGM
    0xE305, // D;JNE
    0x03C0, // -D comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (NEGM) 233
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x245,  // @END // TODO: syntax error
    0xE305, // D;JNE
    0x1CC0, // -M comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP

    // (BINARY) 241
    0x1,    // @R1
    0xFC20, // A=M
    0xFC10, // D=M
    0x41,   // ascii A
    0xE4D0, // D=D-A
    0x106,  // @BNRYA
    0xE302, // D;JEQ
    0x3,    // ascii D - ascii A
    0xE4D0, // D=D-A
    0x12A,  // @BNRYD
    0xE302, // D;JEQ
    0x9,    // ascii M - ascii D
    0xE4D0, // D=D-A
    0x102,  // @BNRYM
    0xE302, // D;JEQ
    0x245,  // @END // TODO: syntax error
    0xEA87, // 0;JMP

    // (BNRYM) 258
    // set the M flag, then fall through to A case
    0x1000, // 0001 0000 0000 0000
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    // (BNRYA) 262
    // read the next two chars as 8bit values into D then compare
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xD410, // D=M<<8
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xF550, // D=D|M
    0x2B31, // ascii +1 = 0x2B31
    0xE4D0, // D=D-A
    0x114,  // @MIN1
    0xE305, // D;JNE
    0x0DC0, // A/M +1 comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (MIN1) 276
    0x0200, // ascii -1 = 0x2D31, 0x0200 more than +1
    0xE4D0, // D=D-A
    0x11C,  // @MIND
    0xE305, // D;JNE
    0x0C80, // A/M -1 comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (MIND) 284
    0x0013, // ascii -D = 0x2D44, 0x0013 more than -1
    0xE4D0, // D=D-A
    0x124,  // @ASHIFT
    0xE305, // D;JNE
    0x01C0, // A/M -D comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (ASHIFT) 292
    0x0EF8, // ascii << = 0x3C3C, 0x0EF8 more than -D
    0xE4D0, // D=D-A
    0x187,  // @SHIFT
    0xE302, // D;JEQ
    0x245,  // @END // TODO: syntax error
    0xEA87, // 0;JMP

    // (BNRYD) 298
    // read the next two chars as 8bit values into D then compare
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xD410, // D=M<<8
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xF550, // D=D|M
    0x2641, // ascii &A = 0x2641
    0xE4D0, // D=D-A
    0x137,  // @ANDM
    0xE305, // D;JNE
    // D&A comp bits are all 0
    0xEA90, // D=0
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (ANDM) 311
    0x000C, // ascii &M = 0x264D, 0x000C more than &A
    0xE4D0, // D=D-A
    0x13F,  // @PLUS1
    0xE305, // D;JNE
    0x1000, // D&M comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (PLUS1) 319
    0x04E4, // ascii +1 = 0x2B31, 0x04E4 more than &M
    0xE4D0, // D=D-A
    0x147,  // @PLUSA
    0xE305, // D;JNE
    0x07C0, // D+1 comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (PLUSA) 327
    0x0010, // ascii +A = 0x2B41, 0x0010 more than +1
    0xE4D0, // D=D-A
    0x14F,  // @PLUSM
    0xE305, // D;JNE
    0x0080, // D+A comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (PLUSM) 335
    0x000C, // ascii +M = 0x2B4D, 0x000C more than +A
    0xE4D0, // D=D-A
    0x157,  // @DMIN1
    0xE305, // D;JNE
    0x1080, // D+M comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (DMIN1) 343
    0x01E4, // ascii -1 = 0x2D31, 0x01E4 more than +M
    0xE4D0, // D=D-A
    0x15F,  // @MINA
    0xE305, // D;JNE
    0x0380, // D-1 comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (MINA) 351
    0x0010, // ascii -A = 0x2D41, 0x0010 more than -1
    0xE4D0, // D=D-A
    0x167,  // @MINM
    0xE305, // D;JNE
    0x04C0, // D-A comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (MINM) 359
    0x000C, // ascii -M = 0x2D4D, 0x000C more than -A
    0xE4D0, // D=D-A
    0x16F,  // @ORA
    0xE305, // D;JNE
    0x14C0, // D-M comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (ORA) 367
    0x4EF4, // ascii |A = 0x7C41, 0x4EF4 more than -M
    0xE4D0, // D=D-A
    0x177,  // @ORM
    0xE305, // D;JNE
    0x0540, // D|A comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (ORM) 375
    0x000C, // ascii |M = 0x7C4D, 0x000C more than |A
    0xE4D0, // D=D-A
    0x17F,  // @DSHIFT
    0xE305, // D;JNE
    0x1540, // D|M comp bits
    0xEC10, // D=A
    0x191,  // @ENDCOMP
    0xEA87, // 0;JMP // goto ENDCOMP
    // (DSHIFT) 383
    0x4011, // ascii << = 0x3C3C, 0x4011 LESS than |M
    0xE090, // D=D+A
    0x245,  // @END // TODO syntax error
    0xE305, // D;JNE
    0x0800, // D bit for the shift operation
    0xEC10, // D=A
    0x6,    // @R6
    0xF548, // M=M|D
    // fall through to SHIFT

    // (SHIFT) 391
    // We can come here from either A, D or M parsed, flags have been set accordingly
    // set the 3rd bit to 0 again using AND mask (but cant use highest bit in A instr!)
    // so instead we simply subtract 0x2000 (since we know that bit is set to 1 earlier)
    0x2000, // 0010 0000 0000 0000
    0xEC10, // D=A
    0x6,    // @R6
    0xF1C8, // M=M-D
    // TODO: for now maybe just assume valid 1-8 char follows? 0x31 - 0x38
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x30,   // maps [0x31, 0x38] -> [1, 8] 
    0xE4D0, // D=D-A
    0xCB90, // D=D<<7
    // fall through to endcomp

    // (ENDCOMP) 401
    // whenever we come to jump, we have just set D=bits to add to instr
    // so the first thing we do is to add those to R6
    0x6,    // @R6
    0xF548, // M=M|D
    // check whether we stop early in ENDLINE func
    0x19A,  // @JUMP
    0xEC10, // D=A
    0x0,    // @R0
    0xFDE8, // AM=M+1
    0xE308, // M=D
    0x21C,  // @ENDLINE
    0xEA87, // 0;JMP
    // otherwise parse the jump instruction part
    // (JUMP) 410
    // parse ENTER or ;J then two letter combo. set jump bits accordingly
    0x45,   // ascii ENTER - ascii ;
    0xE090, // D=D+A
    0x245,  // @END // TODO syntax error
    0xE305, // D;JNE
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    0x4A,   // ascii J
    0xE4D0, // D=D-A
    0x245,  // @END // TODO syntax error
    0xE305, // D;JNE
    // read the next two chars as 8bit values into D then compare
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xD410, // D=M<<8
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xF550, // D=D|M
    0x4551, // ascii EQ = 0x4551
    0xE4D0, // D=D-A
    0x1B3,  // @JGE
    0xE305, // D;JNE
    0x0002, // JEQ jump bits
    0xEC10, // D=A
    0x1E1,  // @ENDJUMP
    0xEA87, // 0;JMP ; goto ENDJUMP
    // (JGE) 435
    0x01F4, // ascii GE = 0x4745, 0x01F4 more than EQ
    0xE4D0, // D=D-A
    0x1BB,  // @JGT
    0xE305, // D;JNE
    0x0003, // JGE jump bits
    0xEC10, // D=A
    0x1E1,  // @ENDJUMP
    0xEA87, // 0;JMP ; goto ENDJUMP
    // (JGT) 443
    0x000F, // ascii GT = 0x4754, 0x000F more than GE
    0xE4D0, // D=D-A
    0x1C3,  // @JLE
    0xE305, // D;JNE
    0x0001, // JGT jump bits
    0xEC10, // D=A
    0x1E1,  // @ENDJUMP
    0xEA87, // 0;JMP ; goto ENDJUMP
    // (JLE) 451
    0x04F1, // ascii LE = 0x4C45, 0x04F1 more than GT
    0xE4D0, // D=D-A
    0x1CB,  // @JLT
    0xE305, // D;JNE
    0x0006, // JLE jump bits
    0xEC10, // D=A
    0x1E1,  // @ENDJUMP
    0xEA87, // 0;JMP ; goto ENDJUMP
    // (JLT) 459
    0x000F, // ascii LT = 0x4C54, 0x000F more than LE
    0xE4D0, // D=D-A
    0x1D3,  // @JMP
    0xE305, // D;JNE
    0x0004, // JLT jump bits
    0xEC10, // D=A
    0x1E1,  // @ENDJUMP
    0xEA87, // 0;JMP ; goto ENDJUMP
    // (JMP) 467
    0x00FC, // ascii MP = 0x4D50, 0x00FC more than LT
    0xE4D0, // D=D-A
    0x1DB,  // @JNE
    0xE305, // D;JNE
    0x0007, // JMP jump bits
    0xEC10, // D=A
    0x1E1,  // @ENDJUMP
    0xEA87, // 0;JMP ; goto ENDJUMP
    // (JNE) 475
    0x00F5, // ascii NE = 0x4E45, 0x00F5 more than MP
    0xE4D0, // D=D-A
    0x245,  // @END // TODO syntax error
    0xE305, // D;JNE
    0x0005, // JNE jump bits
    0xEC10, // D=A
    // fall through to ENDJUMP

    // whenever we come to endjump, we have just set D=bits to add to instr
    // so the first thing we do is to add those to R6
    // (ENDJUMP) 481
    0x6,    // @R6
    0xF548, // M=M|D
    // check whether we stop in ENDLINE func
    // if we do not, we return to syntax error
    0x245,  // @END // TODO syntax error
    0xEC10, // D=A
    0x0,    // @R0
    0xFDE8, // AM=M+1
    0xE308, // M=D
    0x21C,  // @ENDLINE
    0xEA87, // 0;JMP

    // (COMMENT) 490 -> consume second slash
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    // if D!=0x2F (/) goto END TODO syntax error
    0x2F,   // ascii /
    0xE4D0, // D=D-A
    0x245,  // @END
    0xE305, // D;JNE
    // (COMMENTREC) 497 -> consume rest of the line
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    // if D==0x80 (ENTER) goto WRITE 
    0x80,   // ascii ENTER
    0xE4D0, // D=D-A
    0x231,  // @WRITE
    0xE302, // D;JEQ
    // else goto COMMENTREC
    0x1F1,  // @COMMENTREC
    0xEA87, // 0;JMP

    // (AINSTR) 506 -> parse rest as hex value
    // TODO: assumes max 4 valid hex chars follow!
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    // if D==0x80 (ENTER) goto WRITE 
    0x80,   // ascii ENTER
    0xE4D0, // D=D-A
    0x231,  // @WRITE
    0xE302, // D;JEQ
    // TODO: if not 0-9 or A-F, goto END
    0x80,   // otherwise set D back to read value
    0xE090, // D=D+A
    0x6,    // @R6
    0xD208, // M=M<<4
    0x41,   // ascii A
    0xE4D0, // D=D-A
    0x20B,  // @ALPHANUM 
    0xE303, // D;JGE // if D-65 >= 0, we have a A-F char
    // only for digits: add back to map 0-9A-F continuous
    0x7, // ascii A - ascii 9 - 1
    0xE090, // D=D+A
    // (ALPHANUM) 523
    // now [0,9] -> [-10,-1] and [A-F] -> [0,5]
    0xA,    // 10
    0xE090, // D=D+A
    0x6,    // @R6
    0xF548, // M=M|D
    0x7,    // @R7
    0xFDD8, // DM=M+1 // R7+=1
    0x4,    // @4
    0xE4D0, // D=D-A
    0x1FA, // @AINSTR
    0xE305, // D;JNE
    // after parsing 4 hex chars, check whether we stop in ENDLINE func
    // if we do not, we return to syntax error
    0x245,  // @END // TODO syntax error
    0xEC10, // D=A
    0x0,    // @R0
    0xFDE8, // AM=M+1
    0xE308, // M=D
    0x21C,  // @ENDLINE
    0xEA87, // 0;JMP

    // (ENDLINE) 540
    // function called from end of comp, jump and ainstr
    // consumes trailing spaces, comments and newline
    // jumps back if it FAILS to find any of those
    // otherwise cuts parsing the line short and never returns
    0x0,    // @R0
    0xFC88, // M=M-1  // whatever happens, decrement stack pointer
    // (ENDLINELOOP) 542
    0x1,    // @R1
    0xFDE8, // AM=M+1
    0xFC10, // D=M
    // if D==0x20 (SPACE) goto ENDLINELOOP, allowing trailing spaces
    0x20,   // ascii SPACE
    0xE4D0, // D=D-A
    0x21E,  // @ENDLINELOOP
    0xE302, // D;JEQ
    // if D==0x2F (/) goto COMMENT
    0xF,    // ascii / - ascii SPACE
    0xE4D0, // D=D-A
    0x1EA,  // @COMMENT
    0xE302, // D;JEQ
    // if D==0x80 (ENTER) goto WRITE
    0x51,   // ascii ENTER - ascii /
    0xE4D0, // D=D-A
    0x231,  // @WRITE
    0xE302, // D;JEQ
    // goto the address @R0 pointed to
    0x0,    // @R0
    0xFDE0, // A=M+1 // already decremented stack pointer
    0xFC20, // A=M
    0xEA87, // 0;JMP // return to caller

    // (WRITE) 561 write instruction to mem
    // R6 should contain the instruction now
    0x6, // @R6
    0xFC10, // D=M
    0x7, // @R7
    0xF090, // D=D+M
    // if R6 is 0 and R7 is 0 then this was a full line comment, do not write anything
    0x23D,  // @RESET
    0xE302, // D;JEQ
    // otherwise write to mem at R2 and advance counter
    0x6, // @R6
    0xFC10, // D=M
    0x2,    // @R2
    0xFDC8, // M=M+1
    0xFCA0, // A=M-1
    0xE308, // M=D

    // (RESET) 573 consume newline (assume found, otherwise shouldve been syntax error) and parse next line
    0x6,    // @R6
    0xEA88, // M=0 // R6 = 0
    0x7,    // @R7
    0xEA88, // M=0 // R7 = 0
    0x1,    // @R1
    0xFDC8, // M=M+1
    0x10,   // @START
    0xEA87, // 0;JMP

    // (END) 581
    0x245,  // @END
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

// assembled using firstpass, then removed comments in front of label statements
var twopass = []uint16{
    0x0010,
    0xec10,
    0x0000,
    0xe308,
    0x1000,
    0xec10,
    0x0001,
    0xe308,
    0x1000,
    0xec10,
    0x0002,
    0xe308,
    0x0006,
    0xea88,
    0x0007,
    0xea88,
    0x0020,
    0xec10,
    0x0008,
    0xe308,
    0x0001,
    0xfc20,
    0xfc10,
    0x005e,
    0xe302,
    0x0028,
    0xe4d0,
    0x002e,
    0xe302,
    0x0007,
    0xe4d0,
    0x0023,
    0xe302,
    0x0006,
    0xfdc8,
    0x0001,
    0xfde8,
    0xfc10,
    0x0080,
    0xe4d0,
    0x0023,
    0xe305,
    0x0001,
    0xfdc8,
    0x0014,
    0xea87,
    0x0001,
    0xfde8,
    0xfc10,
    0x0029,
    0xe4d0,
    0x004c,
    0xe302,
    0x0007,
    0xfdd8,
    0x0001,
    0xe010,
    0x0043,
    0xe302,
    0x0001,
    0xfc20,
    0xd410,
    0x0008,
    0xfc20,
    0xe308,
    0x002e,
    0xea87,
    0x0001,
    0xfc20,
    0xfc10,
    0x0008,
    0xfdc8,
    0xfca0,
    0xf548,
    0x002e,
    0xea87,
    0x0007,
    0xfc10,
    0x0001,
    0xe010,
    0x0054,
    0xe302,
    0x0008,
    0xfdc8,
    0x0006,
    0xfc10,
    0x0008,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0007,
    0xea88,
    0x0023,
    0xea87,
    0x005e,
    0xea87,
}
