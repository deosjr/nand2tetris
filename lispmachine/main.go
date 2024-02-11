package main

import (
    "fmt"
    //"time"
)

var debug = false

func main() {
    //out, err := compile("lisp/list.scm", "lisp/main.scm")
    out, err := compile("lisp/main.scm")
    if err != nil {
        fmt.Println(err)
        return
    }

    //asm, err := Translate([]string{"vm/eval.vm"}, out)
    asm, err := Translate([]string{}, out)
    if err != nil {
        fmt.Println(err)
        return
    }

    program, err := assembleFromString(asm)
    if err != nil {
        fmt.Println(err)
        return
    }

    cpu := NewLispCPU()
    computer := NewLispMachine(cpu)
    fmt.Println("loading ROM")
    computer.LoadProgram(NewROM32K(program))
    computer.data_mem.writer.LoadOutputWriter(charPrinter{})

    var debugger Debugger
    if debug {
        debugger = &standardDebugger{}
        //debugger = &analysisDebugger{}
    }
    run(computer, debugger)

/*
    for i:=0x10; i < 0x7ff; i++ {
        v := computer.data_mem.ramCar.mem[i]
        fmt.Printf("%04x %04x\n", i, v)
    }
*/

/*
    for i:=0x800; i < 0x840; i++ {
        car := computer.data_mem.ramCar.mem[i]
        cdr := computer.data_mem.ramCdr.mem[i]
        fmt.Printf("%04x %04x %04x\n", i, car, cdr)
    }
    if debug {
        rom := debugger.(*analysisDebugger).rom

        type romRange struct {
            from, to, n int
        }
        var ranges []romRange
        current := romRange{n:rom[0]}
        for i, v := range rom {
            if v != current.n {
                current.to = i-1
                ranges = append(ranges, current)
                current = romRange{from:i, n:v}
            }
        }
        current.to = len(rom)-1
        ranges = append(ranges, current)
        for _, r := range ranges {
            fmt.Println(r.from, r.to, r.n)
        }
    }
*/
}

func run(computer *LispMachine, debugger Debugger) {
    computer.SendReset(true)
    computer.ClockTick()
    computer.SendReset(false)
    fmt.Println("booting...")

    ticks := 0

    if debugger != nil {
        debugger.BeforeLoop()
    }
    var pprev, prev uint16
    for {
        if debugger != nil {
            debugger.BeforeTick(computer)
        }
        computer.ClockTick()
        ticks++
        if debugger != nil {
            debugger.AfterTick(computer)
        }
        // NOTE: without this sleep, output printing can lag behind program ending!
        //time.Sleep(10*time.Nanosecond)
        //time.Sleep(10000*time.Nanosecond)
        //time.Sleep(10*time.Millisecond)
        // NOTE: this halts running the computer after finding a tight loop
        if pprev == computer.cpu.PC() {
            break
        }
        pprev = prev
        prev = computer.cpu.PC()
    }
    fmt.Println("ticks:", ticks)
}

type Debugger interface {
    BeforeLoop()
    BeforeTick(*LispMachine)
    AfterTick(*LispMachine)
}

type standardDebugger struct {
    i int
    sp uint16
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
    sp := c.data_mem.ramCar.mem[1]
    if sp == sd.sp { return }
    sd.sp = sp
    v := c.data_mem.ramCar.mem[sd.sp-1]
    if v == 0x8825 {
        fmt.Println(c.cpu.pc.Out())
    }
    //fmt.Printf(" %04x %04x %04x\n", c.cpu.a.Out(), c.cpu.d.Out(), c.cpu.OutCarM())
    //fmt.Printf("%03d: %04x %04x %04x %04x %04x\n", c.cpu.pc.Out(), c.cpu.instr, c.cpu.a.Out(), c.cpu.d.Out(), c.cpu.inCarM, c.cpu.inCdrM)
    //fmt.Printf(" %04x %04x %04x %04x %04x", c.data_mem.ramCar.mem[13], c.data_mem.ramCar.mem[14], c.data_mem.ramCar.mem[15], c.data_mem.ramCar.mem[1], c.data_mem.ramCar.mem[3])
    //fmt.Printf(" [%04x %04x %04x %04x %04x", c.data_mem.ramCar.mem[256], c.data_mem.ramCar.mem[257], c.data_mem.ramCar.mem[258], c.data_mem.ramCar.mem[259], c.data_mem.ramCar.mem[260])
    //fmt.Printf(" %04x %04x %04x %04x %04x]", c.data_mem.ramCar.mem[261], c.data_mem.ramCar.mem[262], c.data_mem.ramCar.mem[263], c.data_mem.ramCar.mem[264], c.data_mem.ramCar.mem[265])
    //fmt.Printf(" [SP:%04x ENV:%04x ARG:%04x FREE:%04x", c.data_mem.ramCar.mem[1], c.data_mem.ramCar.mem[2], c.data_mem.ramCar.mem[3], c.data_mem.ramCar.mem[4])
    //fmt.Println()
    //fmt.Printf(" %04x %04x\n", c.data_mem.ramCar.mem[0], c.data_mem.ramCdr.mem[0])
    // TODO: wait for keyboard press to make step-through debugger
    //fmt.Println()
}

type analysisDebugger struct {
    rom [32768]int
}

func (*analysisDebugger) BeforeLoop() {}
func (ad *analysisDebugger) BeforeTick(c *LispMachine) {
    pc := c.cpu.PC()
    ad.rom[pc] = ad.rom[pc] + 1
}
func (*analysisDebugger) AfterTick(c *LispMachine) {}

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
