// FIRST PASS ASSEMBLER WITHOUT SYMBOLS
    // init vars
    @0010
    D=A
    @0000 // @R0
    M=D   // R0 = 0x10
    @1000
    D=A
    @0001 // @R1
    M=D   // R1 = 0x1000
    @2000
    D=A
    @0002 // @R2
    M=D   // R2 = 0x2000
    @0006 // @R6
    M=0   // R6 = 0
    @0007 // @R7
    M=0   // R7 = 0
// (START) 16
    // read char
    @0001 // @R1
    A=M
    D=M
    // if D==0 goto END 
    @0233   // @END
    D;JEQ
    // if D==0x20 (SPACE) goto START, allowing leading spaces (indents)
    @0020   // ascii SPACE
    D=D-A
    @001D   // @STARTCOMMENT
    D;JNE
    @0001   // @R1
    M=M+1
    @0010   // @START
    0;JMP
// (STARTCOMMENT) 29
    // if D==0x2F (/) goto COMMENT
    @000F   // ascii / - ascii SPACE
    D=D-A
    @01FC   // @COMMENT
    D;JEQ
    // if D==0x40 (@) goto AINSTR
    @0011   // ascii @ - ascii /
    D=D-A
    @020C   // @AINSTR
    D;JEQ
    // else LOOKAHEAD for start C instr
    // R6 here to build up our instruction, setting individual bits
    @7000   // 0111 0000 0000 0000
    D=A
    @0006   // @R6
    M=D     // R6 = 0x7000
    M=M<<1  // R6 = 0xE000
// (LOOKAHEAD) 42
    // idea here is to find out whether we need to parse a destination or not
    // consume A/M/D until we find something else, then switch on whether its an =
    // NOTE we already consumed a first token! therefore we start by incr R7 to 1
    @0007   // @R7
    DM=M+1
    @0001   // @R1
    A=M+D
    D=M
    // if D==A goto LOOKAHEAD 
    @0041   // ascii A
    D=D-A
    @002A   // @LOOKAHEAD
    D;JEQ
    // if D==D goto LOOKAHEAD 
    @0003   // ascii D - ascii A
    D=D-A
    @002A   // @LOOKAHEAD
    D;JEQ
    // if D==M goto LOOKAHEAD 
    @0009   // ascii M - ascii D
    D=D-A
    @002A   // @LOOKAHEAD
    D;JEQ
    // else if we read = goto DEST otherwise goto COMP
    @0010   // ascii M - ascii =
    D=D+A
    @006A   // @COMP
    D;JNE
// DEST
    // if A/M/D set dest bits accordingly, loop
    @0001   // @R1
    M=M-1   // because we want to start the loop with incr
// (DESTA) 65
    // We know that there is an = sign at R1+R7 but checking each character
    // again allows for syntax errors to be detected
    // TODO: this will allow duplicate dest, ie AAAM=D+1
    @0001   // @R1
    AM=M+1
    D=M
    @0041   // ascii A
    D=D-A
    @004E   // @DESTD
    D;JNE
    // D==A
    @0020   // @A as dest bit
    D=A
    @0006   // @R6
    M=M|D
    @0041   // @DESTA
    0;JMP
// (DESTD) 78
    @0003   // ascii D - ascii A
    D=D-A
    @0058   // @DESTM
    D;JNE
    // D==D
    @0010   // @D as dest bit
    D=A
    @0006   // @R6
    M=M|D
    @0041   // @DESTA
    0;JMP
// (DESTM) 88
    @0009   // ascii M - ascii D
    D=D-A
    @0062   // @DESTEQ
    D;JNE
    // D==M
    @0008   // @M as dest bit
    D=A
    @0006   // @R6
    M=M|D
    @0041   // @DESTA
    0;JMP
// (DESTEQ) 98
    // if equals sign, goto COMP
    @0010   // ascii M - ascii =
    D=D+A
    @0001   // @R1
    M=M+1
    @006A   // @COMP
    D;JEQ
    @0233   // @END // TODO syntax error
    0;JMP
// (COMP) 106
    // lookahead one character for an operator: + - & | <
    // if operator, goto binary, otherwise parse unary comp
    @0001   // @R1
    A=M+1
    D=M
    @0026   // ascii &
    D=D-A
    @00F1   // @BINARY
    D;JEQ
    @0005   // ascii + - ascii &
    D=D-A
    @00F1   // @BINARY
    D;JEQ
    @0002   // ascii - - ascii +
    D=D-A
    @00F1   // @BINARY
    D;JEQ
    @000F   // ascii < - ascii -
    D=D-A
    @00F1   // @BINARY
    D;JEQ
    @0040   // ascii | - ascii <
    D=D-A
    @00F1   // @BINARY
    D;JEQ
