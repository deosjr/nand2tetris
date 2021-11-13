package main

import (
    "testing"
)

func TestNand(t *testing.T) {
    for _, tt := range []struct{
        a uint16
        b uint16
        want uint16
    }{
        {a:0, b:0, want:1},
        {a:0, b:1, want:1},
        {a:1, b:0, want:1},
        {a:1, b:1, want:0},
    }{
        got := Nand(toBit(tt.a), toBit(tt.b))
        if got != toBit(tt.want) {
            t.Errorf("got %t but want %t", got, toBit(tt.want))
        }
    }
}

func TestNot(t *testing.T) {
    for _, tt := range []struct{
        in uint16
        want uint16
    }{
        {in:0, want:1},
        {in:1, want:0},
    }{
        got := Not(toBit(tt.in))
        if got != toBit(tt.want) {
            t.Errorf("got %t but want %t", got, toBit(tt.want))
        }
    }
}

func TestAnd(t *testing.T) {
    for _, tt := range []struct{
        a uint16
        b uint16
        want uint16
    }{
        {a:0, b:0, want:0},
        {a:0, b:1, want:0},
        {a:1, b:0, want:0},
        {a:1, b:1, want:1},
    }{
        got := And(toBit(tt.a), toBit(tt.b))
        if got != toBit(tt.want) {
            t.Errorf("got %t but want %t", got, toBit(tt.want))
        }
    }
}

func TestOr(t *testing.T) {
    for _, tt := range []struct{
        a uint16
        b uint16
        want uint16
    }{
        {a:0, b:0, want:0},
        {a:0, b:1, want:1},
        {a:1, b:0, want:1},
        {a:1, b:1, want:1},
    }{
        got := Or(toBit(tt.a), toBit(tt.b))
        if got != toBit(tt.want) {
            t.Errorf("got %t but want %t", got, toBit(tt.want))
        }
    }
}

func TestXor(t *testing.T) {
    for _, tt := range []struct{
        a uint16
        b uint16
        want uint16
    }{
        {a:0, b:0, want:0},
        {a:0, b:1, want:1},
        {a:1, b:0, want:1},
        {a:1, b:1, want:0},
    }{
        got := Xor(toBit(tt.a), toBit(tt.b))
        if got != toBit(tt.want) {
            t.Errorf("got %t but want %t", got, toBit(tt.want))
        }
    }
}

func TestMux(t *testing.T) {
    for _, tt := range []struct{
        a uint16
        b uint16
        sel uint16
        want uint16
    }{
        {a:0, b:0, sel:0, want:0},
        {a:0, b:1, sel:0, want:0},
        {a:1, b:0, sel:0, want:1},
        {a:1, b:1, sel:0, want:1},
        {a:0, b:0, sel:1, want:0},
        {a:0, b:1, sel:1, want:1},
        {a:1, b:0, sel:1, want:0},
        {a:1, b:1, sel:1, want:1},
    }{
        got := Mux(toBit(tt.a), toBit(tt.b), toBit(tt.sel))
        if got != toBit(tt.want) {
            t.Errorf("got %t but want %t", got, toBit(tt.want))
        }
    }
}

func TestDMux(t *testing.T) {
    for _, tt := range []struct{
        in uint16
        sel uint16
        wantA uint16
        wantB uint16
    }{
        {in:0, sel:0, wantA:0, wantB:0},
        {in:1, sel:0, wantA:1, wantB:0},
        {in:0, sel:1, wantA:0, wantB:0},
        {in:1, sel:1, wantA:0, wantB:1},
    }{
        gotA, gotB := DMux(toBit(tt.in), toBit(tt.sel))
        if gotA != toBit(tt.wantA) {
            t.Errorf("gotA %t but wantA %t", gotA, toBit(tt.wantA))
        }
        if gotB != toBit(tt.wantB) {
            t.Errorf("gotB %t but wantB %t", gotB, toBit(tt.wantB))
        }
    }
}

func TestMux4Way16(t *testing.T) {
    for _, tt := range []struct{
        a uint16
        b uint16
        c uint16
        d uint16
        sel [2]bit
        want uint16
    }{
        {a:42, b:9485, c:458, d:9, sel:[2]bit{false, false}, want:42},
        {a:42, b:9485, c:458, d:9, sel:[2]bit{true, false}, want:9485},
        {a:42, b:9485, c:458, d:9, sel:[2]bit{false, true}, want:458},
        {a:42, b:9485, c:458, d:9, sel:[2]bit{true, true}, want:9},
    }{
        got := Mux4Way16(toBit16(tt.a), toBit16(tt.b), toBit16(tt.c), toBit16(tt.d), tt.sel)
        if fromBit16(got) != tt.want {
            t.Errorf("got %d but want %d", fromBit16(got), tt.want)
        }
    }
}
