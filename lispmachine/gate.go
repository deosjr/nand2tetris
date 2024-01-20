package main

// primitive, uses builtin go
func Nand(a, b bit) bit {
    return !(a && b)
}

// rest is not primitive, derives from Nand
func Not(in bit) bit {
    return Nand(in, in)
}

func And(a, b bit) bit {
    return Not(Nand(a, b))
}

func Or(a, b bit) bit {
    return Nand(Not(a), Not(b))
}

func Xor(a, b bit) bit {
    // without this declaration we would have to calculate nandab twice
    nandab := Nand(a, b)
    return Nand(Nand(a, nandab), Nand(b, nandab))
}

func Mux(a, b, sel bit) bit {
    return Or(And(a, Not(sel)), And(b, sel))
}

func DMux(in, sel bit) (bit, bit) {
    return And(in, Not(sel)), And(in, sel)
}

// [16]bit is uint16, probably faster
// for-loops like this are serial connections between gates
// other control statements like if/while are not allowed

func Not16(in [16]bit) (out [16]bit) {
    for i:=0; i<16; i++ {
        out[i] = Not(in[i])
    }
    return out
}

func And16(a, b [16]bit) (out [16]bit) {
    for i:=0; i<16; i++ {
        out[i] = And(a[i], b[i])
    }
    return out
}

func Or16(a, b [16]bit) (out [16]bit) {
    for i:=0; i<16; i++ {
        out[i] = Or(a[i], b[i])
    }
    return out
}

func Mux16(a, b [16]bit, sel bit) (out [16]bit) {
    for i:=0; i<16; i++ {
        out[i] = Mux(a[i], b[i], sel)
    }
    return out
}

func Or8Way(in [8]bit) bit {
    out := in[0]
    for i:=1; i<8; i++ {
        out = Or(out, in[i])
    }
    return out
}

func Or16Way(in [16]bit) bit {
    out := in[0]
    for i:=1; i<16; i++ {
        out = Or(out, in[i])
    }
    return out
}

func And16Way(in [16]bit) bit {
    out := in[0]
    for i:=1; i<16; i++ {
        out = And(out, in[i])
    }
    return out
}

func Mux4Way16(a, b, c, d [16]bit, sel [2]bit) [16]bit {
    return Mux16(Mux16(a, b, sel[0]), Mux16(c, d, sel[0]), sel[1])
}

func Mux8Way16(a, b, c, d, e, f, g, h [16]bit, sel [3]bit) [16]bit {
    sel2 := [2]bit{sel[0], sel[1]}
    return Mux16(Mux4Way16(a, b, c, d, sel2), Mux4Way16(e, f, g, h, sel2), sel[2])
}

func DMux4Way(in bit, sel [2]bit) (bit, bit, bit, bit) {
    a, b := DMux(in, sel[0])
    c, d := DMux(in, sel[0])
    notsel := Not(sel[1])
    return And(a, notsel), And(b, notsel), And(c, sel[1]), And(d, sel[1])
}

func DMux8Way(in bit, sel [3]bit) (bit, bit, bit, bit, bit, bit, bit, bit) {
    sel2 := [2]bit{sel[0], sel[1]}
    a, b, c, d := DMux4Way(in, sel2)
    e, f, g, h := DMux4Way(in, sel2)
    notsel := Not(sel[2])
    return And(a, notsel), And(b, notsel), And(c, notsel), And(d, notsel),
            And(e, sel[2]), And(f, sel[2]), And(g, sel[2]), And(h, sel[2])
}
