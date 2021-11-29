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

// assembled with asm/assembler.asm from fig 6.2 in book (prog.asm)
// should add 1+...+100 and store in @sum (so 0x11), which should be 5050 or 0x13ba
var program = []uint16{
    0x0010,
    0xefc8,
    0x0011,
    0xea88,
    0x0010,
    0xfc10,
    0x0064,
    0xe4d0,
    0x0012,
    0xe301,
    0x0010,
    0xfc10,
    0x0011,
    0xf088,
    0x0010,
    0xfdc8,
    0x0004,
    0xea87,
    0x0012,
    0xea87,
}

// first instr jumps to main program, drawchar comes first though (easier)
//var program = append(append([]uint16{0x329, 0xEA87}, drawChar...), helloworld...)
//var program = append(append([]uint16{0x329, 0xEA87}, drawChar...), keyboardLoop...)
//var program = append(append([]uint16{0x329, 0xEA87}, drawChar...), writeHex...)
//var program = assembleFirstPass
//var program = assembleTwoPass
//var program = assembleTwoPassPlus
//var program = countLines

// maybe take an output func that prints to terminal?
func run(computer *Computer) {
    computer.SendReset(true)
    computer.ClockTick()
    computer.SendReset(false)
    fmt.Println("booting...")

    // set test data in ram: assemble the assembler using itself!
    ram := computer.data_mem.ram
    datapointer := 0x1000
    //f, _ := os.Open("asm/assembler.asm")
    f, _ := os.Open("test2")
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
    /*
    // HELLO WORLD!
    ram.mem[0x1000] = 0x48
    ram.mem[0x1001] = 0x45
    ram.mem[0x1002] = 0x4C
    ram.mem[0x1003] = 0x4C
    ram.mem[0x1004] = 0x4F
    ram.mem[0x1005] = 0x20
    ram.mem[0x1006] = 0x57
    ram.mem[0x1007] = 0x4F
    ram.mem[0x1008] = 0x52
    ram.mem[0x1009] = 0x4C
    ram.mem[0x100A] = 0x44
    ram.mem[0x100B] = 0x21
    */

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
        if pprev == computer.cpu.PC() {
            break
        }
        pprev = prev
        prev = computer.cpu.PC()
    }
    output := []uint16{}
    //outputpointer := 0x1000
    //endoutput := int(computer.data_mem.ram.mem[0x2])
    //outputpointer := 0x20
    //endoutput := int(computer.data_mem.ram.mem[0x8])
    outputpointer := 0x11
    endoutput := 0x12
    for {
        if outputpointer == endoutput {
            break
        }
        v := computer.data_mem.ram.mem[outputpointer]
        output = append(output, v)
        outputpointer++
    }
    for _, v := range output {
        fmt.Printf("%04x\n", v)
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
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    fmt.Println("loading ROM")
    computer.LoadProgram(NewROM32K(program))

    if headless {
        run(computer)
    } else {
        go run(computer)
        pixelgl.Run(runPeripherals(computer))
    }
}
