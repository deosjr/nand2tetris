// FULL HACK ASSEMBLER ACCORDING TO SPEC
// with the following exceptions/additions:
// - LABELS are always in uppercase
// - variables are always in lowercase
// - the << bitshift operator is added to the language
// supported arguments are 1 through 8 inclusive
// - the pc jump (read from memory) instruction is NOT supported (yet?)
// - TODO: AINSTR decimal constants
    // init vars
    @0010
    D=A
    @SP
    M=D     // R0 = 0x10
    @R6
    M=0     // R6 = 0
    @R7
    M=0     // R7 = 0
    // R8 (label->line mappings) start at 0x30
    // but we predefine a few (such as SP and ARG) at the start
// PREDEFINE SP, LCL, ARG, THIS, THAT, SCREEN, KBD
    // TODO: SCREEN and KBD dont match the constraint that linenumbers cap at 0x0FFF
    // R0-R15 are handled separately in AINSTR TODO R10-R15
    @5350   // SP in ascii
    D=A
    @0030
    M=D     // 0x30 = SP, 0x31 = 0x0
    @4C43
    D=A
    @0032
    M=D
    @4C00
    D=A
    @0033
    M=D
    @0001
    D=A
    @0034
    M=D     // 0x32-0x33 = LCL, 0x34 = 0x1
    @4152
    D=A
    @0035
    M=D
    @4700
    D=A
    @0036
    M=D
    @0002
    D=A
    @0037
    M=D     // 0x35-0x36 = ARG, 0x37 = 0x2
    @5448
    D=A
    @0038
    M=D
    @4953
    D=A
    @0039
    M=D
    @0003
    D=A
    @003A
    M=D     // 0x38-0x39 = THIS, 0x3A = 0x3
    @5448
    D=A
    @003B
    M=D
    @4154
    D=A
    @003C
    M=D
    @0004
    D=A
    @003D
    M=D     // 0x3B-0x3C = THAT, 0x3D = 0x4
    @003E
    D=A
    @R8
    M=D     // R8 = 0x30 + offset of predefined symbols
(FIRSTPASS)
    @6001
    DM=M
    @001C   // EOF marker
    D=D-A
    @SECONDPASS
    D;JEQ
    @0004   // ascii SPACE - ascii EOF
    D=D-A
    @FIRSTPASS
    D;JEQ
    @0008   // ascii ( - ascii SPACE
    D=D-A
    @LABEL
    D;JEQ
    @0007   // ascii / - ascii (
    D=D-A
    @SKIPLINE
    D;JEQ
    @R6
    M=M+1   // R6 holds the current linenumber skipping comments/labels
(SKIPLINE)
    @6001
    DM=M
    @0080   // ascii ENTER
    D=D-A
    @SKIPLINE
    D;JNE
    @FIRSTPASS
    0;JMP
(LABEL)
    // here R7 is used to count length of label
    @6001
    D=M
    @0029   // ascii )
    D=D-A
    @ENDLABEL
    D;JEQ
    @R7
    DM=M+1
    @0001   // @1
    D=D&A   // mask, D=0 if R7 is even and 1 if R7 is uneven
    @EVEN
    D;JEQ
// ODD
    @6001
    DM=M<<8
    @R8
    A=M
    M=D
    @LABEL
    0;JMP
(EVEN)
    @6001
    DM=M
    @R8
    M=M+1
    A=M-1
    M=D|M
    @LABEL
    0;JMP
(ENDLABEL)
    // last found label should be substituted with current value of R6 in next pass
    @R7
    D=M
    @0001   // @1
    D=D&A   // same mask (un)even check
    @LINENUMBER
    D;JEQ
    @R8
    M=M+1
(LINENUMBER)
    @R6
    D=M
    @R8
    M=M+1
    A=M-1
    M=D
    @R7
    M=0   
    @SKIPLINE
    0;JMP
(SECONDPASS)
    // init vars
    @R6
    M=0     // R6 = 0
    @0010
    D=A
    @R9
    M=D     // R9 = 0x10
    @R8
    D=M
    @R3
    M=D     // R3 is the end of label list pointing at empty value
