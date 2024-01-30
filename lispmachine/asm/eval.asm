// (eval env e) -> evaluation of e in env, or NIL if error
// only user of the stack vm abstraction!
(EVALEVAL)
    // prepare 1 local variable, initialised to 0
	@SP
	M=M+1
	A=M-1
	M=0
(EVALSTART)
    // TODO: unsafe version that doesnt check overflow? saves instructions!
    @SP
    D=M
    @0x07ff         // end of stack
    D=D-A
    @SYSSTACKOVERFLOW
    D;JGE           // if @SP - 0x07ff >= 0 -> stack overflow
    @FREE
    D=M
    @0x3fff         // end of heap
    D=D-A
    @SYSHEAPOVERFLOW
    D;JGE           // if @FREE - 0x3fff >= 0 -> heap overflow
    @ARG
    A=M
    ISPROC
    @EVALSELF
    !D;JEQ          // if (ISPROC e) then goto evalself
    @ARG
    A=M
    ISSYMB
    @EVALSYMBOL
    !D;JEQ          // if (ISSYMB e) then goto evalsymbol
    @ARG
    A=M
    ISPRIM
    @EVALSELF
    !D;JEQ          // if (ISPRIM e) then goto evalself
(EVALPAIR)          // fallthrough
    @ENV
    A=M
    D=M
	@SP
	M=M+1
	A=M-1
	M=D
    @ARG
    A=M
    A=M
    MCAR
	@SP
	M=M+1
	A=M-1
	M=D
    @EVALEVAL
    D=A
    @R13
    M=D
    @EVALFIRSTRET
    D=A
    @R15
    M=D
    @SYSCALL
    0;JMP
(EVALFIRSTRET)
    // if return value is not proc
    // -> goto sys.errapplynonproc
    @SP
    A=M-1
    ISPROC
    @SYSERRAPPLYNONPROC
    !D;JNE
    // else, local 0 = evalled proc
    // and goto evalprocedure
    @SP
    A=M-1
    ISSPECIAL
    @EVALSPECIAL
    !D;JEQ
    @ARG
    D=M
    @4
    D=D+A
    @R14
    M=D
    @SP
    A=M-1
    D=M
    @R14
    A=M
    M=D             // local 0 = evalled proc
    @EVALPROCEDURE
    0;JMP
(EVALSELF)
	@ARG
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SYSRETURN
	0;JMP
(EVALSYMBOL)
    // construct ( env . ( arg . () ))
    // because we are calling (assq assoclist key)
    // FREE = ( arg . () )
	D=0
	@FREE
	A=M
	SETCDR
	@ARG
    A=M
	D=M
	@FREE
	A=M
	SETCAR
    // FREE+1 = ( env . FREE )
	@FREE
	D=M
	AM=D+1
	SETCDR
    @ENV
    A=M
    D=M
	@FREE
	A=M
	SETCAR
	@FREE
	D=M
	M=D+1
	@SP
    M=M+1
	A=M-1
	M=D         // put on stack as arg
    @EVALASSQRET
    D=A
    @R15
    M=D
    @BUILTINASSQ
    0;JMP
(EVALASSQRET)
    @SP
    A=M-1
    D=M
    @SYSERRSYMBOLNOTFOUND
    D;JEQ           // if assq returns NIL, that means not found
	@SYSRETURN
	0;JMP
(EVALPROCEDURE)
	@ARG
	A=M
    ACAR
    EMPTYCDR
	@EVALZEROARGFUNC
	!D;JEQ
(EVALARGS)
    @FREE
    D=M
    @SP
    M=M+1
    A=M-1
    M=D             // push FREE: head of list of evalled args
    @SP
    M=M+1
    A=M-1
    M=D
    @FREE
    M=M+1           // push FREE again then incr
(EVALARGLOOP)
    @ARG
    A=M
    A=M
    MCDR
    @ARG
    A=M
    M=D             // @ARG = (CDR ARG)
    @ENV
    A=M
    D=M
	@SP
	M=M+1
	A=M-1
	M=D
    @ARG
    A=M
    A=M
    MCAR            // call eval of next arg
	@SP
	M=M+1
	A=M-1
	M=D
    @EVALEVAL
    D=A
    @R13
    M=D
    @EVALARGLOOPRET
    D=A
    @R15
    M=D
    @SYSCALL
    0;JMP
