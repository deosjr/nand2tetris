package main

import (
    "fmt"
    //"image/color"
    //"math/rand"
    "time"
    /*
    "sync"

    "net/http"
    _ "net/http/pprof"
    "runtime"
    */

    "github.com/faiface/pixel"
    "github.com/faiface/pixel/pixelgl"
)

// maybe take an output func that prints to terminal?
func run(computer *Computer) {
    computer.SendReset(true)
    computer.ClockTick()
    computer.SendReset(false)
    fmt.Println("booting...")
    //fmt.Println("pc: inst| in | ax | dx | out")
    for {
        //cpu := computer.cpu
        //fmt.Printf("%02d: %04x %04x", cpu.PC(), computer.instr_mem.Out(), cpu.inM)
        computer.ClockTick()
        //fmt.Printf(" %04x %04x %04x", cpu.a.Out(), cpu.d.Out(), cpu.OutM())
        //fmt.Printf(" %04x %04x\n", computer.data_mem.ram.mem[0x42], computer.data_mem.ram.mem[0x99])
        time.Sleep(1*time.Millisecond)
    }
}

func runPeripherals(computer *Computer) func() {
    return func() {
        cfg := pixelgl.WindowConfig{
		    Title:  "nand2tetris",
		    Bounds: pixel.R(0, 0, 256, 512),
		    VSync:  true,
	    }
	    win, err := pixelgl.NewWindow(cfg)
	    if err != nil {
		    panic(err)
	    }

	    for !win.Closed() {
            computer.data_mem.keyboard.RunKeyboard(win)
            computer.data_mem.screen.RunScreen(win)
		    win.Update()
	    }
    }
}

func main() {
    interactive := true
    computer := NewComputer()
    computer.LoadProgram(NewROM32K(drawAv2))

    /*
    // set the screen mem to some test data
    m := computer.data_mem.ram.mem
    // A
    m[16] = 0x00
    m[17] = 0x00
    m[18] = 0x0F00
    m[19] = 0x0F00
    m[20] = 0x3FC0
    m[21] = 0x3FC0
    m[22] = 0xF0F0
    m[23] = 0xF0F0
    m[24] = 0xF0F0
    m[25] = 0xF0F0
    m[26] = 0xFFF0
    m[27] = 0xFFF0
    m[28] = 0xF0F0
    m[29] = 0xF0F0
    m[30] = 0xF0F0
    m[31] = 0xF0F0
    computer.data_mem.ram.mem = m
    */

    if interactive {
        go run(computer)
        pixelgl.Run(runPeripherals(computer))
    } else {
        run(computer)
    }
    /*
    rand.Seed(time.Now().UnixNano())
    r := NewROM32K(nil)
    for i:=0; i<16384; i++ {
        r.rams[0].mem[i] = uint16(i)//uint16(rand.Intn(65536))
        r.rams[1].mem[i] = uint16(i+16384)//uint16(rand.Intn(65536))
    }
    builtins := []Builtin{r}
    //for {
    for i:=0; i<10; i++ {
        // at the start of each clock cycle, outputs send to connected inputs
        // internal builtin chips handle this themselves, but outward connected ones do not!
        // in general outputs only change at end of cycle (clocktick)
        // EXCEPTION: memory, where changing address changes output immediately (!)

        // do all kinds of special updates
        randuint := uint16(rand.Intn(32768))
        r.SendAddress(randuint)

        // at the end of each clock cycle, update
        for _, b := range builtins {
            b.ClockTick()
        }
        fmt.Println(randuint, r.Out())
    }
    */
/*
    runtime.SetBlockProfileRate(1)
    // OBSERVATIONS:
    // - allowing DFF to miss clock updates fixes the below code
    // - enabling bit16input repeater breaks it again..
    // - adding RAM16 and RAM32; 16 never breaks but 32 does more than 50%
    // - changing the clock from milli to 10*milli fixes the ram64 case completely
    // - clock ticks are dropped by the time.Ticker if we are slow
    // --------
    // simpler usecase: running one bit with a nanosecond clock.
    // ^ I think this fixes the partial deadlock but not the overall problem
    // --------
    // with the deadlock fixes to ram, the clock timer is no longer the problem (good!)
    // but now the issue is how long we need to wait for things to stabilize
    // which is a problem for our test; not sure its a problem for the simulation..

    //ticker := time.NewTicker(10*time.Millisecond)
    ticker := time.NewTicker(time.Nanosecond)
    defer ticker.Stop()
    clock := make(chan bit, 1)
    go func() {
        for {
            select {
            case <-ticker.C:
                clock<-true
                fmt.Println("clock")
            }
        }
    }()
    // setup our inputs as repeaters
    loada := make(chan bit)
    ina := make(chan [16]bit)
    outa := make(chan [16]bit)
    bit16input := NewRegister(ina, outa, loada)
    loada<-true

    addr := make(chan [12]bit, 1)
    //addr <- [6]bit{false, true, false, true, false, true}

    inb := make(chan bit)
    outb := make(chan bit)
    loadinput := NewDFF(inb, outb)

    out := make(chan [16]bit)
    test := NewRAM4K(outa, out, addr, outb)
    //test := NewRegister(outa, out, outb)
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
                fmt.Println(output)
            }
        }
    }()

    lock.Lock()
    if fromBit16(output) != 0 {
        fmt.Println("expected 0")
    }
    lock.Unlock()
    bit16input.In<-toBit16(42)
    // adding this sleep puts us in deadlock again
    time.Sleep(10*time.Millisecond)
    go func() {loadinput.In<-true}()
    // wait for things to stabilize
    timer := time.NewTimer(500*time.Millisecond)
    Loop:
    for {
        lock.Lock()
        out16 := fromBit16(output)
        lock.Unlock()
        if out16 != 0 {
            break
        }
        time.Sleep(100*time.Millisecond)
        //fmt.Println(out16)
        select {
        case <-timer.C:
            break Loop
        default:
        }
    }
    lock.Lock()
    if fromBit16(output) != 42 {
        fmt.Println("expected 42")
    } else {
        fmt.Println("success!")
        ticker.Stop()
    }
    lock.Unlock()
    http.ListenAndServe("localhost:6060", nil)
    */
}
