package main

func HalfAdder(a, b bit) (carry, sum bit) {
    return And(a,b), Xor(a, b)
}

func FullAdder(a, b, c bit) (carry, sum bit) {
    c1, s1 := HalfAdder(a, b)
    c2, s2 := HalfAdder(s1, c)
    return Or(c1, c2), s2
}

func Add16(a, b [16]bit) (out [16]bit) {
    var c bit
    for i:=0; i<16; i++ {
        carry, sum := FullAdder(a[15-i], b[15-i], c)
        out[15-i] = sum
        c = carry
    }
    return out
}

func Inc16(in [16]bit) [16]bit {
    return Add16(in, toBit16(1))
}

func Alu(x, y [16]bit, zx, nx, zy, ny, f, no bit) (out [16]bit, zr, ng bit) {
    zxo := Mux16(x, toBit16(0), zx)
    nxo := Mux16(zxo, Not16(zxo), nx)
    zyo := Mux16(y, toBit16(0), zy)
    nyo := Mux16(zyo, Not16(zyo), ny)
    fo := Mux16(And16(nxo, nyo), Add16(nxo, nyo), f)
    out = Mux16(fo, Not16(fo), no)
    out8 := [8]bit{out[0],out[1],out[2],out[3],out[4],out[5],out[6],out[7]}
    out16 := [8]bit{out[8],out[9],out[10],out[11],out[12],out[13],out[14],out[15]}
    zr = Not(Or(Or8Way(out8), Or8Way(out16)))
    ng = out[0]
    return
}
