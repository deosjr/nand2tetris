package main

// now that keyboard and screen work, lets do character rendering
// first program for ROM simulating IBM MDA-inspired text mode
// using code page 437 (or at least a subset)
// drawing at 200% so each char is 16 pixels wide (incl whitespace)

// REALISATION: instead of hacking pcr cpu,
// I couldve just written a temp value to memory storing pc...
// -> still that wouldve been slower

// input: char in R2 (i.e. M[0x2]) (is overwritten in the process
// uses: @i at 0x42, @screen at 0x99, R3
// TODO: still writes some garbage to screen at start?
var drawChar = []uint16{
    // @65
    0x41,
    // D=A
    0xEC10,
    // @R2
    0x2,
    // M=M-D // R2=R2-65
    0xF1C8,
    // @DEFA
    0x38,
    // D=A
    0xEC10,
    // @R3
    0x3,
    // M=D // R3=DEFA
    0xE308,

    // each char takes 16x3 ops space
    // loop R2-65 times to get D=(R2-65)*48
    // (INIT) 8
    // @R2
    0x2,
    // DM=M-1
    0xFC98,
    // @ENDINIT
    0x14,
    // D;JLE
    0xE306,
    // @R3
    0x3,
    // D=M
    0xFC10,
    // @48
    0x30,
    // D=D+A
    0xE090,
    // @R3
    0x3,
    // M=D     // R3 = R3+48
    0xE308,
    // @INIT
    0x8,
    // 0;JMP 
    0xEA87,
    // (ENDINIT) 20

    // @R3
    0x3,
    // D=M // D = start offset char
    0xFC10,

    // @i // init location var i, say 0x42
    0x42,
    // DM=D // i=offset start
    0xE318,

    // @48
    0x30,
    // D=D+A
    0xE090,
    // @R3
    0x3,
    // M=D // R3 = i+48 (3x16, charsize in ROM)
    0xE308,

    // @0x4000
    0x4000,
    // D=A
    0xEC10,
    // @screen // init location var screen, say 0x99
    0x99,
    // M=D // screen = 0x4000
    0xE308,
    // (LOOP) 32
    // @i
    0x42,
    // D=M // D=i
    0xFC10,
    // @R3
    0x3,

    // D=M-D // D= R3 - i
    0xF1D0,
    // @END
    0x36,
    // D;JEQ
    0xE302,
    // @i
    0x42,
    // A=M // A=M;JMP is too risky, conflicting use of A register
    0xFC20,
    // 0;JMP(pcrl) // goto i, which does A=value and then D=A + jmp back to next instr below
    0xAA87,
    // @screen (we come back here after getting line of A
    0x99,
    // A=M // A=screen
    // 1111 1100 0010 0000
    0xFC20,
    // M=D // mem[screen] = linevalue out of ROM
    0xE308,
    // // i = i + 3
    // @3
    0x3,
    // D=A
    0xEC10,
    // @i
    0x42,
    // M=D+M
    // 1111 0000 1000 1000
    0xF088,
    // // screen = screen + 16
    // @16
    0x10,
    // D=A
    0xEC10,
    // @screen
    0x99,
    // M=D+M
    0xF088,
    // @LOOP
    0x20,
    // 0;JMP // goto LOOP
    0xEA87,
    // ------------
    // inf loop is canonical end
    // (END) 54
    //      @END
    0x36,
    //      0;JMP 1110 1010 1000 0111
    0xEA87,
    // (DEFA) 56
    0x00, 0xAC10, 0xC7C7,
    0x00, 0xAC10, 0xC7C7,
    0x03C0, 0xAC10, 0xC7C7,
    0x03C0, 0xAC10, 0xC7C7,
    0x0FF0, 0xAC10, 0xC7C7,
    0x0FF0, 0xAC10, 0xC7C7,
    0x3C3C, 0xAC10, 0xC7C7,
    0x3C3C, 0xAC10, 0xC7C7,
    0x3C3C, 0xAC10, 0xC7C7,
    0x3C3C, 0xAC10, 0xC7C7,
    0x3FFC, 0xAC10, 0xC7C7,
    0x3FFC, 0xAC10, 0xC7C7,
    0x3C3C, 0xAC10, 0xC7C7,
    0x3C3C, 0xAC10, 0xC7C7,
    0x3C3C, 0xAC10, 0xC7C7,
    0x3C3C, 0xAC10, 0xC7C7,
}

// 100% version
//        //0x00
//  xx    //0x60
// xxxx   //0x78
//xx  xx  //0xCC
//xx  xx  //0xCC
//xxxxxx  //0xFC
//xx  xx  //0xCC
//xx  xx  //0xCC

