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
    // a live pointer is a pair that's not NIL or a userdefined func
    // but even easier, it is a pointer into the heap (or usrdef)
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
    ISUSRDEF
    @GCPOINTERFOUND
    !D;JEQ
    @R6
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
(GCPOINTERFOUND)
    @R6
    A=M-1
    D=M
    @R8
    M=D         // copy over because duplicate check is also jumped to from (1c)
// (1b) only copy over if we haven't seen pointer yet
(GCDUPLICATE)
    @R8
    D=M
    @0x7fff
    D=D&A       // set first bit to 0, making pointer out of usrdef
    @R8
    M=D
    @GCDUPRET
    D=A
    @R14
    M=D
    @GCFINDINHEAP
    0;JMP
(GCDUPRET)
    @R10
    D=M+1
    @GCSTACKWALK
    D;JNE // jump if not R10=-1
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
    @R8         // put car on stack if it is a live pointer
    A=M
    ISUSRDEF
    @GCCARFOUND
    !D;JEQ
    @R8
    A=M
    MCAR
    @0x0800
    D=D-A
    @GCCDR
    D;JLT
    @0x3800
    D=D-A
    @GCCDR
    D;JGE
(GCCARFOUND)
    @R8
    A=M
    MCAR
    @SP
    M=M+1
    A=M-1
    M=D
(GCCDR)
    @R8         // goto GCSTACKWALK if cdr is not a live pointer, otherwise recurse
    A=M
    USRDEFCDR
    @GCCDRFOUND
    !D;JEQ
    @R8
    A=M
    MCDR
    @0x0800
    D=D-A
    @GCSTACKWALK
    D;JLT
    @0x3800
    D=D-A
    @GCSTACKWALK
    D;JGE
(GCCDRFOUND)
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
// (2)  quicksort the secondary heap in-place
    @0x2010     // bottom of secondary heap
    D=A
    @SP
    M=M+1
    A=M-1
    M=D
    @R7
    D=M-1       // top of secondary heap
    @R12        // R12 stores top of secondary heap
    M=D
    @SP
    M=M+1
    A=M-1
    M=D
(GCQUICKSORT)
    // if @SP == R5, we are done and goto next step (3)
    // assumes SP never goes below R5, only grows!
    @R5
    D=M
    @SP
    D=D-M
    @GCCOPY
    D;JEQ
    // takes two arguments from heap: lo and hi
    // assumes each element is unique, which is guaranteed by duplicate check
    // reuses R6 (low), R7 (high) and R10 (pivot)
    // R8 is left, R9 is right
    @SP
    AM=M-1
    D=M
    @R7
    M=D
    @R9
    M=D-1
    @SP
    AM=M-1
    D=M
    @R6
    M=D
    @R8
    M=D
    @R7
    D=M-D
    @GCQUICKSORT
    D;JLE       // if low >= high, list is small enough to be sorted automatically
    @R7
    A=M
    D=M
    @R10
    M=D         // pivot
(GCSORTLOOP)
    @R9
    D=M
    @R8
    D=D-M
    @GCENDLOOP  // while l <= r do, i.e.
    D;JLT       // if r < l then jump to end
(GCLEFTLOOP)
    @R8
    D=M
    @R9
    D=M-D
    @GCRIGHTLOOP // while l <= r AND *l <= pivot
    D;JLT        // if l > r then jump
    @R8
    A=M
    D=M
    @R10
    D=D-M
    @GCRIGHTLOOP
    D;JGT       // if *l > pivot then jump
    @R8
    M=M+1       // left++
    @GCLEFTLOOP
    0;JMP
(GCRIGHTLOOP)
    @R9
    D=M
    @R8
    D=D-M
    @GCWHILESWAP    // while r >= l AND *r >= pivot
    D;JLT           // if r < l then jump
    @R9
    A=M
    D=M
    @R10
    D=M-D
    @GCWHILESWAP
    D;JGT       // if *r < pivot then jump
    @R9
    M=M-1       // right--
    @GCRIGHTLOOP
    0;JMP
