// COMPILED FUNCTIONS
(COMPILEDADD)
    // (define add (lambda (x y) (+ x y)))
    // ARG 0 = x
    // ARG 1 = y
    // create (x y), then call builtin +
    D=0
    @FREE
    A=M
    SETCDR
    @ARG
    A=M
    A=M
    ACDR
    MCAR
    @FREE
    A=M
    SETCAR
    @FREE
    D=M
    M=D+1
    @FREE
    A=M
    SETCDR
    @ARG
    A=M
    A=M
    MCAR
    @FREE
    A=M
    SETCAR
    @FREE
    D=M
    M=D+1
	@SP
	M=M+1
	A=M-1
	M=D
	@SYSRETURN
	D=A
	@R15
	M=D
    @BUILTINADD
    0;JMP
(MAINMAIN)
// add compiled funcs to env
    @0x6042     // random chosen symbol for 'add'
    D=A
    @FREE
    A=M
    SETCAR
    @COMPILEDADD
    D=A
// COMPILED prefix is 110
// assumes label is smaller than 0x2000 !
    @0x7fff
    D=D+A
    @0x4001
    D=D+A
    @FREE
    A=M
    SETCDR
    @FREE
    D=M
    AM=D+1
    SETCAR
    @ENV
    A=M
    D=M
    @FREE
    A=M
    SETCDR
    @FREE
    D=M
    M=D+1
    @ENV
    A=M
    M=D
// set up and eval interpreted code using compiled funcs
// we will interpret "(add 5 3)" using compiled add func
	@ENV
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
    D=0
    @FREE
    A=M
    SETCDR
    @0x4003
    D=A
    @FREE
    A=M
    SETCAR
    @FREE
    D=M
    M=D+1
    @FREE
    A=M
    SETCDR
    @0x4004
    D=A
    @FREE
    A=M
    SETCAR
    @FREE
    D=M
    M=D+1
    @FREE
    A=M
    SETCDR
    @0x6042     // add
    //@0x6007   // +
    D=A
    @FREE
    A=M
    SETCAR
    @FREE
    D=M
    M=D+1
    @SP
    M=M+1
    A=M-1
    M=D
// call eval
    @EVALEVAL
    D=A
    @R13
    M=D
    @MAINRET
    D=A
    @R15
    M=D
    @SYSCALL
    0;JMP
(MAINRET)
    @SP
    A=M-1
    D=M
    @0x1fff
    D=D&A
    @48
    D=D+A
    @0x6002
    M=D
    @SYSEND
    0;JMP