// 200% version
//                //0x00
//                //0x00
//    xxxx        //0x0F00
//    xxxx        //0x0F00
//  xxxxxxxx      //0x3FC0
//  xxxxxxxx      //0x3FC0
//xxxx    xxxx    //0xF0F0
//xxxx    xxxx    //0xF0F0
//xxxx    xxxx    //0xF0F0
//xxxx    xxxx    //0xF0F0
//xxxxxxxxxxxx    //0xFFF0
//xxxxxxxxxxxx    //0xFFF0
//xxxx    xxxx    //0xF0F0
//xxxx    xxxx    //0xF0F0
//xxxx    xxxx    //0xF0F0
//xxxx    xxxx    //0xF0F0

// 200% version with space-offset
//                //0x00
//                //0x00
//      xxxx      //0x03C0
//      xxxx      //0x03C0
//    xxxxxxxx    //0x0FF0
//    xxxxxxxx    //0x0FF0
//  xxxx    xxxx  //0x3C3C
//  xxxx    xxxx  //0x3C3C
//  xxxx    xxxx  //0x3C3C
//  xxxx    xxxx  //0x3C3C
//  xxxxxxxxxxxx  //0x3FFC
//  xxxxxxxxxxxx  //0x3FFC
//  xxxx    xxxx  //0x3C3C
//  xxxx    xxxx  //0x3C3C
//  xxxx    xxxx  //0x3C3C
//  xxxx    xxxx  //0x3C3C

    // M=A: 1110 1100 0000 1000
// Depends on RAM[16-32] to store the lines of the A char in pixels/bits
// NOTE: would like to define chars in ROM but how to get without jumping?
// -> I can add a set+jumpback after each line read, but that quadruples storage
// -  need to make this a loop over lines first
var drawA = []uint16{
    // lets draw an A
    // @16 // address in RAM of first line of A char
    0x10,
    // D=M: 1111 1100 0001 0000
    0xFC10,
    // @SCREEN     //0x4000 in RAM 
    0x4000,
    // M=D: 1110 0011 0000 1000
    0xE308,
    // @SCREEN+16  //0x4010
    0x11,
    0xFC10,
    0x4010,
    0xE308,
    // @SCREEN+32  //0x4020
    0x12,
    0xFC10,
    0x4020,
    0xE308,
    // @SCREEN+48  //0x4030
    0x13,
    0xFC10,
    0x4030,
    0xE308,
    // @SCREEN+64  //0x4040
    0x14,
    0xFC10,
    0x4040,
    0xE308,
    // @SCREEN+etc //0x4050
    0x15,
    0xFC10,
    0x4050,
    0xE308,
    // @SCREEN+etc //0x4060
    0x16,
    0xFC10,
    0x4060,
    0xE308,
    // @SCREEN+etc //0x4070
    0x17,
    0xFC10,
    0x4070,
    0xE308, //line 31

    0x18, 0xFC10, 0x4080, 0xE308,
    0x19, 0xFC10, 0x4090, 0xE308,
    0x1A, 0xFC10, 0x40A0, 0xE308,
    0x1B, 0xFC10, 0x40B0, 0xE308,
    0x1C, 0xFC10, 0x40C0, 0xE308,
    0x1D, 0xFC10, 0x40D0, 0xE308,
    0x1E, 0xFC10, 0x40E0, 0xE308,
    0x1F, 0xFC10, 0x40F0, 0xE308,
    // inf loop is canonical end
    // (END)
    //      @END
    0x40,
    //      0;JMP 1110 1010 1000 0111
    0xEA87,
}

