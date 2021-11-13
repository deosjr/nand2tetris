package main

import "sync"

// intermediate values are reevaluated upon each input change
// sequential chips consist of DFFs in between combinational chip layers

type SequentialChip interface {
    Run(clock chan bit)
}

func Run(clock chan bit, chips ...SequentialChip) {
    clocks := fanout(clock, len(chips))
    for i, chip := range chips {
        go chip.Run(clocks[i])
    }
}

// primitive, uses builtin go
type DFF struct {
    In chan bit
    Out chan bit
}

func NewDFF(in, out chan bit) DFF {
    return DFF{
        In: in,
        Out: out,
    }
}

func (c DFF) Run(clock chan bit) {
    var in bit
    var lock sync.Mutex
    go func() {
        for {
            b := <-c.In
            lock.Lock()
            in = b
            lock.Unlock()
        }
    }()
    // this is the builtin part
    // other chips are not allowed to listen to clock directly
    for {
        <-clock
        lock.Lock()
        temp := in
        lock.Unlock()
        c.Out<-temp
    }
}

// rest is not primitive, derives from DFF
type Bit struct {
    In chan bit
    Load chan bit
    Out chan bit
    dff DFF
}

func NewBit(in, load, out chan bit) Bit {
    dffin := make(chan bit, 1)
    dffout := make(chan bit, 1)
    dff := NewDFF(dffin, dffout)
    return Bit{
        In: in,
        Load: load,
        Out: out,
        dff: dff,
    }
}

func (c Bit) Run(clock chan bit) {
    go c.dff.Run(clock)
    var in, load, outloop bit
    var lock sync.Mutex
    go func() {
        for {
            b := <-c.Load
            lock.Lock()
            load = b
            tin, tload, toutloop := in, load, outloop
            lock.Unlock()
            c.dff.In<-Mux(toutloop, tin, tload)
        }
    }()
    for {
        select {
        case b := <-c.In:
            lock.Lock()
            in = b
            c.dff.In<-Mux(outloop, in, load)
            lock.Unlock()
        case b := <-c.dff.Out:
            lock.Lock()
            outloop = b
            c.dff.In<-Mux(outloop, in, load)
            lock.Unlock()
            c.Out <- b
        }
        // eval the combinational part at the end of each case
        // since each of the above 3 is an input 
    }
}

type Register struct {
    In chan [16]bit
    Load chan bit
    Out chan [16]bit
    bits []SequentialChip
}

func NewRegister(in, out chan [16]bit, load chan bit) Register {
    loads := fanout(load, 16)
    ins := chanBitArray16()
    split16(in, ins)
    outs := chanBitArray16()
    gather16(outs, out)
    bits := make([]SequentialChip, 16)
    for i:=0; i<16; i++ {
        bits[i] = NewBit(ins[i], loads[i], outs[i])
    }
    return Register{
        In: in,
        Out: out,
        Load: load,
        bits: bits,
    }
}

func (c Register) Run(clock chan bit) {
    go Run(clock, c.bits...)
}

// TODO: counter