(EVALARGLOOPRET)
	@SP
	AM=M-1
	D=M             // result of eval
	@SP
	AM=M-1
	A=M             // allocated cons cell
    SETCAR
    D=A
    @R14
    M=D             // store address in R14
	@ARG
	A=M
	A=M
    EMPTYCDR
	@EVALENDARGLOOP
	!D;JEQ
    @FREE
    D=M
    @R14
    A=M
    SETCDR          // setcdr to next available cell
	@SP
	M=M+1
	A=M-1
	M=D
    @FREE
    M=M+1           // preallocate cons cell for next loop iteration
    @EVALARGLOOP
    0;JMP
(EVALZEROARGFUNC)
	@SP
	M=M+1
	A=M-1
	M=0
    @EVALAPPLY
    0;JMP
(EVALENDARGLOOP)
    @R14
    A=M
    D=0
    SETCDR          // fallthrough to apply
(EVALAPPLY)
	@SP
	AM=M-1
	D=M
    @R6
    M=D             // R6 holds evalled args
	@4
	D=A
	@ARG
	A=D+M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	A=M-1
	ISBUILTIN
	M=D
	@SP
	AM=M-1
	D=M
	@EVALCALLBUILTIN
	!D;JEQ
//(EVAL USER DEFINED FUNC)
	@4
	D=A
	@ARG
	A=D+M
	D=M
	@0x1fff
	D=D&A
    @R5
	AM=D            // R5 = userdefined unmasked
    ACDR
    MCAR
    @R7
    M=D             // R7 = f.params
    @R5
    A=M
    MCAR
    @SP
    M=M+1
    A=M-1
    M=D             // push f.env = x
(EVALAPPLYREC)
    // TODO: check if len(params) == len(args)
    // NOTE: order actually doesn't matter here!
    // so we cons in reverse on top of f.env
    @R7
    ISEMPTY
    @EVALAPPLYEND
    !D;JEQ
    @R6
    A=M
    MCAR
    @FREE
    A=M
    SETCDR
    @R7
    A=M
    MCAR
    @FREE
    A=M
    SETCAR          // (cons (car R7) (car R6)) = y
    @FREE
    M=M+1
    @SP
    AM=M-1
    D=M
    @FREE
    A=M
    SETCDR
    @FREE
    D=M-1
    A=M
    SETCAR          // (cons y x) = z
    @FREE
    D=M
    M=M+1
    @SP
    M=M+1
    A=M-1
    M=D             // put z on stack as new env
    @R6
    A=M
    MCDR
    @R6
    M=D             // R6 = (cdr R6)
    @R7
    A=M
    MCDR        
    @R7
    M=D             // R7 = (cdr R7)
    @EVALAPPLYREC
    0;JMP
(EVALAPPLYEND)
	@SP
	AM=M-1
	D=M
	@ENV
	A=M
	M=D             // ENV = oldenv + arg bindings
    @R5
    A=M
    ACDR
    ACDR
    MCAR            // f.body
	@ARG
	A=M
	M=D             // ARG = f.body
	@EVALSTART
	0;JMP
(EVALCALLBUILTIN)
    @R6
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SYSRETURN
	D=A
	@R15
	M=D
	@4
	D=A
	@ARG
	A=D+M
	D=M
	@0x1fff
	A=D&A
	0;JMP
(EVALSPECIAL)
	@ARG
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	A=M-1
	A=M
	MCDR
	@SP
	A=M-1
	M=D
	@SP
	AM=M-1
	D=M
	@ARG
	A=M
	M=D             // @ARG = (cdr ARG)
    @SP
    A=M-1
	D=M             // evalled proc is still on the stack
    @0x1fff
    D=D&A
    @EVALIF
    D;JEQ
    D=D-1
    @EVALDEFINE
    D;JEQ
    D=D-1
    @EVALQUOTE
    D;JEQ
    D=D-1
    @EVALSET
    D;JEQ
    D=D-1
    @EVALLAMBDA
    D;JEQ
    D=D-1
    @EVALBEGIN
    D;JEQ
	@SYSERRUNKNOWNBUILTIN
	0;JMP
(EVALIF)
	@ENV
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@ARG
	A=M
    A=M
    MCAR
	@SP
	M=M+1
	A=M-1
	M=D
	@EVALEVAL
	D=A
	@R13
	M=D
	@EVALIFRET
	D=A
	@R15
	M=D
	@SYSCALL
	0;JMP
