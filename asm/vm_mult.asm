// handmade translation of Figure 7.9 VM code into Hack assembly
// always call Sys.init at the start (which lives in Sys.vm)
// annotations mostly from Figs 7.9 and 8.5
// R13-R15 can be used by the VM implementation as general-purpose registers
// R13: FUNC, R14: FRAME/NUMARGS/NUMLCL, R15: RET
// NOTE true = -1 or 0xFFFF and false = 0 or 0x0000
    @256
    D=A
    @SP // SP always points to first available (empty) place on top of stack
    M=D
    @SYSINIT
    0;JMP
(SYSINIT)
    // call main 0
    @MAIN
    D=A
    @R13    // FUNC
    M=D
    @0x0000 // since main has 0 args
    D=A
    @R14    // NUMARGS
    M=D
    @SYSEND
    D=A
    @R15    // RET
    M=D
    @SYSCALL
    0;JMP
// (return address)
(SYSEND)
    @SYSEND
    0;JMP
(MAIN)
    // NOTE numlcl is not always numargs!
    // function declaration: push numlcl times 0 (so none here)
    // push args to mult: x=2 and y=3
    @2
    D=A
    @SP
    M=M+1
    A=M-1
    M=D
    @3
    D=A
    @SP
    M=M+1
    A=M-1
    M=D
    // call mult 2
    @MULT
    D=A
    @R13    // FUNC
    M=D
    @2      // since mult has 2 args
    D=A
    @R14    // NUMARGS
    M=D
    @RETFROMMULT
    D=A
    @R15    // RET
    M=D
    @SYSCALL
    0;JMP
// (return address)
(RETFROMMULT)
    // pop and print answer (expect 6)
    @SP
    AM=M-1
    D=M
    @0x6002
    M=D
    // before returning, the called function must push a value onto the stack (0 for void?)
    @SP
    M=M+1
    A=M-1
    M=0
    // return to caller
    @SYSRETURN
    0;JMP
//function mult 2 // 2 local variables
(MULT)
    // function declaration: push numlcl times 0
    @2
    D=A
    @R14    // NUMLCL
    M=D
    @MULTLCL
    D=A
    @R15    // RET
    M=D
    @SYSPUSHLCL
    0;JMP
(MULTLCL)
    //push constant 0
    @SP
    M=M+1
    A=M-1
    M=0
    //pop local 0 // sum=0
    @SP
    AM=M-1
    D=M
    @LCL
    A=M
    M=D
    //push argument 1
    @ARG
    A=M+1
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    //pop local 1 // j=y
    @SP
    AM=M-1
    D=M
    @LCL
    A=M+1
    M=D
//label loop
(LOOP)
    //push constant 0
    @SP
    M=M+1
    A=M-1
    M=0
    //push local 1
    @LCL
    A=M+1
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    //Eq
    @SP
    AM=M-1
    D=M
    A=A-1
    D=D-M
    // if D=0 set M=0xffff otherwise set M=0x0000
    M=0
    @FALSE // TODO in vm->asm, use hardcoded line here instead of yet more labels
    D;JNE
//TRUE
    @SP
    A=M-1
    M=!M
(FALSE)
    //if-goto end // if j=0 goto end
    @SP
    AM=M-1
    D=M
    @END
    !D;JEQ
    //push local 0
    @LCL
    A=M
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    //push argument 0
    @ARG
    A=M
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    //Add
    @SP
    AM=M-1
    D=M
    A=A-1
    M=D+M
    //pop local 0 // sum=sum+x
    @SP
    AM=M-1
    D=M
    @LCL
    A=M
    M=D
    //push local 1
    @LCL
    A=M+1
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    //push constant 1
    @SP
    M=M+1
    A=M-1
    M=1
    //Sub
    @SP
    AM=M-1
    D=M
    A=A-1
    M=M-D
    //pop local 1 // j=j-1
    @SP
    AM=M-1
    D=M
    @LCL
    A=M+1
    M=D
    //goto loop
    @LOOP
    0;JMP
//label end
(END)
    //push local 0
    @LCL
    A=M
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    //return // return sum
    @SYSRETURN
    0;JMP
(SYSCALL)
    // push return-address
    @R15       // RET
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // push LCL    // save LCL of calling function
    @LCL
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
    // push THIS   // save THIS of calling function
    @THIS
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // push THAT   // save THAT of calling function
    @THAT
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // ARG=SP-n-5  // reposition ARG (n=number of args)
    @SP
    D=M
    @R14    // NUMARGS
    D=D-M
    @5
    D=D-A
    @ARG
    M=D
    // LCL=SP      // reposition LCL
    @SP
    D=M
    @LCL
    M=D
    // goto f      // transfer control
    @R13    // FUNC
    A=M
    0;JMP
(SYSPUSHLCL)
    // func def here i.e. push 0 numlcl times
    @R14    // NUMLCL
    D=M
    @R15    // RET
    A=M
    D;JEQ
    @SP
    M=M+1
    A=M-1
    M=0
    @R14    // NUMLCL
    M=M-1
    @SYSPUSHLCL
    0;JMP
(SYSRETURN)
    // FRAME = LCL
    @LCL
    D=M
    @R14    // FRAME
    DM=D
    // RET = *(FRAME-5)
    @5
    A=D-A
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
    // SP = ARG+1
    @ARG
    D=M+1
    @SP
    M=D
    // THAT = *(FRAME-1)
    @R14
    AM=M-1
    D=M
    @THAT
    M=D
    // THIS = *(FRAME-2)
    @R14
    AM=M-1
    D=M
    @THIS
    M=D
    // ARG = *(FRAME-3)
    @R14
    AM=M-1
    D=M
    @ARG
    M=D
    // LCL = *(FRAME-4)
    @R14
    AM=M-1
    D=M
    @LCL
    M=D
    // goto RET
    @R15
    A=M
    0;JMP
