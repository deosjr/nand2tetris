package main

// now that keyboard and screen work, lets do character rendering
// first program for ROM simulating IBM MDA-inspired text mode
// using code page 437 (or at least a subset)
// drawing at 200% so each char is 16 pixels wide (incl whitespace)

// REALISATION: instead of hacking pcr cpu,
// I couldve just written a temp value to memory storing pc...
// -> still that wouldve been slower
// (DONE) With some further cpu hacking this could likely be 1 instr per charline
// by adding an instruction that jumps, then reads A but also jumps back to PCR

// R0: instr pointer between routines (but lets not invent a complete stackpointer yet)
// R5: screenpointer starting at 0x4000
var keyboardLoop = []uint16{
    0x4000, // @0x4000
    0xEC10, // D=A
    0x5,    // @screen
    0xE308, // M=D // screen = 0x4000

    // (WAIT) 4 + 809 = 813
    0x6000, // @keyboard
    0xFC10, // D=M
    0x32D,  // @WAIT
    0xE302, // D;JEQ // loop until keyboard != 0

    // if D==0x80 (ENTER) goto LINEBREAK
    0x80,   // ascii ENTER
    0xE4D0, // D=D-A
    0x34D,  // @LINEBREAK
    0xE302, // D;JEQ
    // if D==0x20 (SPACE) goto ADD1
    0x60,   // 0x80 - 0x20
    0xE090, // D=D+A
    0x349,  // @ADD1
    0xE302, // D;JEQ
    // otherwise set D back to read value
    0x20,   // ascii SPACE
    0xE090, // D=D+A

    0x2,    // @R2
    0xE308, // M=D // R2=keyboard
    0x343,  // @SCRN
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0=ref
    0x2,    // @2 (start of drawChar)
    0xEA87, // 0;JMP

    // (SCRN) 26 + 809 = 835
    // advance screen pointer, set screen pointer back up
    // if screen % 16 == 15, add 256-15=241 (linebreak) else add 1
    // x % 16 == 15 if x+1 % 16 == 0
    0x5,    // @screen
    0xFDD0, // D=M+1
    0xF,    // @15
    0xE010, // D=D&A
    0x34D,  // @LINEBREAK
    0xE302, // D;JEQ

    // (ADD1) 32 + 809 = 841
    0x5,    // @screen
    0xFDC8, // M=M+1
    // TODO: if this means we go out of bounds, linebreak instead (?)
    0x355,  // @DELAY
    0xEA87, // 0;JMP // goto DELAY

    // (LINEBREAK) 36 + 809 = 845
    // set lowest 4 bits to 0 // sets to start of line
    0x7FF0, // 0111111111110000 // TODO: first bit cant be 1, will that be a problem? why not?
    0xEC10, // D=A
    0x5,    // @screen
    0xF008, // M=D&M
    // then add 256 // jumps a char row down
    0x100,  // @256
    0xEC10, // D=A
    0x5,    // @screen
    0xF088, // M=D+M

    // (DELAY) 44 + 809 = 853
    0x6000, // @keyboard
    0xFC10, // D=M
    0x355,  // @DELAY
    0xE305, // D;JNE // loop until keyboard == 0 
    0x32D,  // @WAIT
    0xEA87, // 0;JMP // goto WAIT
}

// R0: instr pointer between routines (but lets not invent a complete stackpointer yet)
// R1: memory pointer starting at 0x1000
// R5: screenpointer starting at 0x4000
var helloworld = []uint16{
    0x0FFF, // @0x1000 - 1, since we start LOOP by incr
    0xEC10, // D=A
    0x1,    // @R1
    0xE308, // M=D // R1 = 0x1000-1
    0x4000, // @0x4000
    0xEC10, // D=A
    0x5,    // @R5
    0xE308, // M=D // R5 = 0x4000

    // (LOOP) 8 + 809 = 817
    // read value from mem
    0x1,    // @R1
    // AM=M+1 // TODO: check if not broken, do M=M+1 and A=M instead for now
    0xFDC8, // M=M+1
    0xFC20, // A=M
    0xFC10, // D=M
    // if D==0 goto END
    0x359,  // @END
    0xE302, // D;JEQ
    // if D==0x80 (ENTER) goto LINEBREAK
    0x80,   // ascii ENTER
    0xE4D0, // D=D-A
    0x34F,  // @LINEBREAK
    0xE302, // D;JEQ
    // otherwise set D back to read value
    0x80,   // ascii ENTER
    0xE090, // D=D+A

    // write char
    0x2,    // @R2
    0xE308, // M=D // R2=ascii from mem
    0x345,  // @SCRN
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0=ref
    0x2,    // @2, start of drawChar
    0xEA87, // 0;JMP

    // (SCRN) 28 + 809 = 837
    // advance screen pointer, set screen pointer back up
    // if screen % 16 == 15, add 256-15=241 (linebreak) else add 1
    // x % 16 == 15 if x+1 % 16 == 0
    0x5,    // @screen
    0xFDD0, // D=M+1
    0xF,    // @15
    0xE010, // D=D&A
    0x34F,  // @LINEBREAK
    0xE302, // D;JEQ

    // (ADD1) 34
    0x5,    // @screen
    0xFDC8, // M=M+1
    // TODO: if this means we go out of bounds, linebreak instead (?)
    0x331,   // @LOOP
    0xEA87, // 0;JMP // goto LOOP

    // (LINEBREAK) 38 + 809 = 847
    // set lowest 4 bits to 0 // sets to start of line
    0x7FF0, // 0111111111110000 // TODO: first bit cant be 1, will that be a problem? why not?
    0xEC10, // D=A
    0x5,    // @screen
    0xF008, // M=D&M
    // then add 256 // jumps a char row down
    0x100,  // @256
    0xEC10, // D=A
    0x5,    // @screen
    0xF088, // M=D+M
    0x331,   // @LOOP
    0xEA87, // 0;JMP // goto LOOP

    // (END) 48 + 809 = 857
    0x359,  // @END
    0xEA87, // 0;JMP // goto END
}

