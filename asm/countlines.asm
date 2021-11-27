// COUNT LINES
// step from firstpass to twopass assembler, helps counting lines
// but does nothing else. expects labels to be uncommented
    // init vars
    @0010
    D=A
    @0000 // @R0
    M=D   // R0 = 0x10
    @1000
    D=A
    @0001 // @R1
    M=D   // R1 = 0x1000
    @1000 // TODO not enough space! need to overwrite input
    D=A
    @0002 // @R2
    M=D   // R2 = 0x1000
    @0006 // @R6
    M=0   // R6 = 0
    @0007 // @R7
    M=0   // R7 = 0
    @0020
    D=A
    @0008 // @R8
    M=D   // R8 = 0x20
(FIRSTPASS) 20
    @0001   // @R1
    A=M
    D=M
    // if D=0x28 '(' goto LABEL else consume until new line
    @005E   // @END
    D;JEQ
    @0028   // ascii (
    D=D-A
    @002E   // @LABEL
    D;JEQ
    @0007   // ascii / - ascii (
    D=D-A
    @0023   // @SKIPLINE
    D;JEQ
    @0006   // @R6
    M=M+1   // R6 holds the current linenumber skipping comments/labels
(SKIPLINE) 35
    @0001   // @R1
    AM=M+1
    D=M
    @0080   // ascii ENTER
    D=D-A
    @0023   // @SKIPLINE
    D;JNE
    @0001   // @R1
    M=M+1
    @0014   // @FIRSTPASS
    0;JMP
(LABEL) 46
    // here R7 is used to count length of label
    @0001   // @R1
    AM=M+1
    D=M
    @0029   // ascii )
    D=D-A
    @004C   // @ENDLABEL
    D;JEQ
    @0007   // @R7
    DM=M+1
    @0001   // @1
    D=D&A   // mask, D=0 if R7 is even and 1 if R7 is uneven
    @0043   // @EVEN
    D;JEQ
// ODD
    @0001   // @R1
    A=M
    D=M<<8
    @0008   // @R8
    A=M
    M=D
    @002E   // @LABEL
    0;JMP
(EVEN) 67
    @0001   // @R1
    A=M
    D=M
    @0008   // @R8
    M=M+1
    A=M-1
    M=D|M
    @002E   // @LABEL
    0;JMP
(ENDLABEL) 76
    // last found label should be substituted with current value of R6 in next pass
    @0007   // @R7
    D=M
    @0001   // @1
    D=D&A   // same mask (un)even check
    @0054   // @LINENUMBER
    D;JEQ
    @0008   // @R8
    M=M+1
(LINENUMBER) 84
    @0006   // @R6
    D=M
    @0008   // @R8
    M=M+1
    A=M-1
    M=D
    @0007   // @R7
    M=0   
    @0023   // @SKIPLINE
    0;JMP
(END) 94
    @005E   // @END
    0;JMP