(START)
    @1000
    D=A
    @R1
    M=D     // R1 = 0x1000
    // read line from tape into memory starting at R1 0x1000
(READLINE)
    @6001
    DM=M
    @R1
    M=M+1
    A=M-1
    M=D
    @0080   // ascii ENTER
    D=D-A
    @READLINE
    D;JNE
    @1000
    D=A
    @R1
    M=D     // R1 = 0x1000
(OLDSTART)
    // read char
    @R1
    A=M
    D=M
    // if D==0 goto END 
    @END
    D;JEQ
    // if D==0x20 (SPACE) goto START, allowing leading spaces (indents)
    @0020   // ascii SPACE
    D=D-A
    @STARTLINE
    D;JNE
    @R1
    M=M+1
    @OLDSTART
    0;JMP
(STARTLINE)
    // if D==0x28 '(' goto COMMENTREC
    @0008   // ascii ( - ascii SPACE
    D=D-A
    @COMMENTREC
    D;JEQ
    // if D==0x2F (/) goto COMMENT
    @0007   // ascii / - ascii (
    D=D-A
    @COMMENT
    D;JEQ
    // if D==0x40 (@) goto AINSTR
    @0011   // ascii @ - ascii /
    D=D-A
    @AINSTR
    D;JEQ
    // else LOOKAHEAD for start C instr
    // R6 here to build up our instruction, setting individual bits
    @7000   // 0111 0000 0000 0000
    D=A
    @R6
    M=D     // R6 = 0x7000
    M=M<<1  // R6 = 0xE000
(LOOKAHEAD)
    // idea here is to find out whether we need to parse a destination or not
    // consume A/M/D until we find something else, then switch on whether its an =
    // NOTE we already consumed a first token! therefore we start by incr R7 to 1
    @R7
    DM=M+1
    @R1
    A=D+M
    D=M
    // if D==A goto LOOKAHEAD 
    @0041   // ascii A
    D=D-A
    @LOOKAHEAD
    D;JEQ
    // if D==D goto LOOKAHEAD 
    @0003   // ascii D - ascii A
    D=D-A
    @LOOKAHEAD
    D;JEQ
    // if D==M goto LOOKAHEAD 
    @0009   // ascii M - ascii D
    D=D-A
    @LOOKAHEAD
    D;JEQ
    // else if we read = goto DEST otherwise goto COMP
    @0010   // ascii M - ascii =
    D=D+A
    @COMP
    D;JNE
// DEST
    // if A/M/D set dest bits accordingly, loop
    @R1
    M=M-1   // because we want to start the loop with incr
(DESTA)
    // We know that there is an = sign at R1+R7 but checking each character
    // again allows for syntax errors to be detected
    // TODO: this will allow duplicate dest, ie AAAM=D+1
    @R1
    AM=M+1
    D=M
    @0041   // ascii A
    D=D-A
    @DESTD
    D;JNE
    // D==A
    @0020   // @A as dest bit
    D=A
    @R6
    M=D|M
    @DESTA
    0;JMP
(DESTD)
    @0003   // ascii D - ascii A
    D=D-A
    @DESTM
    D;JNE
    // D==D
    @0010   // @D as dest bit
    D=A
    @R6
    M=D|M
    @DESTA
    0;JMP
(DESTM)
    @0009   // ascii M - ascii D
    D=D-A
    @DESTEQ
    D;JNE
    // D==M
    @0008   // @M as dest bit
    D=A
    @R6
    M=D|M
    @DESTA
    0;JMP
(DESTEQ)
    // if equals sign, goto COMP
    @0010   // ascii M - ascii =
    D=D+A
    @R1
    M=M+1
    @COMP
    D;JEQ
    @END   // TODO syntax error
    0;JMP
(COMP)
    // lookahead one character for an operator: + - & | <
    // if operator, goto binary, otherwise parse unary comp
    @R1
    A=M+1
    D=M
    @0026   // ascii &
    D=D-A
    @BINARY
    D;JEQ
    @0005   // ascii + - ascii &
    D=D-A
    @BINARY
    D;JEQ
    @0002   // ascii - - ascii +
    D=D-A
    @BINARY
    D;JEQ
    @000F   // ascii < - ascii -
    D=D-A
    @BINARY
    D;JEQ
    @0040   // ascii | - ascii <
    D=D-A
    @BINARY
    D;JEQ