(GCWHILESWAP)
    // if l < r then swap
    @R8
    D=M
    @R9
    D=D-M
    @GCSORTLOOP     // if l < r then swap
    D;JGE           // if l - r >= 0 then jump
    @R8
    A=M
    D=M
    @R11
    M=D             // R11 is temp
    @R9
    A=M
    D=M
    @R8
    A=M
    M=D
    @R11
    D=M
    @R9
    A=M
    M=D             // swap *R8 and *R9
    @GCSORTLOOP
    0;JMP
(GCENDLOOP)
    @R8
    A=M
    D=M
    @R7
    A=M
    M=D
    @R10
    D=M
    @R8
    A=M
    M=D     // swap *l and pivot
    @R6
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    @R8
    D=M-1
    @SP
    M=M+1
    A=M-1
    M=D     // quicksort(low, l-1)
    @R8
    D=M+1
    @SP
    M=M+1
    A=M-1
    M=D
    @R7
    D=M
    @SP
    M=M+1
    A=M-1
    M=D     // quicksort(l+1, high)
    @GCQUICKSORT
    0;JMP
(GCCOPY)
// (3)  copy each live cons cell onto bottom of the heap
// since we sorted first, we will only copy over dead cells or cells that have already been moved
    @0x200F     // bottom of secondary heap - 1
    D=A
    @R6
    M=D
(GCCOPYLOOP)
    @R6
    D=M
    @R12        // stores top of secondary heap
    D=M-D
    @GCUPDATE
    D;JEQ       // if R6 == R12 we have copied the entire secondary heap, goto (4)
    @R6
    DM=M+1
    @0x1810     // 0x2010 - 0x800, mapping secondary onto primary stack by index
    D=D-A
    @R7
    M=D
    @R6
    A=M
    A=M
    MCDR
    @R7
    A=M
    SETCDR
    @R6
    A=M
    A=M
    MCAR
    @R7
    A=M
    SETCAR
    @GCCOPYLOOP
    0;JMP
(GCUPDATE)
// (4)  update all pointers in stack and primary heap, both car and cdr
    @0x0010
    D=A
    @R6
    M=D
    @R12
    D=M
    @R7
    M=D+1
(GCUPDATESTACKLOOP)
    @R6
    D=M
    @R5
    D=D-M
    @GCUPDATEHEAP
    D;JGE       // if R6 >= R5 we are done
    @R6
    M=M+1
    A=M-1
    ISUSRDEF
    @R13
    M=D         // R13 stores whether this was a userdefined func
    @GCSTACKPOINTERFOUND
    !D;JEQ
    @R6
    A=M-1
    D=M
    @0x0800
    D=D-A
    @GCUPDATESTACKLOOP
    D;JLT
    @0x3800     // 0x4000 - 0x0800 that has already been subtracted
    D=D-A
    @GCUPDATESTACKLOOP
    D;JGE
(GCSTACKPOINTERFOUND)
    @R6
    A=M-1
    D=M
    @0x7fff
    D=D&A
    @R8
    M=D
    @GCSTACKRET
    D=A
    @R14
    M=D
    @GCFINDINHEAP
    0;JMP
(GCSTACKRET)
    @0x7fff     // if @R13=0xffff, this was a userdefined func, therefore add high bit back
    D=A+1       // 0x8000, i.e. highest bit set
    @R13
    D=D&M       // 0x8000 if @R6 pointed at usrdef, 0x0 otherwise
    @R10
    D=D|M       // *R10, + high bit set if this was a usrdef
    @0x1810     // 0x2010 - 0x800, mapping secondary onto primary stack by index
    D=D-A       // D is new pointer value!
    @R6
    A=M-1
    M=D         // overwrite stack pointer value
    @GCUPDATESTACKLOOP
    0;JMP
(GCUPDATEHEAP)
    @R7
    D=M
    @0x1810
    D=D-A
    @R5
    M=D
    @0x0800
    D=A
    @R6
    M=D