// UNARY
    @0001   // @R1
    A=M
    D=M
    @0021   // ascii !
    D=D-A
    @00B3   // @NOT
    D;JEQ
    @000C   // ascii - - ascii !
    D=D-A
    @00CE   // @NEG
    D;JEQ
    @0003   // ascii 0 - ascii -
    D=D-A
    @0094   // @ONE
    D;JNE
// ZERO
    @0A80   // 0 comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (ONE) 148
    D=D-1   // ascii 1 - ascii 0
    @009B   // @UNRYA
    D;JNE
    @0FC0   // 1 comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (UNRYA) 155
    @0010   // ascii A - ascii 1
    D=D-A
    @00A3   // @UNRYD
    D;JNE
    @0C00   // A comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (UNRYD) 163
    @0003   // ascii D - ascii A
    D=D-A
    @00AB   // @UNRYM
    D;JNE
    @0300   // D comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (UNRYM) 171
    @0009   // ascii M - ascii D
    D=D-A
    @0233   // @END // TODO: syntax error
    D;JNE
    @1C00   // M comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (NOT) 179
    @0001   // @R1
    AM=M+1
    D=M
    @0041   // ascii A
    D=D-A
    @00BE   // @NOTD
    D;JNE
    @0C40   // !A comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (NOTD) 190
    @0003   // ascii D - ascii A
    D=D-A
    @00C6   // @NOTM
    D;JNE
    @0340   // !D comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (NOTM) 198
    @0009   // ascii M - ascii D
    D=D-A
    @0233   // @END // TODO: syntax error
    D;JNE
    @1C40   // !M comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (NEG) 206
    @0001   // @R1
    AM=M+1
    D=M
    @0031   // ascii 1
    D=D-A
    @00D9   // @NEGA
    D;JNE
    @0E80   // -1 comp bits
    D=A
    @0191,  // @JUMP
    0;JMP
// (NEGA) 217
    @0010   // ascii A - ascii 1
    D=D-A
    @00E1   // @NEGD
    D;JNE
    @0CC0   // -A comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (NEGD) 225
    @0003   // ascii D - ascii A
    D=D-A
    @00E9   // @NEGM
    D;JNE
    @03C0   // -D comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (NEGM) 233
    @0009   // ascii M - ascii D
    D=D-A
    @0233   // @END // TODO: syntax error
    D;JNE
    @1CC0   // -M comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (BINARY) 241
    @0001   // @R1
    A=M
    D=M
    @0041   // ascii A
    D=D-A
    @0106   // @BNRYA
    D;JEQ
    @0003   // ascii D - ascii A
    D=D-A
    @012A   // @BNRYD
    D;JEQ
    @0009   // ascii M - ascii D
    D=D-A
    @0102   // @BNRYM
    D;JEQ
    @0233   // @END // TODO: syntax error
    0;JMP
// (BNRYM) 258
    // set the M flag, then fall through to A case
    @1000   // 0001 0000 0000 0000
    D=A
    @0006   // @R6
    M=M|D
// (BNRYA) 262
    // read the next two chars as 8bit values into D then compare
    @0001   // @R1
    AM=M+1
    D=M<<8
    @0001   // @R1
    AM=M+1
    D=D|M
    @2B31   // ascii +1 = 0x2B31
    D=D-A
    @0114   // @MIN1
    D;JNE
    @0DC0   // A/M +1 comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (MIN1) 276
    @0200   // ascii -1 = 0x2D31, 0x0200 more than +1
    D=D-A
    @011C   // @MIND
    D;JNE
    @0C80   // A/M -1 comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (MIND) 284
    @0013   // ascii -D = 0x2D44, 0x0013 more than -1
    D=D-A
    @0124   // @ASHIFT
    D;JNE
    @01C0   // A/M -D comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (ASHIFT) 292
    @0EF8   // ascii << = 0x3C3C, 0x0EF8 more than -D
    D=D-A
    @0187   // @SHIFT
    D;JEQ
    @0233   // @END // TODO: syntax error
    0;JMP
// (BNRYD) 298
    // read the next two chars as 8bit values into D then compare
    @0001   // @R1
    AM=M+1
    D=M<<8
    @0001   // @R1
    AM=M+1
    D=D|M
    @2641   // ascii &A = 0x2641
    D=D-A
    @0137   // @ANDM
    D;JNE
    // D&A comp bits are all 0
    D=0
    @0191   // @JUMP
    0;JMP
