    @decimal
    M=0
(READ)
    @6001
    DM=M // assign to M in order to clear the read register, reading next char
    @0080 // ENTER (comes before EOF)
    D=D-A
    @WRITE
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
(WRITE)
    @decimal
    D=M
    @6002
    M=D
(END)
    @END
    0;JMP
