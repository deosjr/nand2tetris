package main

import (
    "math/rand"
    "testing"
    "time"
)

func TestHalfAdder(t *testing.T) {
    for _, tt := range []struct{
        a uint16
        b uint16
        sum uint16
        carry uint16
    }{
        {a:0, b:0, carry:0, sum:0},
        {a:0, b:1, carry:0, sum:1},
        {a:1, b:0, carry:0, sum:1},
        {a:1, b:1, carry:1, sum:0},
    }{
        gotCarry, gotSum := HalfAdder(toBit(tt.a), toBit(tt.b))
        if gotCarry != toBit(tt.carry) {
            t.Errorf("gotCarry %t but want %t", gotCarry, toBit(tt.carry))
        }
        if gotSum != toBit(tt.sum) {
            t.Errorf("gotSum %t but want %t", gotSum, toBit(tt.sum))
        }
    }
}

func TestFullAdder(t *testing.T) {
    for _, tt := range []struct{
        a uint16
        b uint16
        c uint16
        sum uint16
        carry uint16
    }{
        {a:0, b:0, c:0, carry:0, sum:0},
        {a:0, b:0, c:1, carry:0, sum:1},
        {a:0, b:1, c:0, carry:0, sum:1},
        {a:0, b:1, c:1, carry:1, sum:0},
        {a:1, b:0, c:0, carry:0, sum:1},
        {a:1, b:0, c:1, carry:1, sum:0},
        {a:1, b:1, c:0, carry:1, sum:0},
        {a:1, b:1, c:1, carry:1, sum:1},
    }{
        gotCarry, gotSum := FullAdder(toBit(tt.a), toBit(tt.b), toBit(tt.c))
        if gotCarry != toBit(tt.carry) {
            t.Errorf("gotCarry %t but want %t", gotCarry, toBit(tt.carry))
        }
        if gotSum != toBit(tt.sum) {
            t.Errorf("gotSum %t but want %t", gotSum, toBit(tt.sum))
        }
    }
}

// fuzz testing with ten random inputs
func TestAdd16(t *testing.T) {
    rand.Seed(time.Now().UnixNano())
    for i:=0; i<10; i++ {
        a, b := uint16(rand.Intn(65536)), uint16(rand.Intn(65536))
        out := Add16(toBit16(a), toBit16(b))
        got := fromBit16(out)
        want := a + b
        if got != want {
            t.Errorf("expected %d but got %d", want, got)
        }
    }
}

func TestAlu(t *testing.T) {
    rand.Seed(time.Now().UnixNano())
    for _, tt := range []struct{
        c [6]uint16 // controls: zx, nx, zy, ny, f, no
        wantFunc func(a, b uint16) int16
    }{
        {
            c: [6]uint16{1,0,1,0,1,0},
            wantFunc: func(a, b uint16) int16 { return 0 },
        },
        {
            c: [6]uint16{1,1,1,1,1,1},
            wantFunc: func(a, b uint16) int16 { return 1 },
        },
        {
            c: [6]uint16{1,1,1,0,1,0},
            wantFunc: func(a, b uint16) int16 { return -1 },
        },
        {
            c: [6]uint16{0,0,1,1,0,0},
            wantFunc: func(a, b uint16) int16 { return int16(a) },
        },
        {
            c: [6]uint16{1,1,0,0,0,0},
            wantFunc: func(a, b uint16) int16 { return int16(b) },
        },
        {
            c: [6]uint16{0,0,1,1,1,1},
            wantFunc: func(a, b uint16) int16 { return int16(-a) },
        },
        {
            c: [6]uint16{1,1,0,0,1,1},
            wantFunc: func(a, b uint16) int16 { return int16(-b) },
        },
        {
            c: [6]uint16{0,0,0,0,1,0},
            wantFunc: func(a, b uint16) int16 { return int16(a + b) },
        },
        {
            c: [6]uint16{0,0,0,0,0,0},
            wantFunc: func(a, b uint16) int16 { return int16(a & b) },
        },
        {
            c: [6]uint16{0,1,0,1,0,1},
            wantFunc: func(a, b uint16) int16 { return int16(a | b) },
        },
    }{
        for i:=0; i<10; i++ {
            a, b := uint16(rand.Intn(65536)), uint16(rand.Intn(65536))
            zx := toBit(tt.c[0])
            nx := toBit(tt.c[1])
            zy := toBit(tt.c[2])
            ny := toBit(tt.c[3])
            f  := toBit(tt.c[4])
            no := toBit(tt.c[5])
            out, zr, ng := Alu(toBit16(a), toBit16(b), zx, nx, zy, ny, f, no)
            got := fromBit16Signed(out)
            want := tt.wantFunc(a, b)
            if got != want {
                t.Errorf("expected %d but got %d", want, got)
            }
            if zr != (got==0) {
                t.Errorf("expected zr to match out: %t %d", zr, got)
            }
            if ng != (got<0) {
                t.Errorf("expected ng to match out: %t %d", ng, got)
            }
        }
    }
}
