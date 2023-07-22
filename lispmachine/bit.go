package main

type bit bool

func (b bit) String() string {
    if b {
        return "1"
    }
    return "0"
}

// we cannot declare 0 as a bit
// so we use this convenience func
func toBit(in uint16) bit {
    switch in {
    case 0:
        return false
    case 1:
        return true
    }
    panic("expected 0 or 1")
}

func nthBit(in uint16, n uint16) bit {
    if n>16 {
        panic("expected 0<=n<16")
    }
    b := in & (1 << n)
    return b != 0
}

func toBit16(in uint16) (out [16]bit) {
    for i:=0; i<16; i++ {
        out[15-i] = nthBit(in, uint16(i))
    }
    return out
}

func fromBit16(in [16]bit) uint16 {
    var out uint16
    for i:=0; i<16; i++ {
        if in[15-i] {
            out += (1 << i)
        }
    }
    return out
}

func toBit16Signed(in int16) [16]bit {
    if in>=0 {
        return toBit16(uint16(in))
    }
    return toBit16(uint16(65536+int(in)))
}

func fromBit16Signed(in [16]bit) int16 {
    negative := in[0]
    if negative {
        in = Inc16(Not16(in))
    }
    out := int16(fromBit16(in))
    if negative {
        out = out-out-out
    }
    return out
}
