package main

import (
    "image/color"
    //"sync"

    "github.com/faiface/pixel"
    "github.com/faiface/pixel/pixelgl"
)

type Screen interface {
    SendIn(uint16)
    SendLoad(bool)
    SendAddress(uint16)
    Out() uint16
    ClockTick()
    RunScreen(win *pixelgl.Window)
}

// NOTE: no way to write to keyboard register
type Keyboard interface {
    Out() uint16
    ClockTick()
    RunKeyboard(win *pixelgl.Window)
}

type Screen256x512 struct {
    ram *BuiltinRAM16K // actually use only the first 8K
}

func NewScreen256x512() *Screen256x512 {
    return &Screen256x512{
        ram: NewBuiltinRAM16K(),
    }
}

func (s *Screen256x512) SendIn(in uint16) {
    s.ram.SendIn(in)
}

func (s *Screen256x512) SendLoad(load bool) {
    s.ram.SendLoad(load)
}

func (s *Screen256x512) SendAddress(addr uint16) {
    if addr >= 8192 {
        return
        //panic("access screen memory beyond 8K")
    }
    s.ram.SendAddress(addr)
}

func (s *Screen256x512) Out() uint16 {
    return s.ram.Out()
}

func (s *Screen256x512) ClockTick() {
    s.ram.ClockTick()
}

func (s *Screen256x512) RunScreen(win *pixelgl.Window) {
    // listen to mem and show in a window using pixelgl
    // NOTE: pixelgl y increases UP, nand2tetris DOWN
    white := color.RGBA{255,255,255,0}
    pd := pixel.MakePictureData(win.Bounds())
    for row:=0;row<512;row++ {
        for r:=0;r<16;r++ {
            addr := row*16+r
            word := s.ram.mem[addr]
            for c:=0;c<16;c++ {
                b := nthBit(word, uint16(c))
                if b {
                    invrow := (512-row)
                    pd.Pix[invrow*256+16*r+c] = white
                }
            }
        }
    }
    sprite := pixel.NewSprite(pd, pd.Bounds())
    sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
}

type SimpleKeyboard struct {
    reg *BuiltinRegister
}

func NewSimpleKeyboard() *SimpleKeyboard {
    reg := NewBuiltinRegister()
    reg.SendLoad(true)
    return &SimpleKeyboard{
        reg: reg,
    }
}

func (k *SimpleKeyboard) Out() uint16 {
    return k.reg.Out()
}

func (k *SimpleKeyboard) ClockTick() {
    k.reg.ClockTick()
}

func (k *SimpleKeyboard) RunKeyboard(win *pixelgl.Window) {
    switch {
    case win.Pressed(pixelgl.KeyA):
        k.reg.SendIn(0x41)
    case win.Pressed(pixelgl.KeyB):
        k.reg.SendIn(0x42)
    case win.Pressed(pixelgl.KeyC):
        k.reg.SendIn(0x43)
    case win.Pressed(pixelgl.KeyD):
        k.reg.SendIn(0x44)
    case win.Pressed(pixelgl.KeyE):
        k.reg.SendIn(0x45)
    case win.Pressed(pixelgl.KeyF):
        k.reg.SendIn(0x46)
    case win.Pressed(pixelgl.KeyG):
        k.reg.SendIn(0x47)
    case win.Pressed(pixelgl.KeyH):
        k.reg.SendIn(0x48)
    case win.Pressed(pixelgl.KeyI):
        k.reg.SendIn(0x49)
    case win.Pressed(pixelgl.KeyJ):
        k.reg.SendIn(0x4A)
    case win.Pressed(pixelgl.KeyK):
        k.reg.SendIn(0x4B)
    case win.Pressed(pixelgl.KeyL):
        k.reg.SendIn(0x4C)
    case win.Pressed(pixelgl.KeyM):
        k.reg.SendIn(0x4D)
    case win.Pressed(pixelgl.KeyN):
        k.reg.SendIn(0x4E)
    case win.Pressed(pixelgl.KeyO):
        k.reg.SendIn(0x4F)
    case win.Pressed(pixelgl.KeyP):
        k.reg.SendIn(0x50)
    case win.Pressed(pixelgl.KeyQ):
        k.reg.SendIn(0x51)
    case win.Pressed(pixelgl.KeyR):
        k.reg.SendIn(0x52)
    case win.Pressed(pixelgl.KeyS):
        k.reg.SendIn(0x52)
    case win.Pressed(pixelgl.KeyT):
        k.reg.SendIn(0x53)
    case win.Pressed(pixelgl.KeyU):
        k.reg.SendIn(0x54)
    case win.Pressed(pixelgl.KeyV):
        k.reg.SendIn(0x55)
    case win.Pressed(pixelgl.KeyW):
        k.reg.SendIn(0x56)
    case win.Pressed(pixelgl.KeyX):
        k.reg.SendIn(0x57)
    case win.Pressed(pixelgl.KeyY):
        k.reg.SendIn(0x58)
    case win.Pressed(pixelgl.KeyZ):
        k.reg.SendIn(0x59)
    case win.Pressed(pixelgl.KeySpace):
        k.reg.SendIn(0x20)
    case win.Pressed(pixelgl.KeyEnter):
        k.reg.SendIn(0x80)
    default:
        k.reg.SendIn(0)
    }
}
