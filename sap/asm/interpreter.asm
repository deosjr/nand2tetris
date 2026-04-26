; SAP3 Forth interpreter — outer loop + WORD + FIND + NUMBER
;
; Memory layout:
;   0x000       : dictionary (grows up; oldest entry LINK = JMP OUTER trampoline)
;   0xD00       : data stack   (grows up; SP lives in X8)
;   0xE01       : CFAOUT       (FIND output: CFA of found word, or 0 on miss)
;   0xE02       : IMMOUT       (FIND output: non-zero if word is IMMEDIATE)
;   0xE10       : WORDBUF      (WORD output / FIND needle / NUMBER input)
;   0xE30       : NUMVAL       (NUMBER output: parsed value)
;   0xE31       : NUMOK        (NUMBER output: non-zero on success)
;   0xE50       : ERRFLAG      (interpreter output: 0 = ok, 1 = unknown token)
;
; Subroutine calling convention:
;   - Called with JMS, return with BRB.
;   - All subroutines MAY clobber A, B, X0..X7.
;   - X8 = data stack pointer; MUST be preserved by all subroutines.
;   - X9 reserved for future use (IP in DOCOL).
;
; Sysvars (passed as externals to assembleSAP3Labeled):
;   WORDBUF, CFAOUT, IMMOUT, NUMVAL, NUMOK, ERRFLAG
;
; ============================================================
; Dictionary, with the jump into outer interpreter as root link
; ============================================================
PLUS:
        JMP {OUTER}
        0x1
        0x2B00  ; +
        DEX 8   ; SP--
        LDN 8   ; A = top of stack
        DEX 8   ; SP--
        ADN 8   ; add
        STN 8   ; store
        INX 8   ; SP++
        BRB
DUP:
        0x{PLUS}
        0x3
        0x4455  ; DU
        0x5000  ; P
        DEX 8
        LDN 8
        INX 8
        STN 8
        INX 8
        BRB
DROP:
        0x{DUP}
        0x4
        0x4452  ; DR
        0x4F50  ; OP
        DEX 8
        BRB
SENTINEL:
        0x{DROP}
        0x0

; ============================================================
; Outer interpreter loop 
; NOTE: WB_INIT and EIGHT need to fit on same page!
; ============================================================
OUTER:
        LDX 8,{:I_SP_INIT}     ; X8 = stackBase (empty stack)

I_LOOP:
        JMS {WORD}              ; tokenise next word into WORDBUF
        LDA {WORDBUF}           ; load length
        JAZ {I_DONE}            ; length=0 → EOF, clean exit

        JMS {FIND}
        LDA {CFAOUT}
        JAZ {I_TRY_NUMBER}      ; miss → try interpreting as a number

        XCH 0                   ; X0 = CFA
        JSN 0                   ; indirect call through X0; word returns via BRB
        JMP {I_LOOP}

I_TRY_NUMBER:
        JMS {NUMBER}
        LDA {NUMOK}
        JAZ {I_ERROR}           ; neither word nor number
        LDA {NUMVAL}
        STN 8                   ; *SP = value
        INX 8                   ; SP++
        JMP {I_LOOP}

I_DONE:
        CLA
        STA {ERRFLAG}           ; ERRFLAG = 0 (success)
        HLT

I_ERROR:
        LDM
        0x1
        STA {ERRFLAG}           ; ERRFLAG = 1 (unknown token)
        HLT

I_SP_INIT:
        0xD00                   ; stackBase — loaded into X8 at startup

; ============================================================
; WORD — reads INP 1, writes WORDBUF
; Preserved: X8 (SP).  Clobbers: X0..X7, A.
; Output: WORDBUF[0] = byte-length (0 = EOF), WORDBUF[1..] = packed name.
; ============================================================

WORD:
        CLA
        XCH 0                   ; X0 = byte length counter
        CLA
        XCH 1                   ; X1 = high/low mode (0=high, -1=low)
        LDX 2,{:WB_INIT}        ; X2 = write cursor (name words)
        LDX 3,{:WB_INIT}        ; X3 = points at WORDBUF[0] (length slot)

W_LEAD:
        INP 1
        JAZ {W_END}             ; input 0 = EOF
        SBM
        0x21
        JAM {W_LEAD}            ; c < 0x21 → skip leading whitespace
        ADM
        0x21                    ; restore A

W_START:
        JAZ {W_END}             ; c == 0 (EOF after leading skip) → done
        SBM
        0x21
        JAM {W_END}             ; c < 0x21 → delimiter ends token (consumed)
        ADM
        0x21                    ; restore A
        INX 0                   ; length++
        JIZ 1,{:W_PREPHIGH}     ; X1==0 → this char goes in high byte

; W_PREPLOW — char goes in low byte; OR into W_TEMP cell
        ORM
W_TEMP:
        0x0                     ; inline immediate cell; STA {W_TEMP} writes high byte here
        DEX 1                   ; X1 = -1 (next char will be high)
        JMP {W_STORE}

W_PREPHIGH:
        LDX 4,{:EIGHT}          ; shift-counter = 8
W_SHIFT:
        SHL
        DSZ 4
        JMP {W_SHIFT}           ; A = char << 8
        STA {W_TEMP}            ; save high-byte-shifted value for OR below
        INX 2                   ; advance word cursor
        INX 1                   ; X1 = 1 (parity: next is low byte, but fall through flips)

W_STORE:
        STN 2                   ; *X2 = A (packed word)
        INP 1                   ; read next char
        JMP {W_START}

