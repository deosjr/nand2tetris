package main

import (
    "testing"
    "time"
    "sync"
)

func TestDFF(t *testing.T) {
    clock := make(chan bit)
    in := make(chan bit)
    out := make(chan bit)
    dff := NewDFF(in, out)
    go dff.Run(clock)
    clock<-true
    if <-dff.Out == true {
        t.Errorf("expected false")
    }
    dff.In<-true
    dff.In<-false
    dff.In<-true
    select {
    case <-dff.Out:
        t.Errorf("expected nothing on out")
    default:
    }
    clock<-true
    if <-dff.Out == false {
        t.Errorf("expected true")
    }
}

func TestBit(t *testing.T) {
    clock := make(chan bit)
    tin := make(chan bit)
    tload := make(chan bit)
    out := make(chan bit)
    test := NewBit(tin, tload, out)
    go test.Run(clock)

    clock<-true
    if <-test.Out == true {
        t.Errorf("expected false")
    }
    tin<-true
    time.Sleep(10*time.Millisecond)
    clock<-true
    if <-test.Out == true {
        t.Errorf("expected false")
    }
    tload<-true
    time.Sleep(10*time.Millisecond)
    clock<-true
    if <-test.Out == false {
        t.Errorf("expected true")
    }
    tin<-false
    time.Sleep(10*time.Millisecond)
    clock<-true
    if <-test.Out == true {
        t.Errorf("expected false")
    }
}

func TestBitWithClock(t *testing.T) {
    // should be okay with a ridiculously fast (for our purposes) clock
    // since the world should just halt if we cannot keep up
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
    // setup our inputs as repeaters using dffs
    ina := make(chan bit)
    outa := make(chan bit)
    a := NewDFF(ina, outa)

    inb := make(chan bit)
    outb := make(chan bit)
    b := NewDFF(inb, outb)

    out := make(chan bit)
    test := NewBit(outa, outb, out)
    go Run(clock, a, b, test)

    var lock sync.Mutex
    var output bit
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
    if output == true {
        t.Errorf("expected false")
    }
    lock.Unlock()
    a.In<-true
    b.In<-true
    // wait for things to stabilize
    time.Sleep(10*time.Millisecond)
    lock.Lock()
    if output == false {
        t.Errorf("expected true")
    }
    lock.Unlock()
}

func TestRegister(t *testing.T) {
    clock := make(chan bit)
    in := make(chan [16]bit)
    load := make(chan bit)
    out := make(chan [16]bit)
    test := NewRegister(in, out, load)
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
}

func TestRegisterWithClock(t *testing.T) {
    // should be okay with a ridiculously fast (for our purposes) clock
    // since the world should just halt if we cannot keep up
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
    test := NewRegister(outa, out, outb)
    go Run(clock, test, bit16input, loadinput)

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
    for fromBit16(output) == 0 {
        time.Sleep(10*time.Millisecond)
    }
    lock.Lock()
    if fromBit16(output) != 42 {
        t.Errorf("expected 42")
    }
    lock.Unlock()
}