(GCUPDATEHEAPLOOP)
    @R6
    D=M
    @R5
    D=D-M
    @GCRETURN
    D;JGE       // if R6 >= R5 we are done
    @R6
    M=M+1
    A=M-1
    USRDEFCDR
    @R13
    M=D
    @GCCDRPOINTERFOUND
    !D;JEQ
    @R6
    A=M-1
    MCDR
    @0x0800
    D=D-A
    @GCUPDATECAR
    D;JLT
    @0x3800     // 0x4000 - 0x0800 that has already been subtracted
    D=D-A
    @GCUPDATECAR
    D;JGE
(GCCDRPOINTERFOUND)
    @R6
    A=M-1
    MCDR
    @0x7fff
    D=D&A
    @R8
    M=D
    @GCCDRRET
    D=A
    @R14
    M=D
    @GCFINDINHEAP
    0;JMP
(GCCDRRET)
    @0x7fff
    D=A+1
    @R13
    D=D&M
    @R10
    D=D|M
    @0x1810     // 0x2010 - 0x800, mapping secondary onto primary stack by index
    D=D-A       // D is new pointer value!
    @R6
    A=M-1
    SETCDR      // overwrite cdr value in primary heap
(GCUPDATECAR)
    @R6
    A=M-1
    ISUSRDEF
    @R13
    M=D
    @GCCARPOINTERFOUND
    !D;JEQ
    @R6
    A=M-1
    MCAR
    @0x0800
    D=D-A
    @GCUPDATEHEAPLOOP
    D;JLT
    @0x3800     // 0x4000 - 0x0800 that has already been subtracted
    D=D-A
    @GCUPDATEHEAPLOOP
    D;JGE
(GCCARPOINTERFOUND)
    @R6
    A=M-1
    MCAR
    @0x7fff
    D=D&A
    @R8
    M=D
    @GCCARRET
    D=A
    @R14
    M=D
    @GCFINDINHEAP
    0;JMP
(GCCARRET)
    @0x7fff
    D=A+1
    @R13
    D=D&M
    @R10
    D=D|M
    @0x1810     // 0x2010 - 0x800, mapping secondary onto primary stack by index
    D=D-A       // D is new pointer value!
    @R6
    A=M-1
    SETCAR      // overwrite car value in primary heap
    @GCUPDATEHEAPLOOP
    0;JMP
(GCRETURN)
// (5) return
// TODO: clearing secondary heap shouldnt be necessary
// there's a bug somewhere that causes us to read from previous GC run!
    @0x2010
    D=A
    @R6
    M=D
(GCCLEARSECONDARY)
    @R7
    D=M
    @R6
    D=D-M
    @GCCLEARED
    D;JEQ
    @R6
    M=M+1
    A=M-1
    M=0
    @GCCLEARSECONDARY
    0;JMP
(GCCLEARED)
    @R5
    D=M
    @FREE
    M=D
    @R15
    A=M
    0;JMP
(GCFINDINHEAP)
    // R7 = top of secondary heap + 1
    // R8 = value to find
    // R9 = temp pointer
    // R10 = return value
    // R14 = return address
    // returns address (in secondary heap) if found or -1 if not found
    @0x200F     // start of secondary heap - 1
    D=A
    @R9
    M=D
(GCFINDLOOP)
    @R9
    AM=M+1
    D=M         // next value in secondary heap
    @R8
    D=D-M
    @GCFOUND
    D;JEQ       // if *R9 = R8, we have found a match!
    @R7         // end of used secondary heap
    D=M+1
    @R9
    D=D-M
    @GCFINDLOOP
    D;JGT       // if R7+1 - R9 > 0, we still have secondary heap to inspect
(GCNOTFOUND)
    @R10
    M=-1
    @R14
    A=M
    0;JMP
(GCFOUND)
    @R9
    D=M
    @R10
    M=D
    @R14
    A=M
    0;JMP
