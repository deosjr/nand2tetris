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

// barrel shifter, does a circular left shift of size n
// a right shift is just a 16-n left shift
// if you want a logical shift, mask the input beforehand
func Shift(in [16]bit, n [4]bit) [16]bit {
    // we shift by 8 if n[0] is true, then by 4 depending on n[1] etc
    // this could be a loop but lets write it out to show theres no magic, just wires
    // stores in intermediate arrays for readability and not to calculate values multiple times
    s8 := [16]bit{
        Mux(in[0], in[8], n[0]),
        Mux(in[1], in[9], n[0]),
        Mux(in[2], in[10], n[0]),
        Mux(in[3], in[11], n[0]),
        Mux(in[4], in[12], n[0]),
        Mux(in[5], in[13], n[0]),
        Mux(in[6], in[14], n[0]),
        Mux(in[7], in[15], n[0]),
        Mux(in[8], in[0], n[0]),
        Mux(in[9], in[1], n[0]),
        Mux(in[10], in[2], n[0]),
        Mux(in[11], in[3], n[0]),
        Mux(in[12], in[4], n[0]),
        Mux(in[13], in[5], n[0]),
        Mux(in[14], in[6], n[0]),
        Mux(in[15], in[7], n[0]),
    }
    s4 := [16]bit{
        Mux(s8[0], s8[4], n[1]),
        Mux(s8[1], s8[5], n[1]),
        Mux(s8[2], s8[6], n[1]),
        Mux(s8[3], s8[7], n[1]),
        Mux(s8[4], s8[8], n[1]),
        Mux(s8[5], s8[9], n[1]),
        Mux(s8[6], s8[10], n[1]),
        Mux(s8[7], s8[11], n[1]),
        Mux(s8[8], s8[12], n[1]),
        Mux(s8[9], s8[13], n[1]),
        Mux(s8[10], s8[14], n[1]),
        Mux(s8[11], s8[15], n[1]),
        Mux(s8[12], s8[0], n[1]),
        Mux(s8[13], s8[1], n[1]),
        Mux(s8[14], s8[2], n[1]),
        Mux(s8[15], s8[3], n[1]),
    }
    s2 := [16]bit{
        Mux(s4[0], s4[2], n[2]),
        Mux(s4[1], s4[3], n[2]),
        Mux(s4[2], s4[4], n[2]),
        Mux(s4[3], s4[5], n[2]),
        Mux(s4[4], s4[6], n[2]),
        Mux(s4[5], s4[7], n[2]),
        Mux(s4[6], s4[8], n[2]),
        Mux(s4[7], s4[9], n[2]),
        Mux(s4[8], s4[10], n[2]),
        Mux(s4[9], s4[11], n[2]),
        Mux(s4[10], s4[12], n[2]),
        Mux(s4[11], s4[13], n[2]),
        Mux(s4[12], s4[14], n[2]),
        Mux(s4[13], s4[15], n[2]),
        Mux(s4[14], s4[0], n[2]),
        Mux(s4[15], s4[1], n[2]),
    }
    return [16]bit{
        Mux(s2[0], s2[1], n[3]),
        Mux(s2[1], s2[2], n[3]),
        Mux(s2[2], s2[3], n[3]),
        Mux(s2[3], s2[4], n[3]),
        Mux(s2[4], s2[5], n[3]),
        Mux(s2[5], s2[6], n[3]),
        Mux(s2[6], s2[7], n[3]),
        Mux(s2[7], s2[8], n[3]),
        Mux(s2[8], s2[9], n[3]),
        Mux(s2[9], s2[10], n[3]),
        Mux(s2[10], s2[11], n[3]),
        Mux(s2[11], s2[12], n[3]),
        Mux(s2[12], s2[13], n[3]),
        Mux(s2[13], s2[14], n[3]),
        Mux(s2[14], s2[15], n[3]),
        Mux(s2[15], s2[0], n[3]),
    }
}
