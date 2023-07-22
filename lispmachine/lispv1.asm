    @0x48
    D=A
    @0
    SETCAR
    @0x49
    D=A
    @0
    SETCDR
    // ram[0] should have (x48 . x49 )
    @0
    D=M     // D = CAR of RAM[0]
    @0x6002
    M=D
    @0
    MCDR    // D = CDR of RAM[0]
    @0x6002
    M=D
(END)
    @END
    0;JMP
