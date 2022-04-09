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

//var program = tapeAssembler
var program = []uint16{
    0x0100,
    0xec10,
    0x0000,
    0xe308,
    0x0006,
    0xea87,
    0x0016,
    0xec10,
    0x000d,
    0xe308,
    0x0000,
    0xec10,
    0x000e,
    0xe308,
    0x0014,
    0xec10,
    0x000f,
    0xe308,
    0x00b5,
    0xea87,
    0x0014,
    0xea87,
    0x6002,
    0xea88,
    0x0002,
    0xec10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0003,
    0xec10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x003d,
    0xec10,
    0x000d,
    0xe308,
    0x0002,
    0xec10,
    0x000e,
    0xe308,
    0x0032,
    0xec10,
    0x000f,
    0xe308,
    0x00b5,
    0xea87,
    0x0000,
    0xfca8,
    0xfc10,
    0x1772,
    0xe308,
    0x0000,
    0xfdc8,
    0xfca0,
    0xea88,
    0x00ef,
    0xea87,
    0x0002,
    0xec10,
    0x000e,
    0xe308,
    0x0047,
    0xec10,
    0x000f,
    0xe308,
    0x00e2,
    0xea87,
    0x6002,
    0xefc8,
    0x0000,
    0xfdc8,
    0xfca0,
    0xea88,
    0x0000,
    0xfca8,
    0xfc10,
    0x0001,
    0xfc20,
    0xe308,
    0x0002,
    0xfde0,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0000,
    0xfca8,
    0xfc10,
    0x0001,
    0xfde0,
    0xe308,
    0x0000,
    0xfdc8,
    0xfca0,
    0xea88,
    0x0001,
    0xfde0,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0000,
    0xfca8,
    0xfc10,
    0xeca0,
    0xf4d0,
    0xea88,
    0x0076,
    0xe305,
    0x0000,
    0xfca0,
    0xfc48,
    0x0000,
    0xfca8,
    0xfc10,
    0x00ac,
    0xe342,
    0x0001,
    0xfc20,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0002,
    0xfc20,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0000,
    0xfca8,
    0xfc10,
    0xeca0,
    0xf088,
    0x0000,
    0xfca8,
    0xfc10,
    0x0001,
    0xfc20,
    0xe308,
    0x0001,
    0xfde0,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0000,
    0xfdc8,
    0xfca0,
    0xefc8,
    0x0000,
    0xfca8,
    0xfc10,
    0xeca0,
    0xf4c8,
    0x0000,
    0xfca8,
    0xfc10,
    0x0001,
    0xfde0,
    0xe308,
    0x0060,
    0xea87,
    0x0001,
    0xfc20,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x00ef,
    0xea87,
    0x000f,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0001,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0002,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0003,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0004,
    0xfc10,
    0x0000,
    0xfdc8,
    0xfca0,
    0xe308,
    0x0000,
    0xfc10,
    0x000e,
    0xf4d0,
    0x0005,
    0xe4d0,
    0x0002,
    0xe308,
    0x0000,
    0xfc10,
    0x0001,
    0xe308,
    0x000d,
    0xfc20,
    0xea87,
    0x000e,
    0xfc10,
    0x000f,
    0xfc20,
    0xe302,
    0x0000,
    0xfdc8,
    0xfca0,
    0xea88,
    0x000e,
    0xfc88,
    0x00e2,
    0xea87,
    0x0001,
    0xfc10,
    0x000e,
    0xe318,
    0x0005,
    0xe4e0,
    0xfc10,
    0x000f,
    0xe308,
    0x0000,
    0xfca8,
    0xfc10,
    0x0002,
    0xfc20,
    0xe308,
    0x0002,
    0xfdd0,
    0x0000,
    0xe308,
    0x000e,
    0xfca8,
    0xfc10,
    0x0004,
    0xe308,
    0x000e,
    0xfca8,
    0xfc10,
    0x0003,
    0xe308,
    0x000e,
    0xfca8,
    0xfc10,
    0x0002,
    0xe308,
    0x000e,
    0xfca8,
    0xfc10,
    0x0001,
    0xe308,
    0x000f,
    0xfc20,
    0xea87,
}

func main() {
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    fmt.Println("loading ROM")
    computer.LoadProgram(NewROM32K(program))
    computer.data_mem.reader.LoadInputTapes([]string{
        // we feed the input in twice since we run two passes over it
        "asm/vm_mult.asm",
        "asm/vm_mult.asm",
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
        cpu := computer.cpu.(*BarrelShiftCPU)
        fmt.Printf("%03d: %04x", cpu.PC(), computer.instr_mem.Out())
        computer.ClockTick()
        /*
        fmt.Printf(" %04x %04x %04x", cpu.a.Out(), cpu.d.Out(), cpu.OutM())
        fmt.Printf(" SP:%04x LCL:%04x ARG:%04x", computer.data_mem.ram.mem[0x0], computer.data_mem.ram.mem[0x1], computer.data_mem.ram.mem[0x2])
        fmt.Printf(" %04x %04x %04x\n", computer.data_mem.ram.mem[0xD], computer.data_mem.ram.mem[0xE], computer.data_mem.ram.mem[0xF])
        fmt.Print(" STACK:")
        for i:=0x100; i<0x118;i++ {
            fmt.Printf(" %04x", computer.data_mem.ram.mem[i])
        }
        */
        fmt.Println()
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