(EVALIFRET)
	@SP
	AM=M-1
	D=M
	@EVALALT
	D;JEQ
	@ARG
	A=M
    A=M
    ACDR
    MCAR
	@ARG
	A=M
	M=D
	@EVALSTART
	0;JMP
(EVALALT)
    // TODO: if no alt, return 'false'
	@ARG
	A=M
    A=M
    ACDR
    ACDR
    MCAR
	@ARG
	A=M
	M=D
	@EVALSTART
	0;JMP
(EVALDEFINE)
    // TODO: this just adds, doesnt check if already exists in env
    // meaning currently the assoc list could have duplicate keys
	@ARG
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	A=M-1
	A=M
	MCAR
	@SP
	A=M-1
	M=D
	@ENV
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@ARG
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	A=M-1
	A=M
	ACDR
	MCAR
	@SP
	A=M-1
	M=D
	@EVALEVAL
	D=A
	@R13
	M=D
	@EVALDEFINERET
	D=A
	@R15
	M=D
	@SYSCALL
	0;JMP
(EVALDEFINERET)
	@SP
	AM=M-1
	D=M
	@FREE
	A=M
	SETCDR
	@SP
	A=M-1
	D=M
	@FREE
	A=M
	SETCAR
	@FREE
	D=M
	M=D+1
	@SP
	A=M-1
	M=D
	@ENV
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	A=M-1
	A=M
	MCAR
	@SP
	A=M-1
	M=D
	@ENV
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	A=M-1
	A=M
	MCDR
	@SP
	A=M-1
	M=D
	@SP
	AM=M-1
	D=M
	@FREE
	A=M
	SETCDR
	@SP
	A=M-1
	D=M
	@FREE
	A=M
	SETCAR
	@FREE
	D=M
	M=D+1
	@SP
	A=M-1
	M=D
	@SP
	AM=M-1
	D=M
	@FREE
	A=M
	SETCDR
	@SP
	A=M-1
	D=M
	@FREE
	A=M
	SETCAR
	@FREE
	D=M
	M=D+1
	@SP
	A=M-1
	M=D
	@ENV
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	AM=M-1
	D=M
	@R7
	M=D
	@SP
	AM=M-1
	D=M
	@R8
	AM=D
	MCDR
	@R7
	A=M
	SETCDR
	@R8
	A=M
	MCAR
	@R7
	A=M
	SETCAR
	@SP
	M=M+1
	A=M-1
	M=0
	@SYSRETURN
	0;JMP
(EVALQUOTE)
	@ARG
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	A=M-1
	A=M
	MCAR
	@SP
	A=M-1
	M=D
	@SYSRETURN
	0;JMP
(EVALSET)
    // TODO
(EVALLAMBDA)
    // (lambda (params ...) body)
    // assumption: params is a list of symbols
    // TODO: typecheck param arg and check length of args=2
    // TODO: invalid parameter list error if duplicate symbols
	@ENV
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@ARG
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@SP
	AM=M-1
	D=M
	@FREE
	A=M
	SETCDR
	@SP
	A=M-1
	D=M
	@FREE
	A=M
	SETCAR
	@FREE
	D=M
	M=D+1
	@SP
	A=M-1
	M=D
	@0x7fff
	D=A
	@0x0001
	D=D+A
	@SP
	A=M-1
	M=D|M
	@SYSRETURN
	0;JMP
(EVALBEGIN)
	@ARG
	A=M
    A=M
    EMPTYCDR
    @EVALBEGINEND
    !D;JEQ
	@ENV
	A=M
	D=M
	@SP
	M=M+1
	A=M-1
	M=D
	@ARG
	A=M
    A=M
    MCAR
	@SP
	M=M+1
	A=M-1
	M=D
	@EVALEVAL
	D=A
	@R13
	M=D
	@EVALBEGINRET
	D=A
	@R15
	M=D
	@SYSCALL
	0;JMP
(EVALBEGINRET)
    @SP
    M=M-1           // ignore output
	@ARG
	A=M
	A=M
	MCDR
	@ARG
	A=M
	M=D             // @ARG = (cdr ARG)
    @EVALBEGIN
    0;JMP
(EVALBEGINEND)
    @ARG
    A=M
    A=M
    MCAR
    @ARG
    A=M
    M=D
    @EVALSTART
    0;JMP