// UNARY
    @R1
    A=M
    D=M
    @0021   // ascii !
    D=D-A
    @NOT
    D;JEQ
    @000C   // ascii - - ascii !
    D=D-A
    @NEG
    D;JEQ
    @0003   // ascii 0 - ascii -
    D=D-A
    @ONE
    D;JNE
// ZERO
    @0A80   // 0 comp bits
    D=A
    @ENDCOMP
    0;JMP
(ONE)
    D=D-1   // ascii 1 - ascii 0
    @UNRYA
    D;JNE
    @0FC0   // 1 comp bits
    D=A
    @ENDCOMP
    0;JMP
(UNRYA)
    @0010   // ascii A - ascii 1
    D=D-A
    @UNRYD
    D;JNE
    @0C00   // A comp bits
    D=A
    @ENDCOMP
    0;JMP
(UNRYD)
    @0003   // ascii D - ascii A
    D=D-A
    @UNRYM
    D;JNE
    @0300   // D comp bits
    D=A
    @ENDCOMP
    0;JMP
(UNRYM)
    @0009   // ascii M - ascii D
    D=D-A
    @END   // TODO: syntax error
    D;JNE
    @1C00   // M comp bits
    D=A
    @ENDCOMP
    0;JMP
(NOT)
    @R1
    AM=M+1
    D=M
    @0041   // ascii A
    D=D-A
    @NOTD
    D;JNE
    @0C40   // !A comp bits
    D=A
    @ENDCOMP
    0;JMP
(NOTD)
    @0003   // ascii D - ascii A
    D=D-A
    @NOTM
    D;JNE
    @0340   // !D comp bits
    D=A
    @ENDCOMP
    0;JMP
(NOTM)
    @0009   // ascii M - ascii D
    D=D-A
    @END   // TODO: syntax error
    D;JNE
    @1C40   // !M comp bits
    D=A
    @ENDCOMP
    0;JMP
(NEG)
    @R1
    AM=M+1
    D=M
    @0031   // ascii 1
    D=D-A
    @NEGA
    D;JNE
    @0E80   // -1 comp bits
    D=A
    @ENDCOMP
    0;JMP
(NEGA)
    @0010   // ascii A - ascii 1
    D=D-A
    @NEGD
    D;JNE
    @0CC0   // -A comp bits
    D=A
    @ENDCOMP
    0;JMP
(NEGD)
    @0003   // ascii D - ascii A
    D=D-A
    @NEGM
    D;JNE
    @03C0   // -D comp bits
    D=A
    @ENDCOMP
    0;JMP
(NEGM)
    @0009   // ascii M - ascii D
    D=D-A
    @END   // TODO: syntax error
    D;JNE
    @1CC0   // -M comp bits
    D=A
    @ENDCOMP
    0;JMP
(BINARY)
    @R1
    A=M
    D=M
    @0041   // ascii A
    D=D-A
    @BNRYA
    D;JEQ
    @0003   // ascii D - ascii A
    D=D-A
    @BNRYD
    D;JEQ
    @0009   // ascii M - ascii D
    D=D-A
    @BNRYM
    D;JEQ
    @END   // TODO: syntax error
    0;JMP
(BNRYM)
    // set the M flag, then fall through to A case
    @1000   // 0001 0000 0000 0000
    D=A
    @R6
    M=D|M
(BNRYA)
    // read the next two chars as 8bit values into D then compare
    @R1
    AM=M+1
    D=M<<8
    @R1
    AM=M+1
    D=D|M
    @2B31   // ascii +1 = 0x2B31
    D=D-A
    @MINONE
    D;JNE
    @0DC0   // A/M +1 comp bits
    D=A
    @ENDCOMP
    0;JMP
(MINONE)
    @0200   // ascii -1 = 0x2D31, 0x0200 more than +1
    D=D-A
    @MIND
    D;JNE
    @0C80   // A/M -1 comp bits
    D=A
    @ENDCOMP
    0;JMP