// R0: assumed to store a ref back to caller instruction
// R2: keyboard readout ascii value
// uses: R4 (@i), R5 (@screen)
// NOTES: 
// - length: 807 instructions. each pointer +2 because of jmp at start of ROM
// - screen % 16 = char column (of which there are 16)
// - ( screen >> 4 ) % 16 = drawline ( 16 per char )
// - ( screen >> 8 ) % 16 = char row (of which there are 32)
var drawChar = []uint16{

    // lets calculate D=DEF0 + (8*(R2-32)), where *8 is <<3
    0x20,   // @32
    0xEC10, // D=A
    0x2,    // @R2
    0xF1D0, // D=M-D
    0xC990, // D=D<<3 // not safe if offset too big
    0x31,   // @DEF0
    0xE090, // D=D+A
    0x4,    // @i // init location var i
    0xE308, // M=D // i=offset start

    // (LOOP) 9 -> 11
    0x4,    // @i
    0xFC20, // A=M // A=M;JMP is too risky, conflicting use of A register
    0xAA87, // 0;JMP(pcrl) // goto i, which does A=value jmp back to next instr below without touching A!
    0xEC10, // D=A
    0x5,    // @screen (we come back here after getting line of A)
    0xFC20, // A=M // A=screen
    0xE308, // M=D // mem[screen] = linevalue out of ROM

    0x10,   // @16
    0xEC10, // D=A
    0x5,    // @screen
    0xF088, // M=D+M
    0x4,    // @i
    0xFC20, // A=M // A=M;JMP is too risky, conflicting use of A register
    0xAA87, // 0;JMP(pcrl) // goto i, which does A=value jmp back to next instr below without touching A!
    0xEC10, // D=A
    0x5,    // @screen (we come back here after getting line of A)
    0xFC20, // A=M // A=screen
    0xE308, // M=D // mem[screen] = linevalue out of ROM

    // // i = i + 1
    0x4,    // @i
    0xFDC8, // M=M+1
    // // screen = screen + 16
    0x10,   // @16
    0xEC10, // D=A
    0x5,    // @screen
    //0xF098 // DM=D+M // TODO: seems to set D=D+M but M=M+M+D ?
    0xF088, // M=D+M
    0xFC10, // D=M

    // if 0<=screen%256<16, we are done
    // get screen%256 by masking, ignore last 4 bits and compare to 0
    0xF0,   // 0000000011110000
    0xE010, // D=D&A
    0x2A,   // @END
    0xE302, // D;JEQ
    0xB,    // @LOOP
    0xEA87, // 0;JMP // goto LOOP

    // (END) 40 -> 42
    // subtract 256 from @screen, setting it back
    0x100,  // @256
    0xEC10, // D=A
    0x5,    // @screen
    0xF1C8, // M=M-D
    // goto @R0 (SCRN) in keyboardloop func
    0x0,    // @R0
    0xFC20, // A=M
    0xEA87, // 0;JMP // goto SCRN

    // ------------
    // (DEF0) 47 -> 49
    // space
    0,0,0,0,0,0,0,0,
    // !
    0x0F00,
    0x3FC0,
    0x3FC0,
    0x0F00,
    0x0F00,
    0x00,
    0x0F00,
    0x00,
    // "
    0,0,0,0,0,0,0,0,
    // #
    0,0,0,0,0,0,0,0,
    // $
    0,0,0,0,0,0,0,0,
    // %
    0,0,0,0,0,0,0,0,
    // &
    0,0,0,0,0,0,0,0,
    // '
    0,0,0,0,0,0,0,0,
    // (
    0,0,0,0,0,0,0,0,
    // )
    0,0,0,0,0,0,0,0,
    // *
    0,0,0,0,0,0,0,0,
    // +
    0x00,
    0x0F00,
    0x0F00,
    0xFFF0,
    0x0F00,
    0x0F00,
    0x00,
    0x00,
    // ,
    0,0,0,0,0,0,0,0,
    // -
    0,0,0,0,0,0,0,0,
    // .
    0,0,0,0,0,0,0,0,
    // /
    0,0,0,0,0,0,0,0,
    // 0
    0x3FF0,
    0xF03C,
    0xF0FC,
    0xF3FC,
    0xFF3C,
    0xFC3C,
    0x3FF0,
    0x00,
    // 1
    0x0F00,
    0x3F00,
    0x0F00,
    0x0F00,
    0x0F00,
    0x0F00,
    0xFFF0,
    0x00,
    // 2
    0x3FC0,
    0xF0F0,
    0x00F0,
    0x0FC0,
    0x3C00,
    0xF0F0,
    0xFFF0,
    0x00,
    // 3
    0x3FC0,
    0xF0F0,
    0x00F0,
    0x0FC0,
    0x00F0,
    0xF0F0,
    0x3FC0,
    0x00,
    // 4
    0x03F0,
    0x0FF0,
    0x3CF0,
    0xF0F0,
    0xFFFC,
    0x00F0,
    0x03FC,
    0x00,
    // 5 
    0xFFF0,
    0xF000,
    0xFFC0,
    0x00F0,
    0x00F0,
    0xF0F0,
    0x3FC0,
    0x00,
    // 6
    0x0FC0,
    0x3C00,
    0xF000,
    0xFFC0,
    0xF0F0,
    0xF0F0,
    0x3FC0,
    0x00,
    // 7 
    0xFFF0,
    0xF0F0,
    0x00F0,
    0x03C0,
    0x0F00,
    0x0F00,
    0x0F00,
    0x00,
    // 8 
    0x3FC0,
    0xF0F0,
    0xF0F0,
    0x3FC0,
    0xF0F0,
    0xF0F0,
    0x3FC0,
    0x00,
    // 9 
    0x3FC0,
    0xF0F0,
    0xF0F0,
    0x3FF0,
    0x00F0,
    0x03C0,
    0x3F00,
    0x00,
    // :
    0,0,0,0,0,0,0,0,
    // ;
    0x00,
    0x0F00,
    0x0F00,
    0x00,
    0x0F00,
    0x0F00,
    0x3C00,
    0x00,
    // <
    0,0,0,0,0,0,0,0,
    // =
    0x00,
    0x00,
    0xFFF0,
    0x00,
    0x00,
    0xFFF0,
    0x00,
    0x00,
    // >
    0,0,0,0,0,0,0,0,
    // ?
    0,0,0,0,0,0,0,0,
    // @
    0x3FF0,
    0xF03C,
    0xF3FC,
    0xF3FC,
    0xF3FC,
    0xF000,
    0x3FC0,
    0x00,
    // A
    0x0F00,
    0x3FC0,
    0xF0F0,
    0xF0F0,
    0xFFF0,
    0xF0F0,
    0xF0F0,
    0x00,
    // B
    0xFFF0,
    0x3C3C,
    0x3C3C,
    0x3FF0,
    0x3C3C,
    0x3C3C,
    0xFFF0,
    0x00,
    // C
    0x0FF0,
    0x3C3C,
    0xF000,
    0xF000,
    0xF000,
    0x3C3C,
    0x0FF0,
    0x00,
    // D
    0xFFC0,
    0x3CF0,
    0x3C3C,
    0x3C3C,
    0x3C3C,
    0x3CF0,
    0xFFC0,
    0x00,
    // E
    0xFFFC,
    0x3C0C,
    0x3CC0,
    0x3FC0,
    0x3CC0,
    0x3C0C,
    0xFFFC,
    0x00,
    // F
    0xFFFC,
    0x3C0C,
    0x3CC0,
    0x3FC0,
    0x3CC0,
    0x3C00,
    0xFF00,
    0x00,
    // G
    0x0FF0,
    0x3C3C,
    0xF000,
    0xF000,
    0xF0FC,
    0x3C3C,
    0x0FFC,
    0x00,
    // H
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0xFFF0,
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0x00,
    // I
    0x3FC0,
    0x0F00,
    0x0F00,
    0x0F00,
    0x0F00,
    0x0F00,
    0x3FC0,
    0x00,
    // J
    0x03FC,
    0x00F0,
    0x00F0,
    0x00F0,
    0xF0F0,
    0xF0F0,
    0x3FC0,
    0x00,
    // K
    0xFC3C,
    0x3C3C,
    0x3CF0,
    0x3FC0,
    0x3CF0,
    0x3C3C,
    0xFC3C,
    0x00,
    // L
    0xFF00,
    0x3C00,
    0x3C00,
    0x3C00,
    0x3C00,
    0x3C3C,
    0xFFFC,
    0x00,
    // M
    0xF03C,
    0xFCFC,
    0xFFFC,
    0xFFFC,
    0xF33C,
    0xF03C,
    0xF03C,
    0x00,
    // N
    0xF03C,
    0xFC3C,
    0xFF3C,
    0xF3FC,
    0xF0FC,
    0xF03C,
    0xF03C,
    0x00,
    // O 
    0x0FC0,
    0x3CF0,
    0xF03C,
    0xF03C,
    0xF03C,
    0x3CF0,
    0x0FC0,
    0x00,
    // P
    0xFFF0,
    0x3C3C,
    0x3C3C,
    0x3FF0,
    0x3C00,
    0x3C00,
    0xFF00,
    0x00,
    // Q
    0x3FC0,
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0xF3F0,
    0x3FC0,
    0x03FC,
    0x00,
    // R
    0xFFF0,
    0x3C3C,
    0x3C3C,
    0x3FF0,
    0x3CF0,
    0x3C3C,
    0xFC3C,
    0x00,
    // S
    0x3FC0,
    0xF0F0,
    0x3C00,
    0x0F00,
    0x03C0,
    0xF0F0,
    0x3FC0,
    0x00,
    // T
    0xFFF0,
    0xCF30,
    0x0F00,
    0x0F00,
    0x0F00,
    0x0F00,
    0x3FC0,
    0x00,
    // U
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0xFFF0,
    0x00,
    // V
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0x3FC0,
    0x0F00,
    0x00,
    // W
    0xF03C,
    0xF03C,
    0xF03C,
    0xF33C,
    0xFFFC,
    0xFCFC,
    0xF03C,
    0x00,
    // X
    0xF03C,
    0xF03C,
    0x3CF0,
    0x0FC0,
    0x0FC0,
    0x3FF0,
    0xF03C,
    0x00,
    // Y
    0xF0F0,
    0xF0F0,
    0xF0F0,
    0x3FC0,
    0x0F00,
    0x0F00,
    0x3FC0,
    0x00,
    // Z
    0xFFFC,
    0xF03C,
    0xC0F0,
    0x03C0,
    0x0F0C,
    0x3C3C,
    0xFFFC,
    0x00,
    // [
    0,0,0,0,0,0,0,0,
    // \
    0,0,0,0,0,0,0,0,
    // ]
    0,0,0,0,0,0,0,0,
    // ^
    0,0,0,0,0,0,0,0,
    // _
    0,0,0,0,0,0,0,0,
    // `
    0,0,0,0,0,0,0,0,
    // a
    0,0,0,0,0,0,0,0,
    // b
    0,0,0,0,0,0,0,0,
    // c
    0,0,0,0,0,0,0,0,
    // d
    0,0,0,0,0,0,0,0,
    // e
    0,0,0,0,0,0,0,0,
    // f
    0,0,0,0,0,0,0,0,
    // g
    0,0,0,0,0,0,0,0,
    // h
    0,0,0,0,0,0,0,0,
    // i
    0,0,0,0,0,0,0,0,
    // j
    0,0,0,0,0,0,0,0,
    // k
    0,0,0,0,0,0,0,0,
    // l
    0,0,0,0,0,0,0,0,
    // m
    0,0,0,0,0,0,0,0,
    // n
    0,0,0,0,0,0,0,0,
    // o
    0,0,0,0,0,0,0,0,
    // p
    0,0,0,0,0,0,0,0,
    // q
    0,0,0,0,0,0,0,0,
    // r
    0,0,0,0,0,0,0,0,
    // s
    0,0,0,0,0,0,0,0,
    // t
    0,0,0,0,0,0,0,0,
    // u
    0,0,0,0,0,0,0,0,
    // v
    0,0,0,0,0,0,0,0,
    // w
    0,0,0,0,0,0,0,0,
    // x
    0,0,0,0,0,0,0,0,
    // y
    0,0,0,0,0,0,0,0,
    // z
    0,0,0,0,0,0,0,0,
    // {
    0,0,0,0,0,0,0,0,
    // |
    0,0,0,0,0,0,0,0,
    // }
    0,0,0,0,0,0,0,0,
    // ~
    0,0,0,0,0,0,0,0,
}



//        //0x00
//        //0x00
//        //0x00
//        //0x00
//        //0x00
//        //0x00
//        //0x00
//        //0x00
