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

// uses: R2, R3 (temp), R4 (@i), R5 (@screen)
//
// NOTES: 
// - screen % 16 = char column (of which there are 16)
// - ( screen >> 4 ) % 16 = drawline ( 16 per char )
// - ( screen >> 8 ) % 16 = char row (of which there are 32)
// - I can do some shifts by masking and subtracting (?)
// - Im tempted to implement a barrel shifter circuit..
var drawChar = []uint16{
    // @0x4000
    0x4000,
    // D=A
    0xEC10,
    // @screen
    0x5,
    // M=D // screen = 0x4000
    0xE308,

    // (WAIT)
    // @keyboard
    0x6000,
    // D=M
    0xFC10,
    // @WAIT
    0x4,
    // D;JEQ // loop until keyboard != 0
    0xE302,

    // @65
    0x41,
    // D=D-A
    0xE4D0,
    // @R2
    0x2,
    // M=D // R2=keyboard-65
    0xE308,
    // @DEFA
    0x4A,
    // D=A
    0xEC10,
    // @R3
    0x3,
    // M=D // R3=DEFA
    0xE308,

    // each char takes 16 ops space (TODO: could even think about only 8)
    // loop R2-65 times to get D=(R2-65)*16
    // TODO: *16 is the same as << 4
    // (INIT) 16
    // @R2
    0x2,
    // DM=M-1
    0xFC98,
    // @ENDINIT
    0x1C,
    // D;JLT
    0xE304,
    // @R3
    0x3,
    // D=M
    0xFC10,
    // @16
    0x10,
    // D=D+A
    0xE090,
    // @R3
    0x3,
    // M=D     // R3 = R3+16
    0xE308,
    // @INIT
    0x10,
    // 0;JMP 
    0xEA87,
    // (ENDINIT) 28 // now R3 = DEFA + (keyboard-65)*16

    // @R3
    0x3,
    // D=M // D = start offset char
    0xFC10,

    // @i // init location var i
    0x4,
    // M=D // i=offset start
    0xE308,

    // (LOOP) 32
    // @i
    0x4,
    // A=M // A=M;JMP is too risky, conflicting use of A register
    0xFC20,
    // 0;JMP(pcrl) // goto i, which does A=value jmp back to next instr below without touching A!
    0xAA87,
    // D=A
    0xEC10,
    // @screen (we come back here after getting line of A)
    0x5,
    // A=M // A=screen
    0xFC20,
    // M=D // mem[screen] = linevalue out of ROM
    0xE308,
    // // i = i + 1
    // @i
    0x4,
    // M=M+1
    0xFDC8,
    // // screen = screen + 16
    // @16
    0x10,
    // D=A
    0xEC10,
    // @screen
    0x5,
    // DM=D+M // TODO: seems to set D=D+M but M=M+M+D ?
    //0xF098,
    //M=D+M
    0xF088,
    //D=M
    0xFC10,
    // if 0<=screen%256<16, we are done
    // get screen%256 by masking, ignore last 4 bits and compare to 0
    0xF0,
    // D=D&A
    0xE010,

    // @END
    0x34,
    // D;JEQ
    0xE302,
    // @LOOP
    0x20,
    // 0;JMP // goto LOOP
    0xEA87,
    // (END) 52
    // advance screen pointer
    // set screen pointer back up
    // @screen
    0x5,
    // if screen % 16 == 15, add 256-15=241 (linebreak) else add 1
    // x % 16 == 15 if x+1 % 16 == 0
    // D=M+1
    0xFDD0,
    // @15
    0xF,
    // D=D&A
    0xE010,
    // @LINEBREAK
    0x3E,
    // D;JEQ
    0xE302,
    // (ADD1)
    0x5,
    // M=M+1
    0xFDC8,
    // GOTO BACK
    0x44,
    0xEA87,
    // (LINEBREAK) 62
    // set lowest 4 bits to 0
    0x7FF0,
    // D=A
    0xEC10,
    0x5,
    // M=D&M
    0xF008,
    // then add 256 (but those would be removed again in back so just jump to WAIT)
    0x4,
    0xEA87,
    // (BACK) // -256 and goto WAIT // 68
    // @256 
    0x100,
    // D=A
    0xEC10,
    0x5,
    // M=M-D
    0xF1C8,
    0x4,
    0xEA87,
    // ------------
    // (DEFA) 74
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