(MIND)
    @0013   // ascii -D = 0x2D44, 0x0013 more than -1
    D=D-A
    @ASHIFT
    D;JNE
    @01C0   // A/M -D comp bits
    D=A
    @ENDCOMP
    0;JMP
(ASHIFT)
    @0EF8   // ascii << = 0x3C3C, 0x0EF8 more than -D
    D=D-A
    @SHIFT
    D;JEQ
    @END   // TODO: syntax error
    0;JMP
(BNRYD)
    // read the next two chars as 8bit values into D then compare
    @R1
    AM=M+1
    D=M<<8
    @R1
    AM=M+1
    D=D|M
    @2641   // ascii &A = 0x2641
    D=D-A
    @ANDM
    D;JNE
    // D&A comp bits are all 0
    D=0
    @ENDCOMP
    0;JMP
(ANDM)
    @000C   // ascii &M = 0x264D, 0x000C more than &A
    D=D-A
    @PLUSONE
    D;JNE
    @1000   // D&M comp bits
    D=A
    @ENDCOMP
    0;JMP
(PLUSONE)
    @04E4   // ascii +1 = 0x2B31, 0x04E4 more than &M
    D=D-A
    @PLUSA
    D;JNE
    @07C0   // D+1 comp bits
    D=A
    @ENDCOMP
    0;JMP
(PLUSA)
    @0010   // ascii +A = 0x2B41, 0x0010 more than +1
    D=D-A
    @PLUSM
    D;JNE
    @0080   // D+A comp bits
    D=A
    @ENDCOMP
    0;JMP
(PLUSM)
    @000C   // ascii +M = 0x2B4D, 0x000C more than +A
    D=D-A
    @DMINONE
    D;JNE
    @1080   // D+M comp bits
    D=A
    @ENDCOMP
    0;JMP
(DMINONE)
    @01E4   // ascii -1 = 0x2D31, 0x01E4 more than +M
    D=D-A
    @MINA
    D;JNE
    @0380   // D-1 comp bits
    D=A
    @ENDCOMP
    0;JMP
(MINA)
    @0010   // ascii -A = 0x2D41, 0x0010 more than -1
    D=D-A
    @MINM
    D;JNE
    @04C0   // D-A comp bits
    D=A
    @ENDCOMP
    0;JMP
(MINM)
    @000C   // ascii -M = 0x2D4D, 0x000C more than -A
    D=D-A
    @ORA
    D;JNE
    @14C0   // D-M comp bits
    D=A
    @ENDCOMP
    0;JMP
(ORA)
    @4EF4   // ascii |A = 0x7C41, 0x4EF4 more than -M
    D=D-A
    @ORM
    D;JNE
    @0540   // D|A comp bits
    D=A
    @ENDCOMP
    0;JMP
(ORM)
    @000C   // ascii |M = 0x7C4D, 0x000C more than |A
    D=D-A
    @DSHIFT
    D;JNE
    @1540   // D|M comp bits
    D=A
    @ENDCOMP
    0;JMP
(DSHIFT)
    @4011   // ascii << = 0x3C3C, 0x4011 LESS than |M
    D=D+A
    @END   // TODO syntax error
    D;JNE
    @0800   // D bit for the shift operation
    D=A
    @R6
    M=D|M
    // fall through to SHIFT
(SHIFT)
    // We can come here from either A, D or M parsed, flags have been set accordingly
    // set the 3rd bit to 0 again using AND mask (but cant use highest bit in A instr!)
    // so instead we simply subtract 0x2000 (since we know that bit is set to 1 earlier)
    @2000   // 0010 0000 0000 0000
    D=A
    @R6
    M=M-D
    // TODO: for now maybe just assume valid 1-8 char follows? 0x31 - 0x38
    @R1
    AM=M+1
    D=M
    @0030   // maps [0x31, 0x38] -> [1, 8] 
    D=D-A
    D=D<<7
    // fall through to endcomp
(ENDCOMP)
    // whenever we come to jump, we have just set D=bits to add to instr
    // so the first thing we do is to add those to R6
    @R6
    M=D|M
    // check whether we stop early in ENDLINE func
    @JUMP
    D=A
    @SP
    AM=M+1
    M=D
    @ENDLINE
    0;JMP
    // otherwise parse the jump instruction part
