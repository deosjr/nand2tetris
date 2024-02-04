// Assumptions: stack runs until 0x07ff
// full heap starts at 0x0800 and ends at 0x3fff
// practically we will use only 0x0800 until 0x1fff, first half or primary heap
// then a buffer for overrun of size 16 (? would be nice to have a guarantee)
// so second half of heap is 0x2010 until 0x3fff, or secondary heap
// This GC implementation is a modified Cheney's algorithm
(GCSTART)
    @SP
    D=M
    @R5
    M=D         // R5 = end of stack at start of GC
    @0x0010     // we start traversing at bottom of stack
    D=A
    @R6
    M=D         // R6 = pointer in gray set
    @0x2010
    D=A
    @R7
    M=D         // R7 = pointer in secondary heap
// (1a) traverse stack, copy live pointers to secondary heap
(GCTRAVERSE)
    // a live pointer is a pair that's not NIL
    // but even easier, it is a pointer into the heap
    // so any value between 0x0800 and 0x3fff (arguably only in primary heap)
    @R6
    D=M
    @R5
    D=D-M
    @GCSORTHEAP
    D;JGE       // if R6 >= R5, end traversal
    @R6
    M=M+1
    A=M-1
    D=M
    @0x0800
    D=D-A
    @GCTRAVERSE
    D;JLT
    @0x3800     // 0x4000 - 0x0800 that has already been subtracted
    D=D-A
    @GCTRAVERSE
    D;JGE
    @R6
    A=M-1
    D=M
    @R8
    M=D         // copy over because duplicate check is also jumped to from (1c)
// (1b) only copy over if we haven't seen pointer yet
(GCDUPLICATE)
    @0x200F     // start of secondary heap - 1
    D=A
    @R9
    M=D
(GCDUPLICATELOOP)
    @R9
    AM=M+1
    D=M         // next value in secondary heap
    @R8
    D=D-M
    @GCSTACKWALK
    D;JEQ       // if *R9 = R8, we have found a duplicate and continue walking
    @R7         // end of used secondary heap
    D=M
    @R9
    D=D-M
    @GCDUPLICATELOOP
    D;JGT       // if R7 - R9 > 0, we still have secondary heap to inspect
    @R8
    D=M
    @R7
    M=M+1
    A=M-1
    M=D         // actually write value to secondary heap
// (1c) each new live pointer can lead to more, so traverse each cons
// we can traverse cdrs and park pointers on the stack if car also contains live pointer
// storing part on working stack needs O(log n) space, same as quicksort does
// we're done once we have consumed the stack back to R5
(GCCDRWALK)
    @R8
    A=M
    MCAR        // put car on stack if is a live pointer
    @0x0800
    D=D-A
    @GCCDR
    D;JLT
    @0x3800
    D=D-A
    @GCCDR
    D;JGE
    @R8
    A=M
    MCAR
    @SP
    M=M+1
    A=M-1
    M=D
(GCCDR)
    @R8
    A=M
    MCDR        // goto GCSTACKWALK if cdr is not a live pointer, otherwise recurse
    @0x0800
    D=D-A
    @GCSTACKWALK
    D;JLT
    @0x3800
    D=D-A
    @GCSTACKWALK
    D;JGE
    @R8
    A=M
    MCDR
    @R8
    M=D
    @GCDUPLICATE
    0;JMP
(GCSTACKWALK)
    // if @SP == R5, we are done and goto GCTRAVERSE
    // assumes SP never goes below R5, only grows!
    @R5
    D=M
    @SP
    D=D-M
    @GCTRAVERSE
    D;JEQ
    // otherwise, we consume a pointer from the stack, store it in R8
    // then goto GCDUPLICATE
    @SP
    AM=M-1
    D=M
    @R8
    M=D
    @GCDUPLICATE
    0;JMP
(GCSORTHEAP)
// (2)  (bubble/quick)sort the secondary heap in-place
    // TODO
// (3)  copy each live cons cell onto bottom of the heap
// since we sorted first, we will only copy over dead cells or cells that have already been moved
    // TODO
// (4)  update all pointers in stack and primary heap, both car and cdr
    // TODO
// (5) return
    // TODO TEST: for now terminate program after GC run
    //@R15
    //A=M
    @SYSEND
    0;JMP