// char values now need to be valid A instructions so need to start with 0
// hence we change the offset of our characters from space-at-end to start
// TODO: writes some garbage to screen at start?
var drawAv2 = []uint16{
    // @32 // def A start
    0x20,
    // D=A: 1110 1100 0001 0000
    0xEC10,
    // @i // init location var i, say 0x42
    0x42,
    // M=D // i=32
    0xE308,
    // @0x4000
    0x4000,
    // D=A
    0xEC10,
    // @screen // init location var screen, say 0x99
    0x99,
    // M=D // screen = 0x4000
    0xE308,
    // (LOOP)
    // @i
    0x42,
    // D=M // D=i
    0xFC10,
    // @96
    0x60,
    // D=D-A //D=i-96 (i starts at 32 so loop 16x, each instr is 4 ops)
    // 1110 0100 1101 0000
    0xE4D0,
    // @END
    0x1E,
    // D;JGE // if (i-80)>=0 goto END
    // 1110 0011 0000 0011
    0xE303,
    // @i
    0x42,
    // A=M // A=M;JMP is too risky, conflicting use of A register
    0xFC20,
    // 0;JMP // goto i, which does A=value and then D=A + jmp back to next instr below
    0xEA87,
    // @screen
    0x99,
    // A=M // A=screen
    // 1111 1100 0010 0000
    0xFC20,
    // M=D // mem[screen] = linevalue out of ROM
    0xE308,
    // // i = i + 4
    // @4
    0x4,
    // D=A
    0xEC10,
    // @i
    0x42,
    // M=D+M
    // 1111 0000 1000 1000
    0xF088,
    // // screen = screen + 16
    // @16
    0x10,
    // D=A
    0xEC10,
    // @screen
    0x99,
    // M=D+M
    0xF088,
    // @LOOP
    0x8,
    // 0;JMP // goto LOOP
    0xEA87,
    // ------------
    // inf loop is canonical end
    // (END)
    //      @END
    0x1E,
    //      0;JMP 1110 1010 1000 0111
    0xEA87,
    // A
    0x00,
    // D=A, A=17, 0;JMP
    0xEC10, 0x11, 0xEA87,
    0x00,
    0xEC10, 0x11, 0xEA87,
    0x03C0,
    0xEC10, 0x11, 0xEA87,
    0x03C0,
    0xEC10, 0x11, 0xEA87,
    0x0FF0,
    0xEC10, 0x11, 0xEA87,
    0x0FF0,
    0xEC10, 0x11, 0xEA87,
    0x3C3C,
    0xEC10, 0x11, 0xEA87,
    0x3C3C,
    0xEC10, 0x11, 0xEA87,
    0x3C3C,
    0xEC10, 0x11, 0xEA87,
    0x3C3C,
    0xEC10, 0x11, 0xEA87,
    0x3FFC,
    0xEC10, 0x11, 0xEA87,
    0x3FFC,
    0xEC10, 0x11, 0xEA87,
    0x3C3C,
    0xEC10, 0x11, 0xEA87,
    0x3C3C,
    0xEC10, 0x11, 0xEA87,
    0x3C3C,
    0xEC10, 0x11, 0xEA87,
    0x3C3C,
    0xEC10, 0x11, 0xEA87,
}

// use pcregistercpu to reduce ops when retrieving char info
var drawAv3 = []uint16{
    // @DEFA
    0x20,
    // D=A: 1110 1100 0001 0000
    0xEC10,
    // @i // init location var i, say 0x42
    0x42,
    // M=D // i=32
    0xE308,
    // @0x4000
    0x4000,
    // D=A
    0xEC10,
    // @screen // init location var screen, say 0x99
    0x99,
    // M=D // screen = 0x4000
    0xE308,
    // (LOOP)
    // @i
    0x42,
    // D=M // D=i
    0xFC10,
    // @80
    0x50,
    // D=D-A //D=i-80 (i starts at 32 so loop 16x, each instr is 3 ops)
    // 1110 0100 1101 0000
    0xE4D0,
    // @END
    0x1E,
    // D;JGE // if (i-80)>=0 goto END
    // 1110 0011 0000 0011
    0xE303,
    // @i
    0x42,
    // A=M // A=M;JMP is too risky, conflicting use of A register
    0xFC20,
    // 0;JMP(pcrl) // goto i, which does A=value and then D=A + jmp back to next instr below
    0xAA87,
    // @screen (we come back here after getting line of A
    0x99,
    // A=M // A=screen
    // 1111 1100 0010 0000
    0xFC20,
    // M=D // mem[screen] = linevalue out of ROM
    0xE308,
    // // i = i + 3
    // @3
    0x3,
    // D=A
    0xEC10,
    // @i
    0x42,
    // M=D+M
    // 1111 0000 1000 1000
    0xF088,
    // // screen = screen + 16
    // @16
    0x10,
    // D=A
    0xEC10,
    // @screen
    0x99,
    // M=D+M
    0xF088,
    // @LOOP
    0x8,
    // 0;JMP // goto LOOP
    0xEA87,
    // ------------
    // inf loop is canonical end
    // (END)
    //      @END
    0x1E,
    //      0;JMP 1110 1010 1000 0111
    0xEA87,
    //(DEFA)
    0x00,
    // D=A(pcrl), PCR+1;JMPPCR
    0xAC10, 0xC7C7,
    0x00,
    0xAC10, 0xC7C7,
    0x03C0,
    0xAC10, 0xC7C7,
    0x03C0,
    0xAC10, 0xC7C7,
    0x0FF0,
    0xAC10, 0xC7C7,
    0x0FF0,
    0xAC10, 0xC7C7,
    0x3C3C,
    0xAC10, 0xC7C7,
    0x3C3C,
    0xAC10, 0xC7C7,
    0x3C3C,
    0xAC10, 0xC7C7,
    0x3C3C,
    0xAC10, 0xC7C7,
    0x3FFC,
    0xAC10, 0xC7C7,
    0x3FFC,
    0xAC10, 0xC7C7,
    0x3C3C,
    0xAC10, 0xC7C7,
    0x3C3C,
    0xAC10, 0xC7C7,
    0x3C3C,
    0xAC10, 0xC7C7,
    0x3C3C,
    0xAC10, 0xC7C7,
}