(JUMP)
    // parse ENTER or ;J then two letter combo. set jump bits accordingly
    @0045   // ascii ENTER - ascii ;
    D=D+A
    @END   // TODO syntax error
    D;JNE
    @R1
    AM=M+1
    D=M
    @004A   // ascii J
    D=D-A
    @END   // TODO syntax error
    D;JNE
    // read the next two chars as 8bit values into D then compare
    @R1
    AM=M+1
    D=M<<8
    @R1
    AM=M+1
    D=D|M
    @4551   // ascii EQ = 0x4551
    D=D-A
    @JGE
    D;JNE
    @0002   // JEQ jump bits
    D=A
    @ENDJUMP
    0;JMP
(JGE)
    @01F4   // ascii GE = 0x4745, 0x01F4 more than EQ
    D=D-A
    @JGT
    D;JNE
    @0003   // JGE jump bits
    D=A
    @ENDJUMP
    0;JMP
(JGT)
    @000F   // ascii GT = 0x4754, 0x000F more than GE
    D=D-A
    @JLE
    D;JNE
    @0001   // JGT jump bits
    D=A
    @ENDJUMP
    0;JMP
(JLE)
    @04F1   // ascii LE = 0x4C45, 0x04F1 more than GT
    D=D-A
    @JLT
    D;JNE
    @0006   // JLE jump bits
    D=A
    @ENDJUMP
    0;JMP
(JLT)
    @000F   // ascii LT = 0x4C54, 0x000F more than LE
    D=D-A
    @JMP
    D;JNE
    @0004   // JLT jump bits
    D=A
    @ENDJUMP
    0;JMP
(JMP)
    @00FC   // ascii MP = 0x4D50, 0x00FC more than LT
    D=D-A
    @JNE
    D;JNE
    @0007   // JMP jump bits
    D=A
    @ENDJUMP
    0;JMP
(JNE)
    @00F5   // ascii NE = 0x4E45, 0x00F5 more than MP
    D=D-A
    @END   // TODO syntax error
    D;JNE
    @0005   // JNE jump bits
    D=A
    // fall through to ENDJUMP
(ENDJUMP)
    // whenever we come to endjump, we have just set D=bits to add to instr
    // so the first thing we do is to add those to R6
    @R6
    M=D|M
    // check whether we stop in ENDLINE func
    // if we do not, we return to syntax error
    @END
    D=A
    @SP
    AM=M+1
    M=D
    @ENDLINE
    0;JMP
(COMMENT)
    // consume second slash
    @R1
    AM=M+1
    D=M
    // if D!=0x2F (/) goto END TODO syntax error
    @002F   // ascii /
    D=D-A
    @END
    D;JNE
(COMMENTREC)
    // consume rest of the line
    @R1
    AM=M+1
    D=M
    // if D==0x80 (ENTER) goto WRITE 
    @0080   // ascii ENTER
    D=D-A
    @WRITE
    D;JEQ
    // else goto COMMENTREC
    @COMMENTREC
    0;JMP
(AINSTR)
    // parse rest as hex value,label OR variable
    // TODO: decimal numbers as default, hexvalues starting with 0x, labels in allcaps, or variables in lowercase
    @R1
    A=M+1
    D=M     // lookahead for digit (hex) uppercase (label) or lowercase (var) letter as first value
    @0041   // ascii A
    D=D-A
    @AHEX
    D;JLT
    @0020   // 0x20
    D=A
    @R7
    M=D     // R7 = 0x20
    @0030   // 0x30
    D=A
    @R8
    M=D     // R8 = 0x30
    @R1
    A=M+1
    D=M
    @0061   // ascii a
    D=D-A
    @VAR
    D;JGE
(ALABEL)
    // labels should have at least length two; abuse to parse R0-R15
    // first has to be an uppercase letter, if its an R then second can be a digit
    @R1
    AM=M+1
    D=M
    @0041   // ascii A
    D=D-A
    @ENDALABEL
    D;JLT
    @0019   // ascii Z - ascii A
    D=D-A
    @ENDALABEL
    D;JGT
    // if R7=0x20 we have parsed the first letter; if it is an R
    // then we can first try to parse R0-R15
    @R7
    D=M<<8
    @R1
    A=M
    D=D|M   // D=R7R1
    @2052   // 0x20 followed by ascii R
    D=D-A
    @RDIGIT
    D;JEQ
