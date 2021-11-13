package main

type RAM8 struct {
    In chan [16]bit
    Out chan [16]bit
    Address chan [3]bit
    Load chan bit
    registers []Register
}

func NewRAM8(in, out chan [16]bit, addr chan [3]bit, load chan bit) RAM8 {
    ins := fanout16(in, 8)
    registers := make([]Register, 8)
    for i:=0; i<8; i++ {
        rout := make(chan [16]bit, 1)
        rload := make(chan bit, 1)
        registers[i] = NewRegister(ins[i], rout, rload)
    }
    return RAM8{
        In: in,
        Out: out,
        Address: addr,
        Load: load,
        registers: registers,
    }
}

func (ram8 RAM8) Run(clock chan bit) {
    go Run(clock, ram8.registers[0], ram8.registers[1], ram8.registers[2], ram8.registers[3],
        ram8.registers[4], ram8.registers[5], ram8.registers[6], ram8.registers[7])
    addrs := fanout3(ram8.Address, 2)
    // combinational in
    go func(addr chan [3]bit) {
        var address [3]bit
        var load bit
        for {
            select {
            case b := <-addr:
                address = b
            case b := <-ram8.Load:
                load = b
            }
            a,b,c,d,e,f,g,h := DMux8Way(load, address)
            for i, x := range [8]bit{a,b,c,d,e,f,g,h} {
                ram8.registers[i].Load <- x
            }
        }
    }(addrs[0])
    // combinational out
    out8chan := make(chan [8][16]bit, 1)
    gather16x8([8]chan [16]bit{
        ram8.registers[0].Out,
        ram8.registers[1].Out,
        ram8.registers[2].Out,
        ram8.registers[3].Out,
        ram8.registers[4].Out,
        ram8.registers[5].Out,
        ram8.registers[6].Out,
        ram8.registers[7].Out,
    }, out8chan)
    var address [3]bit
    var o [8][16]bit
    for {
        select {
        case b := <-addrs[1]:
            address = b
        case b := <-out8chan:
            o = b
        }
        ram8.Out<-Mux8Way16(o[0], o[1], o[2], o[3], o[4], o[5], o[6], o[7], address)
    }
}

type RAM64 struct {
    In chan [16]bit
    Out chan [16]bit
    Address chan [6]bit
    Load chan bit
    ram8s []RAM8
}

func NewRAM64(in, out chan [16]bit, addr chan [6]bit, load chan bit) RAM64 {
    ins := fanout16(in, 8)
    ram8s := make([]RAM8, 8)
    for i:=0; i<8; i++ {
        rout := make(chan [16]bit, 1)
        raddr := make(chan [3]bit, 1)
        rload := make(chan bit, 1)
        ram8s[i] = NewRAM8(ins[i], rout, raddr, rload)
    }
    return RAM64{
        In: in,
        Out: out,
        Address: addr,
        Load: load,
        ram8s: ram8s,
    }
}

func (ram64 RAM64) Run(clock chan bit) {
    go Run(clock, ram64.ram8s[0], ram64.ram8s[1], ram64.ram8s[2], ram64.ram8s[3],
        ram64.ram8s[4], ram64.ram8s[5], ram64.ram8s[6], ram64.ram8s[7])
    addrs := fanout6(ram64.Address, 2)
    // combinational in
    go func(addr chan [6]bit) {
        var addrHigh [3]bit
        var addrLow [3]bit
        var load bit
        for {
            select {
            case b := <-addr:
                addrHigh = [3]bit{b[0], b[1], b[2]}
                addrLow = [3]bit{b[3], b[4], b[5]}
                for _, r := range ram64.ram8s {
                    r.Address <- addrLow
                }
            case b := <-ram64.Load:
                load = b
            }
            a,b,c,d,e,f,g,h := DMux8Way(load, addrHigh)
            for i, x := range [8]bit{a,b,c,d,e,f,g,h} {
                ram64.ram8s[i].Load <- x
            }
        }
    }(addrs[0])
    // combinational out
    out8chan := make(chan [8][16]bit, 1)
    gather16x8([8]chan [16]bit{
        ram64.ram8s[0].Out,
        ram64.ram8s[1].Out,
        ram64.ram8s[2].Out,
        ram64.ram8s[3].Out,
        ram64.ram8s[4].Out,
        ram64.ram8s[5].Out,
        ram64.ram8s[6].Out,
        ram64.ram8s[7].Out,
    }, out8chan)
    var address [3]bit
    var o [8][16]bit
    for {
        select {
        case b := <-addrs[1]:
            address = [3]bit{b[0], b[1], b[2]}
            // we dont output here since an out will come from the ram8s
        case b := <-out8chan:
            o = b
            ram64.Out<-Mux8Way16(o[0], o[1], o[2], o[3], o[4], o[5], o[6], o[7], address)
        }
    }
}

type RAM512 struct {
    In chan [16]bit
    Out chan [16]bit
    Address chan [9]bit
    Load chan bit
    ram64s []RAM64
}

