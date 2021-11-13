package main

import (
    "fmt"
    "time"

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
        //cpu := computer.cpu.(*PCRegisterCPU)
        //fmt.Printf("%02d: %04x %04x", cpu.PC(), computer.instr_mem.Out(), cpu.inM)
        computer.ClockTick()
        //fmt.Printf(" %04x %04x %04x", cpu.a.Out(), cpu.d.Out(), cpu.OutM())
        //fmt.Printf(" %04x %04x\n", computer.data_mem.ram.mem[0x2], computer.data_mem.ram.mem[0x3])
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
    cpu := NewPCRegisterCPU()
    computer := NewComputer(cpu)
    computer.LoadProgram(NewROM32K(drawChar))

    computer.data_mem.ram.mem[2] = 0x41

    if interactive {
        go run(computer)
        pixelgl.Run(runPeripherals(computer))
    } else {
        run(computer)
    }
}
