package main

import (
    "bufio"
    "fmt"
    //"io"
    "os"
    //"strings"
    "time"

    "github.com/faiface/pixel"
    "github.com/faiface/pixel/pixelgl"
)

var headless = true
var debug = false

func main() {
    assembler, err := Translate([]string{"vm/mult.vm", "vm/fact.vm"})
    if err != nil {
        fmt.Println(err)
        return
    }
    program, err := assembleFromString(assembler)
    //program, err := Assemble("asm/decimal.asm")
    if err != nil {
        fmt.Println(err)
        return
    }
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    fmt.Println("loading ROM")
    computer.LoadProgram(NewROM32K(program))
    /*
    computer.data_mem.reader.LoadInputTapes([]string{
        // we feed the input in twice since we run two passes over it
        //"asm/vm_mult.asm",
        //"asm/vm_mult.asm",
    })
    */
    //computer.data_mem.reader.LoadInputReaders([]io.Reader{strings.NewReader("128\n")})
    //computer.data_mem.writer.LoadOutputTape("out")

    var debugger Debugger
    if debug {
        debugger = standardDebugger{}
    }

    if headless {
        run(computer, debugger)
    } else {
        go run(computer, debugger)
        pixelgl.Run(runPeripherals(computer))
    }
}

type Debugger interface {
    BeforeLoop()
    BeforeTick(*Computer)
    AfterTick(*Computer)
}

type standardDebugger struct {}

func (standardDebugger) BeforeLoop() {
    fmt.Println("pc: inst| in | ax | dx | out")
}

func (standardDebugger) BeforeTick(c *Computer) {
    cpu := c.cpu.(*BarrelShiftCPU)
    fmt.Printf("%03d: %04x", cpu.PC(), c.instr_mem.Out())
}

func (standardDebugger) AfterTick(c *Computer) {
    /*
    cpu := c.cpu.(*BarrelShiftCPU)
    fmt.Printf(" %04x %04x %04x", cpu.a.Out(), cpu.d.Out(), cpu.OutM())
    fmt.Printf(" SP:%04x LCL:%04x ARG:%04x", c.data_mem.ram.mem[0x0], c.data_mem.ram.mem[0x1], c.data_mem.ram.mem[0x2])
    fmt.Printf(" %04x %04x %04x\n", c.data_mem.ram.mem[0xD], c.data_mem.ram.mem[0xE], c.data_mem.ram.mem[0xF])
    fmt.Print(" STACK:")
    for i:=0x100; i<0x118;i++ {
        fmt.Printf(" %04x", c.data_mem.ram.mem[i])
    }
    */
    // TODO: wait for keyboard press to make step-through debugger
    fmt.Println()
}

func run(computer *Computer, debugger Debugger) {
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
        time.Sleep(10*time.Nanosecond)
        //time.Sleep(10*time.Millisecond)
        // NOTE: this halts running the computer after finding a tight loop
        if pprev == computer.cpu.PC() {
            break
        }
        pprev = prev
        prev = computer.cpu.PC()
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

// NOTE: outdated since use of tapereader
func setInput(computer *Computer, filename string) {
    // set test data in ram: assemble the assembler using itself!
    ram := computer.data_mem.ram
    datapointer := 0x1000
    f, _ := os.Open(filename)
	defer f.Close()
	scanner := bufio.NewScanner(f)
    scanner.Split(bufio.ScanRunes)

    // TODO: we have to enter the program without comments
    // or it will not fit! solve using read from disk + linker/loader?
    var comment bool
	for scanner.Scan() {
        char := scanner.Text()[0]
        if char == ' ' {
            continue
        }
        if char == '/' {
            comment = true
        }
        if char == '\n' {
            char = 0x80
        }
        if comment && char != 0x80 {
            continue
        }
        comment = false
        if char == 0x80 && (ram.mem[datapointer-1] == 0x80 || datapointer == 0x1000){
            continue
        }
        ram.mem[datapointer] = uint16(char)
        datapointer++
        if datapointer == 0x4000 {
            fmt.Println("program doesnt fit in memory!")
            break
        }
	}
}

// NOTE: outdated since use of tapewriter
func captureOutput(computer *Computer, start, end uint16) []uint16 {
    output := []uint16{}
    outputpointer := start
    for {
        if outputpointer == end {
            break }
        v := computer.data_mem.ram.mem[outputpointer]
        output = append(output, v)
        outputpointer++
    }
    return output
}