func NewRAM512(in, out chan [16]bit, addr chan [9]bit, load chan bit) RAM512 {
    ins := fanout16(in, 8)
    ram64s := make([]RAM64, 8)
    for i:=0; i<8; i++ {
        rout := make(chan [16]bit, 1)
        raddr := make(chan [6]bit, 1)
        rload := make(chan bit, 1)
        ram64s[i] = NewRAM64(ins[i], rout, raddr, rload)
    }
    return RAM512{
        In: in,
        Out: out,
        Address: addr,
        Load: load,
        ram64s: ram64s,
    }
}

// another great case for generics/generation btw
func (ram512 RAM512) Run(clock chan bit) {
    go Run(clock, ram512.ram64s[0], ram512.ram64s[1], ram512.ram64s[2], ram512.ram64s[3],
        ram512.ram64s[4], ram512.ram64s[5], ram512.ram64s[6], ram512.ram64s[7])
    addrs := fanout9(ram512.Address, 2)
    // combinational in
    go func(addr chan [9]bit) {
        var addrHigh [3]bit
        var addrLow [6]bit
        var load bit
        for {
            select {
            case b := <-addr:
                addrHigh = [3]bit{b[0], b[1], b[2]}
                addrLow = [6]bit{b[3], b[4], b[5], b[6], b[7], b[8]}
                for _, r := range ram512.ram64s {
                    r.Address <- addrLow
                }
            case b := <-ram512.Load:
                load = b
            }
            a,b,c,d,e,f,g,h := DMux8Way(load, addrHigh)
            for i, x := range [8]bit{a,b,c,d,e,f,g,h} {
                ram512.ram64s[i].Load <- x
            }
        }
    }(addrs[0])
    // combinational out
    out8chan := make(chan [8][16]bit, 1)
    gather16x8([8]chan [16]bit{
        ram512.ram64s[0].Out,
        ram512.ram64s[1].Out,
        ram512.ram64s[2].Out,
        ram512.ram64s[3].Out,
        ram512.ram64s[4].Out,
        ram512.ram64s[5].Out,
        ram512.ram64s[6].Out,
        ram512.ram64s[7].Out,
    }, out8chan)
    var address [3]bit
    var o [8][16]bit
    for {
        select {
        case b := <-addrs[1]:
            address = [3]bit{b[0], b[1], b[2]}
            // we dont output here since an out will come from the ram8s
        case b := <-out8chan:
            o = b
            ram512.Out<-Mux8Way16(o[0], o[1], o[2], o[3], o[4], o[5], o[6], o[7], address)
        }
    }
}

type RAM4K struct {
    In chan [16]bit
    Out chan [16]bit
    Address chan [12]bit
    Load chan bit
    rams []RAM512
}

func NewRAM4K(in, out chan [16]bit, addr chan [12]bit, load chan bit) RAM4K {
    ins := fanout16(in, 8)
    rams := make([]RAM512, 8)
    for i:=0; i<8; i++ {
        rout := make(chan [16]bit, 1)
        raddr := make(chan [9]bit, 1)
        rload := make(chan bit, 1)
        rams[i] = NewRAM512(ins[i], rout, raddr, rload)
    }
    return RAM4K{
        In: in,
        Out: out,
        Address: addr,
        Load: load,
        rams: rams,
    }
}

func (r RAM4K) Run(clock chan bit) {
    go Run(clock, r.rams[0], r.rams[1], r.rams[2], r.rams[3],
        r.rams[4], r.rams[5], r.rams[6], r.rams[7])
    addrs := fanout12(r.Address, 2)
    // combinational in
    go func(addr chan [12]bit) {
        var addrHigh [3]bit
        var addrLow [9]bit
        var load bit
        for {
            select {
            case b := <-addr:
                addrHigh = [3]bit{b[0], b[1], b[2]}
                addrLow = [9]bit{b[3], b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11]}
                for _, ram := range r.rams {
                    ram.Address <- addrLow
                }
            case b := <-r.Load:
                load = b
            }
            a,b,c,d,e,f,g,h := DMux8Way(load, addrHigh)
            for i, x := range [8]bit{a,b,c,d,e,f,g,h} {
                r.rams[i].Load <- x
            }
        }
    }(addrs[0])
    // combinational out
    out8chan := make(chan [8][16]bit, 1)
    gather16x8([8]chan [16]bit{
        r.rams[0].Out,
        r.rams[1].Out,
        r.rams[2].Out,
        r.rams[3].Out,
        r.rams[4].Out,
        r.rams[5].Out,
        r.rams[6].Out,
        r.rams[7].Out,
    }, out8chan)
    var address [3]bit
    var o [8][16]bit
    for {
        select {
        case b := <-addrs[1]:
            address = [3]bit{b[0], b[1], b[2]}
            // we dont output here since an out will come from the ram8s
        case b := <-out8chan:
            o = b
            r.Out<-Mux8Way16(o[0], o[1], o[2], o[3], o[4], o[5], o[6], o[7], address)
        }
    }
}
// TODO: build up to ram16k
