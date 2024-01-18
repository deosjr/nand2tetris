// BUILTIN SYS FUNCTIONS
(SYSCALL)
    // push return-address
    @R15       // RET
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // push ENV    // save ENV of calling function
    @ENV
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // push ARG    // save ARG of calling function
    @ARG
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // ARG=SP-4  // reposition ARG
    @SP
    D=M
    @4
    D=D-A
    @ARG
    M=D
    // goto f      // transfer control
    @R13    // FUNC
    A=M
    0;JMP
(SYSRETURN)
    // FRAME = ARG
    @ARG
    D=M
    @R14    // FRAME
    // RET = *(FRAME+1)
    AM=D+1
    D=M
    @R15    // RET
    M=D
    // *ARG = pop()
    @SP
    AM=M-1
    D=M
    @ARG
    A=M
    M=D
    // SP = ARG+1 = FRAME
    @R14
    D=M
    @SP
    M=D
    // ENV = *(FRAME+2)
    @R14
    AM=M+1
    D=M
    @ENV
    M=D
    // ARG = *(FRAME+3)
    @R14
    AM=M+1
    D=M
    @ARG
    M=D
    // goto RET
    @R15
    A=M
    0;JMP
(SYSSTACKOVERFLOW)
    @0x0666
    D=A
    @0x6002
    M=D
    @SYSEND
    0;JMP
(SYSHEAPOVERFLOW)
    @0x0667
    D=A
    @0x6002
    M=D
    @SYSEND
    0;JMP
(SYSERRAPPLYNONPROC)
    @0x0668
    D=A
    @0x6002
    M=D
    @SYSEND
    0;JMP
(SYSERRSYMBOLNOTFOUND)
    @0x0669
    D=A
    @0x6002
    M=D
    @SYSEND
    0;JMP
(SYSERRUNKNOWNBUILTIN)
    @0x0670
    D=A
    @0x6002
    M=D
    @SYSEND
    0;JMP
