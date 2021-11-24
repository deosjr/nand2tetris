package main

import (
    "bufio"
    "fmt"
    "os"
    "time"

    "github.com/faiface/pixel"
    "github.com/faiface/pixel/pixelgl"
)

var headless = false

// first instr jumps to main program, drawchar comes first though (easier)
//var program = append(append([]uint16{0x329, 0xEA87}, drawChar...), helloworld...)
//var program = append(append([]uint16{0x329, 0xEA87}, drawChar...), keyboardLoop...)
//var program = append(append([]uint16{0x329, 0xEA87}, drawChar...), writeHex...)
var program = assembleFirstPass

// maybe take an output func that prints to terminal?
func run(computer *Computer) {
    computer.SendReset(true)
    computer.ClockTick()
    computer.SendReset(false)
    fmt.Println("booting...")

    // set test data in ram: assemble the assembler using itself!
    ram := computer.data_mem.ram
    datapointer := 0x1000
    f, _ := os.Open("firstpass.asm")
	defer f.Close()
	scanner := bufio.NewScanner(f)
    scanner.Split(bufio.ScanRunes)

	for scanner.Scan() {
        char := scanner.Text()[0]
        fmt.Print(scanner.Text())
        if char == ' ' {
            continue
        }
        if char == '\n' {
            char = 0x80
        }
        ram.mem[datapointer] = uint16(char)
        datapointer++
        if datapointer == 0x4000 {
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
        //fmt.Printf(" %04x %04x", computer.data_mem.ram.mem[0x4], computer.data_mem.ram.mem[0x5])
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
    outputpointer := 0x1000
    endoutput := int(computer.data_mem.ram.mem[0x2])
    for {
        if outputpointer == endoutput {
            break
        }
        v := computer.data_mem.ram.mem[outputpointer]
        if v == 0 {
            break
        }
        output = append(output, v)
        outputpointer++
    }
    for i, v := range output {
        fmt.Printf("%d: %04x\n", i, v)
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