(BACKFROMRDIGIT)
    @R1
    A=M
    D=M<<8
    @R7
    A=M
    M=D
    @R1
    AM=M+1
    D=M
    @0041   // ascii A
    D=D-A
    @UNENDALABEL
    D;JLT
    @0019   // ascii Z - ascii A
    D=D-A
    @UNENDALABEL
    D;JGT
    @R1
    A=M
    D=M
    @R7
    M=M+1
    A=M-1
    M=D|M
    @ALABEL
    0;JMP
(RDIGIT)
    // we have confirmed the first letter of label is an R
    // try parsing a digit next, otherwise return to ALABEL loop
    @R1
    A=M+1
    D=M
    @0030   // ascii 0
    D=D-A
    @BACKFROMRDIGIT
    D;JLT
    @0009   // ascii 9 - ascii 0
    D=D-A
    @BACKFROMRDIGIT
    D;JGT
    // we have found a digit!
    // TODO: only parses R0-R9 atm, R10-R15 unsupported
    @R1
    AM=M+1
    D=M
    @0030   // ascii 0
    D=D-A
    @R6
    M=D
    @END
    D=A
    @SP
    AM=M+1
    M=D
    @ENDLINE
    0;JMP
(VAR)
    @R1
    AM=M+1
    D=M
    @0061   // ascii a
    D=D-A
    @ENDALABEL
    D;JLT
    @0019   // ascii z - ascii a
    D=D-A
    @ENDALABEL
    D;JGT
    @R1
    A=M
    D=M<<8
    @R7
    A=M
    M=D
    @R1
    AM=M+1
    D=M
    @0061   // ascii a
    D=D-A
    @UNENDALABEL
    D;JLT
    @0019   // ascii z - ascii a
    D=D-A
    @UNENDALABEL
    D;JGT
    @R1
    A=M
    D=M
    @R7
    M=M+1
    A=M-1
    M=D|M
    @VAR
    0;JMP
(UNENDALABEL)
    // if label is of uneven length we need to add 1 more to R7 first
    @R7
    M=M+1
(ENDALABEL)
    // label has been stored between 0x20 and 0x30. Now lookup the line number after 0x30
    @7FFF   // marker for end of read label, largest value we can store in A
    D=A
    @R7
    A=M
    M=D
(REPEATFIND)
    @0020   // 0x20
    D=A
    @R7
    M=D     // R7 = 0x20
(FINDALABEL)
    // if the value at R7 is 0x7FFF, we found our label unless we found a prefix
    @R7
    A=M
    D=M
    @7FFF   // 0x7FFF
    D=D-A
    @FOUNDALABEL
    D;JEQ
    // otherwise, compare R7 with R8
    @R7
    A=M
    D=M
    @R8
    A=M
    D=D-M
    @NOMATCH
    D;JNE
// MATCH
    // if equal, advance both R7 and R8
    @R7
    M=M+1
    @R8
    M=M+1
    @FINDALABEL
    0;JMP
(NOMATCH)
    // if R8 is too big, we give up on a match. This is a syntax error for labels
    // and results in adding a new var name for variables
    @R8
    D=M
    @R3     // pointer to end of label/var list
    D=D-M
    @NOMATCHREC
    D;JLT
    @0020   // 0x20
    D=M<<8
    @00FF   // 0000 0000 1111 1111
    D=D&A   // D is first letter of label/var
    @005A   // ascii Z
    D=D-A
    @END    // TODO syntax error
    D;JLE
    @0020   // 0x20
    D=A
    @R7
    M=D
(WRITEVAR)
    @R7
    A=M
    D=M
    @7FFF
    D=D-A
    @DONEVAR
    D;JEQ
    @R7
    M=M+1
    A=M-1
    D=M
    @R8
    M=M+1
    A=M-1
    M=D
    @R3
    M=M+1
    @WRITEVAR
    0;JMP
