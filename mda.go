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

    // (WAIT) 4
    0x6000, // @keyboard
    0xFC10, // D=M
    0x4,    // @WAIT
    0xE302, // D;JEQ // loop until keyboard != 0

    // if D==0x80 (ENTER) goto LINEBREAK
    0x80,   // ascii ENTER
    0xE4D0, // D=D-A
    0x24,   // @LINEBREAK
    0xE302, // D;JEQ
    // if D==0x20 (SPACE) goto ADD1
    0x60,   // 0x80 - 0x20
    0xE090, // D=D+A
    0x20,   // @ADD1
    0xE302, // D;JEQ
    // otherwise set D back to read value
    0x20,   // ascii SPACE
    0xE090, // D=D+A

    0x2,    // @R2
    0xE308, // M=D // R2=keyboard
    0x1A,   // @SCRN
    0xEC10, // D=A
    0x0,    // @R0
    0xE308, // M=D // R0=ref
    0x34,   // @52, first instr after this func (assumed start of drawChar)
    0xEA87, // 0;JMP

    // (SCRN) 26
    // advance screen pointer, set screen pointer back up
    // if screen % 16 == 15, add 256-15=241 (linebreak) else add 1
    // x % 16 == 15 if x+1 % 16 == 0
    0x5,    // @screen
    0xFDD0, // D=M+1
    0xF,    // @15
    0xE010, // D=D&A
    0x24,   // @LINEBREAK
    0xE302, // D;JEQ

    // (ADD1) 32
    0x5,    // @screen
    0xFDC8, // M=M+1
    0x2E,   // @DELAY
    0xEA87, // 0;JMP // goto DELAY

    // (LINEBREAK) 36
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
    // TODO: these two lines not needed, falls through
    0x2E,   // @DELAY
    0xEA87, // 0;JMP // goto DELAY

    // (DELAY) 46 // wait until keyboard == 0
    0x6000, // @keyboard
    0xFC10, // D=M
    0x2E,   // @DELAY
    0xE305, // D;JNE // loop until keyboard != 0
    0x4,    // @WAIT
    0xEA87, // 0;JMP // goto WAIT
}

// R0: assumed to store a ref back to caller instruction
// R2: keyboard readout ascii value
// uses: R3 (temp), R4 (@i), R5 (@screen)
// NOTES: 
// - since drawChar is directly after keyboardLoop, we need to add 52 to each pointer
// - screen % 16 = char column (of which there are 16)
// - ( screen >> 4 ) % 16 = drawline ( 16 per char )
// - ( screen >> 8 ) % 16 = char row (of which there are 32)
// - I can do some shifts by masking and subtracting (?)
// - Im tempted to implement a barrel shifter circuit..
var drawChar = []uint16{

    0x2,    // @R2
    0xFC10, // D=M
    0x41,   // @65
    0xE4D0, // D=D-A
    0x2,    // @R2
    0xE308, // M=D // R2=keyboard-65
    0x69,   // @DEFA
    0xEC10, // D=A
    0x3,    // @R3
    0xE308, // M=D // R3=DEFA

    // each char takes 16 ops space (TODO: could even think about only 8)
    // loop R2-65 times to get D=(R2-65)*16
    // TODO: *16 is the same as << 4
    // (INIT) 10 -> 62
    0x2,    // @R2
    0xFC98, // DM=M-1
    0x4A,   // @ENDINIT
    0xE304, // D;JLT
    0x3,    // @R3
    0xFC10, // D=M
    0x10,   // @16
    0xE090, // D=D+A
    0x3,    // @R3
    0xE308, // M=D     // R3 = R3+16
    0x3E,   // @INIT
    0xEA87, // 0;JMP

    // (ENDINIT) 22 -> 74 // now R3 = DEFA + (keyboard-65)*16
    0x3,    // @R3
    0xFC10, // D=M // D = start offset char
    0x4,    // @i // init location var i
    0xE308, // M=D // i=offset start

    // (LOOP) 26 -> 78
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
    0x62,   // @END
    0xE302, // D;JEQ
    0x4E,   // @LOOP
    0xEA87, // 0;JMP // goto LOOP

    // (END) 46 -> 98
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
    // (DEFA) 53 -> 105
    0x00,
    0x00,
    0x03C0,
    0x03C0,
    0x0FF0,
    0x0FF0,
    0x3C3C,
    0x3C3C,
    0x3C3C,
    0x3C3C,
    0x3FFC,
    0x3FFC,
    0x3C3C,
    0x3C3C,
    0x3C3C,
    0x3C3C,
    // B
    0x00,
    0x00,
    0x3FFC,
    0x3FFC,
    0x0F0F,
    0x0F0F,
    0x0F0F,
    0x0F0F,
    0x0FFC,
    0x0FFC,
    0x0F0F,
    0x0F0F,
    0x0F0F,
    0x0F0F,
    0x3FFC,
    0x3FFC,
    // C
    0x00,
    0x00,
    0x03FC,
    0x03FC,
    0x0F0F,
    0x0F0F,
    0x3C00,
    0x3C00,
    0x3C00,
    0x3C00,
    0x3C00,
    0x3C00,
    0x0F0F,
    0x0F0F,
    0x03FC,
    0x03FC,
}

// 200% version with space-offset
//                //0x00
//                //0x00
//      xxxxxxxx  //0x03FC
//      xxxxxxxx  //0x03FC
//    xxxx    xxxx//0x0F0F
//    xxxx    xxxx//0x0F0F
//  xxxx          //0x3C00
//  xxxx          //0x3C00
//  xxxx          //0x3C00
//  xxxx          //0x3C00
//  xxxx          //0x3C00
//  xxxx          //0x3C00
//    xxxx    xxxx//0x0F0F
//    xxxx    xxxx//0x0F0F
//      xxxxxxxx  //0x03FC
//      xxxxxxxx  //0x03FC
