package main

import (
    "fmt"
    "time"

    "github.com/faiface/pixel"
    "github.com/faiface/pixel/pixelgl"
)

var headless = false

var program = append(writeHex, drawChar...)
//var program = append(helloworld, drawChar...)
//var program = append(keyboardLoop, drawChar...)

// maybe take an output func that prints to terminal?
func run(computer *Computer) {
    computer.SendReset(true)
    computer.ClockTick()
    computer.SendReset(false)
    fmt.Println("booting...")

    // set test data in ram
    ram := computer.data_mem.ram
    ram.mem[0x1000] = 0x7345
    /*
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
    */

    //fmt.Println("pc: inst| in | ax | dx | out")
    for {
        //cpu := computer.cpu.(*BarrelShiftCPU)
        //fmt.Printf("%02d: %04x %04x", cpu.PC(), computer.instr_mem.Out(), cpu.inM)
        computer.ClockTick()
        /*
        if cpu.PC() >= 0 && cpu.PC() < 30 {
            fmt.Printf("%02d: %04x %04x %04x ", cpu.PC(), computer.instr_mem.Out(), cpu.a.Out(), cpu.d.Out())
            fmt.Printf("%04x\n", computer.data_mem.ram.mem[0x2])
        }
        */
        //fmt.Printf(" %04x %04x %04x", cpu.a.Out(), cpu.d.Out(), cpu.OutM())
        //fmt.Printf(" %04x %04x", computer.data_mem.ram.mem[0x4], computer.data_mem.ram.mem[0x5])
        //fmt.Printf(" %04x %04x\n", computer.data_mem.ram.mem[0x2], computer.data_mem.ram.mem[0x3])
        time.Sleep(1*time.Nanosecond)
        //time.Sleep(100*time.Millisecond)
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
    computer.LoadProgram(NewROM32K(program))

    if headless {
        run(computer)
    } else {
        go run(computer)
        pixelgl.Run(runPeripherals(computer))
    }
}
