package main

import (
    "bufio"
    "fmt"
    "os"
    "time"

    "github.com/faiface/pixel"
    "github.com/faiface/pixel/pixelgl"
)

var headless = true

var program = tapeAssembler

func main() {
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    fmt.Println("loading ROM")
    computer.LoadProgram(NewROM32K(program))
    computer.data_mem.reader.LoadInputTapes([]string{
        // we feed the input in twice since we run two passes over it
        "asm/assembler.asm",
        "asm/assembler.asm",
    })

    if headless {
        run(computer)
    } else {
        go run(computer)
        pixelgl.Run(runPeripherals(computer))
    }
}

func run(computer *Computer) {
    computer.SendReset(true)
    computer.ClockTick()
    computer.SendReset(false)
    fmt.Println("booting...")

    //fmt.Println("pc: inst| in | ax | dx | out")
    var pprev, prev uint16
    for {
        //cpu := computer.cpu.(*BarrelShiftCPU)
        //fmt.Printf("%02d: %04x %04x", cpu.PC(), computer.instr_mem.Out(), cpu.inM)
        computer.ClockTick()
        /*
        if cpu.PC() >= 0 && cpu.PC() < 40 {
            fmt.Printf("%02d: %04x %04x %04x ", cpu.PC(), computer.instr_mem.Out(), cpu.a.Out(), cpu.d.Out())
            fmt.Printf("%04x %04x\n", computer.data_mem.ram.mem[0x2], computer.data_mem.ram.mem[0x7])
        }
        */
        //fmt.Printf(" %04x %04x %04x", cpu.a.Out(), cpu.d.Out(), cpu.OutM())
        //fmt.Printf(" %04x %04x", computer.data_mem.ram.mem[0x1], computer.data_mem.ram.mem[0x2])
        //fmt.Printf(" %04x %04x\n", computer.data_mem.ram.mem[0x7], computer.data_mem.ram.mem[0x8])
        //fmt.Printf(" %04x %04x\n", computer.data_mem.ram.mem[0x30], computer.data_mem.ram.mem[0x31])
        //fmt.Printf(" %04x %04x\n", computer.data_mem.reader.Out(), computer.data_mem.writer.Out())
        //fmt.Printf(" %04x %04x\n", computer.data_mem.ram.mem[0x10], computer.data_mem.ram.mem[0x12])
        /*
        fmt.Printf(" LABEL: %04x %04x", computer.data_mem.ram.mem[0x20], computer.data_mem.ram.mem[0x21])
        fmt.Printf(" %04x %04x", computer.data_mem.ram.mem[0x22], computer.data_mem.ram.mem[0x23])
        fmt.Printf(" %04x %04x", computer.data_mem.ram.mem[0x24], computer.data_mem.ram.mem[0x25])
        fmt.Printf(" %04x %04x", computer.data_mem.ram.mem[0x26], computer.data_mem.ram.mem[0x27])
        fmt.Printf(" %04x %04x\n", computer.data_mem.ram.mem[0x28], computer.data_mem.ram.mem[0x29])
        */
        //fmt.Printf(" %04x %04x %04x\n", computer.data_mem.ram.mem[0x0], computer.data_mem.ram.mem[0x10], computer.data_mem.ram.mem[0x11])
        time.Sleep(1*time.Nanosecond)
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