(DONEVAR)
    @R3
    M=M+1
    @R9
    M=M+1
    D=M-1
    @R8
    A=M
    M=D
    @R6
    M=D
    @R1
    M=M-1
    @END
    D=A
    @SP
    AM=M+1
    M=D
    @ENDLINE
    0;JMP
(NOMATCHREC)
    // otherwise, set R7 back to 0x20 and advance R8 past next value starting with 0
    // NOTE: this means label line values can only go up to 0x0FFF !
    @R8
    M=M+1
    A=M-1
    D=M
    @7000   // 0111 0000 0000 0000
    D=D&A
    @REPEATFIND
    D;JEQ
    @NOMATCH
    0;JMP
(FOUNDALABEL)
    // if R8 starts with 0 were good otherwise we found a prefix
    @R8
    A=M
    D=M
    @7000   // 0111 0000 0000 0000
    D=D&A
    @NOMATCH
    D;JNE
    @R8
    A=M
    D=M
    @R6
    M=D
    @R1
    M=M-1
    @END
    D=A
    @SP
    AM=M+1
    M=D
    @ENDLINE
    0;JMP
(AHEX)
    // TODO: assumes max 4 valid hex chars follow!
    @R1
    AM=M+1
    D=M
    // if D==0x80 (ENTER) goto WRITE 
    @0080   // ascii ENTER
    D=D-A
    @WRITE
    D;JEQ
    // TODO: if not 0-9 or A-F, goto END
    @0080   // otherwise set D back to read value
    D=D+A
    @R6
    M=M<<4
    @0041   // ascii A
    D=D-A
    @ALPHANUM
    D;JGE   // if D-65 >= 0, we have a A-F char
    // only for digits: add back to map 0-9A-F continuous
    @0007   // ascii A - ascii 9 - 1
    D=D+A
(ALPHANUM)
    // now [0,9] -> [-10,-1] and [A-F] -> [0,5]
    @000A   // 10
    D=D+A
    @R6
    M=D|M
    @R7
    DM=M+1  // R7+=1
    @0004   // @4
    D=D-A
    @AHEX
    D;JNE
    // after parsing 4 hex chars, check whether we stop in ENDLINE func
    // if we do not, we return to syntax error
    @END
    D=A
    @SP
    AM=M+1
    M=D
    @ENDLINE
    0;JMP
(ENDLINE)
    // function called from end of comp, jump and ainstr
    // consumes trailing spaces, comments and newline
    // jumps back if it FAILS to find any of those
    // otherwise cuts parsing the line short and never returns
    @SP
    M=M-1   // whatever happens, decrement stack pointer
(ENDLINELOOP)
    @R1
    AM=M+1
    D=M
    // if D==0x20 (SPACE) goto ENDLINELOOP, allowing trailing spaces
    @0020   // ascii SPACE
    D=D-A
    @ENDLINELOOP
    D;JEQ
    // if D==0x2F (/) goto COMMENT
    @000F   // ascii / - ascii SPACE
    D=D-A
    @COMMENT
    D;JEQ
    // if D==0x80 (ENTER) goto WRITE
    @0051   // ascii ENTER - ascii /
    D=D-A
    @WRITE
    D;JEQ
    // goto the address @R0 pointed to
    @SP
    A=M+1   // already decremented stack pointer
    A=M
    0;JMP   // return to caller
(WRITE)
    // write instruction to mem
    // R6 should contain the instruction now
    @R6
    D=M
    @R7
    D=D+M
    // if R6 is 0 and R7 is 0 then this was a full line comment, do not write anything
    @RESET
    D;JEQ
    // otherwise write to output tape register
    @R6
    D=M
    @6002 // output register
    M=D
(RESET)
    @6001
    D=M
    @001C
    D=D-A
    @END
    D;JEQ
    // consume newline (assume found, otherwise shouldve been syntax error) and parse next line
    @R6
    M=0     // R6 = 0
    @R7
    M=0     // R7 = 0
(CLEARLINE)
    // clear read line back to 0x0 in order to read new line
    @R1
    A=M
    M=0
    D=A
    @1000
    D=D-A
    @START
    D;JEQ
    @R1
    M=M-1
    @CLEARLINE
    0;JMP
(END)
    @END
    0;JMP   // goto END
