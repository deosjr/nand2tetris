package main

import (
    "fmt"
    //"strconv"
    "time"
)

var debug = false

func main() {
    program, err := Assemble("lispv1.asm")
    if err != nil {
        fmt.Println(err)
        return
    }

    cpu := NewLispCPU()
    computer := NewLispMachine(cpu)
    fmt.Println("loading ROM")
    computer.LoadProgram(NewROM32K(program))
    computer.data_mem.writer.LoadOutputWriter(charPrinter{})

    // setup data, see sim_test
    for i:=100; i<105; i++ {
        computer.data_mem.ramCar.mem[i] = pair(i+5)
        computer.data_mem.ramCdr.mem[i] = pair(i+1)
        computer.data_mem.ramCar.mem[i+5] = symbol(i-99)
        computer.data_mem.ramCdr.mem[i+5] = primitive(i-94)
        if i == 104 {
            computer.data_mem.ramCdr.mem[i] = emptylist()
            computer.data_mem.ramCdr.mem[i+5] = builtin(0) // PLUS
        }
    }
    computer.data_mem.ramCar.mem[111] = primitive(42) // expect prim(42) = 0x402a
    //computer.data_mem.ramCar.mem[111] = symbol(2)       // expect prim(7) = 0x4007

    var debugger Debugger
    if debug {
        debugger = &standardDebugger{}
    }
    run(computer, debugger)
}

func run(computer *LispMachine, debugger Debugger) {
    computer.SendReset(true)
    computer.ClockTick()
    computer.SendReset(false)
    fmt.Println("booting...")

    if debugger != nil {
        debugger.BeforeLoop()
    }
    var pprev, prev uint16
    for {
        if debugger != nil {
            debugger.BeforeTick(computer)
        }
        computer.ClockTick()
        if debugger != nil {
            debugger.AfterTick(computer)
        }
        // NOTE: without this sleep, output printing can lag behind program ending!
        //time.Sleep(10*time.Nanosecond)
        time.Sleep(10000*time.Nanosecond)
        //time.Sleep(10*time.Millisecond)
        // NOTE: this halts running the computer after finding a tight loop
        if pprev == computer.cpu.PC() {
            break
        }
        pprev = prev
        prev = computer.cpu.PC()
    }
}

type Debugger interface {
    BeforeLoop()
    BeforeTick(*LispMachine)
    AfterTick(*LispMachine)
}

type standardDebugger struct {
    i int
}

func (*standardDebugger) BeforeLoop() {
    //fmt.Println("pc: inst| in | ax | dx | out")
}

func (sd *standardDebugger) BeforeTick(c *LispMachine) {
    //cpu := c.cpu.(*BarrelShiftCPU)
    //fmt.Printf("%03d: %04x", cpu.PC(), c.instr_mem.Out())
    sd.i++
}

func (sd *standardDebugger) AfterTick(c *LispMachine) {
    //fmt.Printf(" %04x %04x %04x\n", c.cpu.a.Out(), c.cpu.d.Out(), c.cpu.OutCarM())
    fmt.Printf("%03d: %04x %04x %04x %04x", c.cpu.pc.Out(), c.cpu.a.Out(), c.cpu.d.Out(), c.cpu.inCarM, c.cpu.inCdrM)
    fmt.Println()
    //fmt.Printf(" %04x %04x\n", c.data_mem.ramCar.mem[0], c.data_mem.ramCdr.mem[0])
    // TODO: wait for keyboard press to make step-through debugger
    //fmt.Println()
}

type charPrinter struct{}

func (cp charPrinter) Write(p []byte) (int, error) {
    fmt.Println(string(p))
    /*
    // some big assumptions here on how tapeWriter writes
    x, err := strconv.ParseInt(string(p)[:4], 16, 16)
    if err != nil {
        return 0, err
    }
    fmt.Printf("%c", x)
    */
    return len(p), nil
}