// (ANDM) 311
    0x000C, // ascii &M = 0x264D, 0x000C more than &A
    D=D-A
    @013F   // @PLUS1
    D;JNE
    @1000   // D&M comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (PLUS1) 319
    @04E4   // ascii +1 = 0x2B31, 0x04E4 more than &M
    D=D-A
    @0147   // @PLUSA
    D;JNE
    @07C0   // D+1 comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (PLUSA) 327
    @0010   // ascii +A = 0x2B41, 0x0010 more than +1
    D=D-A
    @014F   // @PLUSM
    D;JNE
    @0080   // D+A comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (PLUSM) 335
    @000C   // ascii +M = 0x2B4D, 0x000C more than +A
    D=D-A
    @0157   // @DMIN1
    D;JNE
    @1080   // D+M comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (DMIN1) 343
    @01E4   // ascii -1 = 0x2D31, 0x01E4 more than +M
    D=D-A
    @015F   // @MINA
    D;JNE
    @0380   // D-1 comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (MINA) 351
    @0010   // ascii -A = 0x2D41, 0x0010 more than -1
    D=D-A
    @0167   // @MINM
    D;JNE
    @04C0   // D-A comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (MINM) 359
    @000C   // ascii -M = 0x2D4D, 0x000C more than -A
    D=D-A
    @016F   // @ORA
    D;JNE
    @14C0   // D-M comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (ORA) 367
    @4EF4   // ascii |A = 0x7C41, 0x4EF4 more than -M
    D=D-A
    @0177   // @ORM
    D;JNE
    @0540   // D|A comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (ORM) 375
    @000C   // ascii |M = 0x7C4D, 0x000C more than |A
    D=D-A
    @017F   // @DSHIFT
    D;JNE
    @1540   // D|M comp bits
    D=A
    @0191   // @JUMP
    0;JMP
// (DSHIFT) 383
    @4011   // ascii << = 0x3C3C, 0x4011 LESS than |M
    D=D+A
    @0233   // @END // TODO syntax error
    D;JNE
    @0800   // D bit for the shift operation
    D=A
    @0006   // @R6
    M=M|D
    // fall through to SHIFT
// (SHIFT) 391
    // We can come here from either A, D or M parsed, flags have been set accordingly
    // set the 3rd bit to 0 again using AND mask (but cant use highest bit in A instr!)
    // so instead we simply subtract 0x2000 (since we know that bit is set to 1 earlier)
    @2000   // 0010 0000 0000 0000
    D=A
    @0006   // @R6
    M=M-D
    // TODO: for now maybe just assume valid 1-8 char follows? 0x31 - 0x38
    @0001   // @R1
    AM=M+1
    D=M
    @0030   // maps [0x31, 0x38] -> [1, 8] 
    D=D-A
    D=D<<7
    // fall through to jump
// (JUMP) 401
    // whenever we come to jump, we have just set D=bits to add to instr
    // so the first thing we do is to add those to R6
    @0006   // @R6
    M=M|D
    // parse ENTER or ;J then two letter combo. set jump bits accordingly
    // TODO: OR some whitespace OR goto COMMENT
    // TODO: copy from ENDLINE, should be a function using stackpointer?
    // (PREJUMP) 403
    @0001   // @R1
    AM=M+1
    D=M
    // if D==0x20 (SPACE) goto PREJUMP, allowing trailing spaces
    @0020   // ascii SPACE
    D=D-A
    @0193   // @PREJUMP
    D;JEQ
    // if D==0x2F (/) goto COMMENT
    @000F   // ascii / - ascii SPACE
    D=D-A
    @01FC   // @COMMENT
    D;JEQ
    // if D==0x80 (ENTER) goto WRITE
    @0051   // ascii ENTER - ascii /
    D=D-A
    @0223   // @WRITE
    D;JEQ
    // otherwise parse the jump instruction part
    @0045   // ascii ENTER - ascii ;
    D=D+A
    @0233   // @END // TODO syntax error
    D;JNE
    @0001   // @R1
    AM=M+1
    D=M
    @004A   // ascii J
    D=D-A
    @0233   // @END // TODO syntax error
    D;JNE
    // read the next two chars as 8bit values into D then compare
    @0001   // @R1
    AM=M+1
    D=M<<8
    @0001   // @R1
    AM=M+1
    D=D|M
    @4551   // ascii EQ = 0x4551
    D=D-A
    @01BB   // @JGE
    D;JNE
    @0002   // JEQ jump bits
    D=A
    @01E9   // @ENDJUMP
    0;JMP
// (JGE) 443
    @01F4   // ascii GE = 0x4745, 0x01F4 more than EQ
    D=D-A
    @01C3   // @JGT
    D;JNE
    @0003   // JGE jump bits
    D=A
    @01E9   // @ENDJUMP
    0;JMP
