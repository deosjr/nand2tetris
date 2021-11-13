package main

import (
    "testing"
    "time"
    "sync"
)

func TestRAM8(t *testing.T) {
    clock := make(chan bit)
    in := make(chan [16]bit)
    load := make(chan bit)
    sel := make(chan [3]bit)
    out := make(chan [16]bit)
    test := NewRAM8(in, out, sel, load)
    go test.Run(clock)

    clock<-true
    v := fromBit16(<-test.Out)
    if v != 0 {
        t.Errorf("expected 0")
    }
    in<-toBit16(42)
    time.Sleep(10*time.Millisecond)
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 0 {
        t.Errorf("expected 0")
    }
    load<-true
    time.Sleep(10*time.Millisecond)
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 42 {
        t.Errorf("expected 42")
    }
    in<-toBit16(4058)
    time.Sleep(10*time.Millisecond)
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 4058 {
        t.Errorf("expected 4058")
    }
    sel<-[3]bit{false, false, true}
    time.Sleep(10*time.Millisecond)
    //clock<-true // NOTE: changing addr causes output to change!
    // this is a combinational operation, independent of the clock
    v = fromBit16(<-test.Out)
    if v != 0 {
        t.Errorf("expected 0")
    }
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 4058 {
        t.Errorf("expected 4058, got %d", v)
    }
    in<-toBit16(15)
    time.Sleep(10*time.Millisecond)
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 15 {
        t.Errorf("expected 15")
    }
    sel<-[3]bit{false, false, false}
    time.Sleep(10*time.Millisecond)
    v = fromBit16(<-test.Out)
    if v != 4058 {
        t.Errorf("expected 4058")
    }
}

func TestRAM8WithClock(t *testing.T) {
    ticker := time.NewTicker(time.Nanosecond)
    defer ticker.Stop()
    clock := make(chan bit)
    go func() {
        for {
            select {
            case <-ticker.C:
                clock<-true
            }
        }
    }()
    // setup our inputs as repeaters
    loada := make(chan bit)
    ina := make(chan [16]bit)
    outa := make(chan [16]bit)
    bit16input := NewRegister(ina, outa, loada)
    loada<-true

    addr := make(chan [3]bit, 1)
    addr <- [3]bit{false, true, false}

    inb := make(chan bit)
    outb := make(chan bit)
    loadinput := NewDFF(inb, outb)

    out := make(chan [16]bit)
    test := NewRAM8(outa, out, addr, outb)
    go Run(clock, bit16input, loadinput, test)

    var lock sync.Mutex
    var output [16]bit
    go func() {
        for {
            select {
            case b := <-test.Out:
                lock.Lock()
                output = b
                lock.Unlock()
            }
        }
    }()

    lock.Lock()
    if fromBit16(output) != 0 {
        t.Errorf("expected 0")
    }
    lock.Unlock()
    bit16input.In<-toBit16(42)
    loadinput.In<-true
    // wait for things to stabilize
    time.Sleep(10*time.Millisecond)
    lock.Lock()
    if fromBit16(output) != 42 {
        t.Errorf("expected 42")
    }
    lock.Unlock()
}
func TestRAM64(t *testing.T) {
    clock := make(chan bit)
    in := make(chan [16]bit)
    load := make(chan bit)
    sel := make(chan [6]bit)
    out := make(chan [16]bit)
    test := NewRAM64(in, out, sel, load)
    go test.Run(clock)

    clock<-true
    v := fromBit16(<-test.Out)
    if v != 0 {
        t.Errorf("expected 0")
    }
    in<-toBit16(42)
    time.Sleep(10*time.Millisecond)
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 0 {
        t.Errorf("expected 0")
    }
    load<-true
    time.Sleep(10*time.Millisecond)
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 42 {
        t.Errorf("expected 42")
    }
    in<-toBit16(4058)
    time.Sleep(10*time.Millisecond)
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 4058 {
        t.Errorf("expected 4058")
    }
    sel<-[6]bit{false, false, true, false, true, true}
    time.Sleep(10*time.Millisecond)
    //clock<-true // NOTE: changing addr causes output to change!
    // this is a combinational operation, independent of the clock
    // then expect 0 from the new mem address (we havent synced the clock yet)
    v = fromBit16(<-test.Out)
    if v != 0 {
        t.Errorf("expected 0")
    }
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 4058 {
        t.Errorf("expected 4058, got %d", v)
    }
    in<-toBit16(15)
    time.Sleep(10*time.Millisecond)
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 15 {
        t.Errorf("expected 15")
    }
    sel<-[6]bit{false, false, false, false, false, false}
    time.Sleep(10*time.Millisecond)
    v = fromBit16(<-test.Out)
    if v != 4058 {
        t.Errorf("expected 4058, got %d", v)
    }
    clock<-true
    v = fromBit16(<-test.Out)
    if v != 15 {
        t.Errorf("expected 15, got %d", v)
    }
}

func TestRAM64WithClock(t *testing.T) {
    ticker := time.NewTicker(time.Nanosecond)
    defer ticker.Stop()
    clock := make(chan bit)
    go func() {
        for {
            select {
            case <-ticker.C:
                clock<-true
            }
        }
    }()
    // setup our inputs as repeaters
    loada := make(chan bit)
    ina := make(chan [16]bit)
    outa := make(chan [16]bit)
    bit16input := NewRegister(ina, outa, loada)
    loada<-true

    addr := make(chan [6]bit, 1)
    addr <- [6]bit{false, true, false, true, false, true}

    inb := make(chan bit)
    outb := make(chan bit)
    loadinput := NewDFF(inb, outb)

    out := make(chan [16]bit)
    test := NewRAM64(outa, out, addr, outb)
    go Run(clock, bit16input, loadinput, test)

    var output [16]bit
    go func() {
        for {
            select {
            case b := <-test.Out:
                output = b
            }
        }
    }()

    if fromBit16(output) != 0 {
        t.Errorf("expected 0")
    }
    bit16input.In<-toBit16(42)
    loadinput.In<-true
    // wait for things to stabilize
    for fromBit16(output) == 0 {
        time.Sleep(10*time.Millisecond)
    }
    if fromBit16(output) != 42 {
        t.Errorf("expected 42")
    }
}
