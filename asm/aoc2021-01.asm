// Advent of Code 2021 - Day 1 part 1
    @0020
    D=A
    @SP
    M=D
    @READFIRST
    D=A
    @SP
    AM=M+1
    M=D
    @DECIMAL
    0;JMP
(READFIRST)
    @decimal
    D=M
    @prev
    M=D
(PARTONE)
    @READNEXT
    D=A
    @SP
    AM=M+1
    M=D
    @DECIMAL
    0;JMP
(READNEXT)
    @decimal
    D=M
    @prev
    D=D-M
    @SETPREV
    D;JLE
    @total
    M=M+1
(SETPREV)
    @decimal
    D=M
    @prev
    M=D
    @6001
    D=M // read without clearing register so we keep this value; peek
    @001C // EOF
    D=D-A
    @WRITEONE
    D;JEQ
    @PARTONE
    0;JMP
(DECIMAL)
    @decimal
    M=0
(READ)
    @6001
    DM=M // assign to M in order to clear the read register, reading next char
    @0080 // ENTER (comes before EOF)
    D=D-A
    @ENDREAD
    D;JEQ
    @0050 // ascii ENTER - ascii 0
    D=D+A
    @END  // syntax error
    D;JLT
    @0009
    D=D-A
    @END  // syntax error
    D;JGT
    @0009
    D=D+A
    @temp
    M=D
    @decimal
    D=M
    M=M<<3 // decimal = decimal * 8
    M=D+M
    M=D+M  // decimal = decimal * 8 + decimal + decimal = decimal * 10
    @temp
    D=M
    @decimal
    M=D+M
    @READ
    0;JMP
(ENDREAD)
    // goto the address @SP pointed to
    @SP
    M=M-1
    A=M+1
    A=M
    0;JMP   // return to caller
(WRITEONE)
    @total
    D=M
    @6002
    M=D
(END)
    @END
    0;JMP