W_END:
        XCH 0                   ; A = length
        STN 3                   ; WORDBUF[0] = length
        BRB

; ============================================================
; FIND — looks up WORDBUF in the dictionary starting from HEAD
; Preserved: X8 (SP).  Clobbers: X4..X7, A.
; Output: CFAOUT = CFA address (0 on miss), IMMOUT = non-zero if IMMEDIATE.
; ============================================================

FIND:
        LDX 4,{:WB_INIT}        ; X4 = WORDBUF address (needle length at *X4)
        LDA {HEAD}              ; A = address of sentinel (newest entry)

F_LOOP:
        STM
F_SCRATCH:
        0x0                     ; scratch cell: current entry address
        LDX 5,{:F_SCRATCH}
        INX 5                   ; X5 → length/flags word

        LDN 5
        ANM
        0x7FFF                  ; mask off IMMEDIATE bit
        SBN 4                   ; compare entry length vs needle length
        JAZ {F_MATCH}

F_FOLLOW:
        LDX 5,{:F_SCRATCH}
        JIZ 5,{:F_FAIL}         ; entry address was 0 → end of chain, not found
        LDN 5                   ; follow LINK field
        JMP {F_LOOP}

F_MATCH:
        LDX 6,{:WB_INIT}        ; X6 = WORDBUF address (needle)
        LDN 6                   ; A = needle length
        ADM
        0x1
        SHR                     ; A = ceil(length/2) = number of name words to compare
        XCH 7                   ; X7 = comparison counter

F_INNER:
        INX 5                   ; X5 → next name word in dict entry
        INX 6                   ; X6 → next name word in WORDBUF
        LDN 5
        SBN 6
        JAZ {F_CONTINUE}
        JMP {F_FOLLOW}          ; mismatch → this entry fails, keep searching

F_CONTINUE:
        DSZ 7
        JMP {F_INNER}

; F_SUCCESS (fall-through from F_CONTINUE)
        INX 5                   ; X5 → CFA
        XCH 5
        STA {CFAOUT}            ; write CFA
        LDX 5,{:F_SCRATCH}
        INX 5                   ; X5 → length/flags word
        LDN 5
        ANM
        0x8000                  ; extract IMMEDIATE flag
        STA {IMMOUT}
        BRB

F_FAIL:
        CLA
        STA {CFAOUT}
        CLA
        STA {IMMOUT}
        BRB

; ============================================================
; NUMBER — parse WORDBUF as an unsigned decimal integer
; Preserved: X8 (SP).  Clobbers: X0..X4, A.
; Output: NUMVAL = parsed value (0 on failure), NUMOK = non-zero on success.
; ============================================================

NUMBER:
        LDX 0,{:WB_INIT}        ; X0 = WORDBUF pointer
        CLA
        XCH 1                   ; X1 = high/low read mode (0=high, -1=low)
        CLA
        XCH 2                   ; X2 = accumulator
        LDN 0
        XCH 3                   ; X3 = remaining byte count (= WORDBUF[0])

N_LOOP:
        JIZ 3,{:N_SUCCESS}      ; processed all bytes → done
        DEX 3
        JIZ 1,{:N_READHIGH}     ; X1==0 → read high byte of next word

; N_READLOW
        DEX 1                   ; X1 = -1 (next will be high)
        LDN 0
        ANM
        0xFF                    ; isolate low byte

N_LOW:
        SBM
        0x30                    ; subtract '0'
        JAM {N_FAIL}            ; A < 0 → not a digit
        SBM
        0xA
        JAM {N_DIGITFOUND}      ; A < 10 → valid digit
        JMP {N_FAIL}            ; A >= 10 → not a digit

N_DIGITFOUND:
        ADM
        0xA                     ; undo second subtraction: digit = A (0-9)
        XCH 2                   ; swap: new digit in X2, old accumulator in A
        STM
N_TEMP:
        0x0                     ; scratch: save old accumulator
        SHL
        SHL
        SHL                     ; A = old_acc << 3  (= old_acc * 8)
        ADD {N_TEMP}
        ADD {N_TEMP}            ; A = old_acc*8 + old_acc + old_acc = old_acc*10
        XCH 2                   ; A = new digit, X2 = old_acc*10
        STA {N_TEMP}            ; save digit
        XCH 2                   ; A = old_acc*10
        ADD {N_TEMP}            ; A = old_acc*10 + digit
        XCH 2                   ; X2 = new accumulator
        JMP {N_LOOP}

N_READHIGH:
        INX 1                   ; X1 = 1 (parity flip back, next read is low)
        LDX 4,{:EIGHT}
        INX 0                   ; advance word pointer
        LDN 0
N_SHIFT:
        SHR
        DSZ 4
        JMP {N_SHIFT}           ; A = *X0 >> 8 (high byte)
        JMP {N_LOW}

N_SUCCESS:
        XCH 2
        STA {NUMVAL}
        XCH 0
        STA {NUMOK}             ; NUMOK = X0 (non-zero if any bytes processed)
        BRB

N_FAIL:
        CLA
        STA {NUMVAL}
        STA {NUMOK}
        BRB

; ============================================================
; Shared read-only constants (used by WORD, FIND, and NUMBER)
; ============================================================

WB_INIT:
        0xE10                   ; WORDBUF base address

EIGHT:
        0x8                     ; shift count (used in WORD and NUMBER)

HEAD:
        0x{SENTINEL}