// (JGT) 451
    @000F   // ascii GT = 0x4754, 0x000F more than GE
    D=D-A
    @01CB   // @JLE
    D;JNE
    @0001   // JGT jump bits
    D=A
    @01E9   // @ENDJUMP
    0;JMP
// (JLE) 459
    @04F1   // ascii LE = 0x4C45, 0x04F1 more than GT
    D=D-A
    @01D3   // @JLT
    D;JNE
    @0006   // JLE jump bits
    D=A
    @01E9   // @ENDJUMP
    0;JMP
// (JLT) 467
    @000F   // ascii LT = 0x4C54, 0x000F more than LE
    D=D-A
    @01DB   // @JMP
    D;JNE
    @0004   // JLT jump bits
    D=A
    @01E9   // @ENDJUMP
    0;JMP
// (JMP) 475
    @00FC   // ascii MP = 0x4D50, 0x00FC more than LT
    D=D-A
    @01E3   // @JNE
    D;JNE
    @0007   // JMP jump bits
    D=A
    @01E9   // @ENDJUMP
    0;JMP
// (JNE) 483
    @00F5   // ascii NE = 0x4E45, 0x00F5 more than MP
    D=D-A
    @0233   // @END // TODO syntax error
    D;JNE
    @0005   // JNE jump bits
    D=A
    // fall through to ENDJUMP
// (ENDJUMP) 489
    // whenever we come to endjump, we have just set D=bits to add to instr
    // so the first thing we do is to add those to R6
    @0006   // @R6
    M=M|D
    // consume trailing whitespace/comment and goto write
// (ENDLINE) 491
    @0001   // @R1
    AM=M+1
    D=M
    // if D==0x20 (SPACE) goto ENDLINE, allowing trailing spaces
    @0020   // ascii SPACE
    D=D-A
    @01EB   // @ENDLINE
    D;JEQ
    // if D==0x2F (/) goto COMMENT
    @000F   // ascii / - ascii SPACE
    D=D-A
    @01FC   // @COMMENT
    D;JEQ
    // if D==0x80 (ENTER) goto WRITE
    @0051   // ascii ENTER - ascii /
    D=D-A
    @0223   // @WRITE
    D;JEQ
    @0233   // @END // TODO syntax error
    0;JMP
// (COMMENT) 508 -> consume second slash
    @0001   // @R1
    AM=M+1
    D=M
    // if D!=0x2F (/) goto END TODO syntax error
    @002F   // ascii /
    D=D-A
    @0233   // @END
    D;JNE
// (COMMENTREC) 515 -> consume rest of the line
    @0001   // @R1
    AM=M+1
    D=M
    // if D==0x80 (ENTER) goto WRITE 
    @0080   // ascii ENTER
    D=D-A
    @0223   // @WRITE
    D;JEQ
    // else goto COMMENTREC
    @0203   // @COMMENTREC
    0;JMP
// (AINSTR) 524 -> parse rest as hex value
    // TODO: assumes max 4 valid hex chars follow!
    @0001   // @R1
    AM=M+1
    D=M
    // if D==0x80 (ENTER) goto WRITE 
    @0080   // ascii ENTER
    D=D-A
    @0223   // @WRITE
    D;JEQ
    // TODO: if not 0-9 or A-F, goto END
    @0080   // otherwise set D back to read value
    D=D+A
    @0006   // @R6
    M=M<<4
    @0041   // ascii A
    D=D-A
    @021D   // @ALPHANUM 
    D;JGE   // if D-65 >= 0, we have a A-F char
    // only for digits: add back to map 0-9A-F continuous
    @0007   // ascii A - ascii 9 - 1
    D=D+A
// (ALPHANUM) 541
    // now [0,9] -> [-10,-1] and [A-F] -> [0,5]
    @000A   // 10
    D=D+A
    @0006   // @R6
    M=M|D
    @020C   // @AINSTR
    0;JMP   // goto AINSTR
// (WRITE) 547 write instruction to mem
    // R6 should contain the instruction now
    @0006   // @R6
    D=M
    // if R6 is 0 then this was a full line comment, do not write anything
    @022B   // @RESET
    D;JEQ
    // otherwise write to mem at R2 and advance counter
    @0002   // @R2
    M=M+1
    A=M-1
    M=D
// (RESET) 555 consume newline (assume found, otherwise shouldve been syntax error) and parse next line
    @0006   // @R6
    M=0     // R6 = 0
    @0007   // @R7
    M=0     // R7 = 0
    @0001   // @R1
    M=M+1
    @0010   // @START
    0;JMP
// (END) 563
    @0233   // @END
    0;JMP   // goto END
