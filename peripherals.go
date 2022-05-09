package main

import (
    "bufio"
    "fmt"
    "image/color"
    "io"
    "os"

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

// writes values from an input file to a specific register
// Load clears the register and signals to read the next character
// -> so writing to this register sets it to 0 instead
type TapeReader interface {
    Out() uint16
    SendLoad(bool)
    ClockTick()
    LoadInputTape(string) error
    LoadInputTapes([]string) error
    LoadInputReaders([]io.Reader)
}

// reads values to write to an output file or stdout
// sets the register to 0 when it is _done_ writing a character
type TapeWriter interface {
    Out() uint16
    SendIn(uint16)
    SendLoad(bool)
    ClockTick()
    // NOTE: truncates!
    LoadOutputTape(string) error
    LoadOutputWriter(io.Writer)
}

type Screen512x256 struct {
    ram *BuiltinRAM16K // actually use only the first 8K
}

func NewScreen512x256() *Screen512x256 {
    return &Screen512x256{
        ram: NewBuiltinRAM16K(),
    }
}

func (s *Screen512x256) SendIn(in uint16) {
    s.ram.SendIn(in)
}

func (s *Screen512x256) SendLoad(load bool) {
    s.ram.SendLoad(load)
}

func (s *Screen512x256) SendAddress(addr uint16) {
    if addr >= 8192 {
        return
        //panic("access screen memory beyond 8K")
    }
    s.ram.SendAddress(addr)
}

func (s *Screen512x256) Out() uint16 {
    return s.ram.Out()
}

func (s *Screen512x256) ClockTick() {
    s.ram.ClockTick()
}

func (s *Screen512x256) RunScreen(win *pixelgl.Window) {
    // listen to mem and show in a window using pixelgl
    // NOTE: pixelgl y increases UP, nand2tetris DOWN
    white := color.RGBA{255,255,255,0}
    black := color.RGBA{0,0,0,0}
    pd := pixel.MakePictureData(win.Bounds())
    for row:=0;row<256;row++ {
        for w:=0;w<32;w++ {
            addr := row*32+w
            word := s.ram.mem[addr]
            for c:=0;c<16;c++ {
                b := nthBit(word, uint16(15-c))
                invrow := (255-row)
                if b {
                    pd.Pix[invrow*512+16*w+c] = white
                } else {
                    pd.Pix[invrow*512+16*w+c] = black
                }
            }
        }
    }
    sprite := pixel.NewSprite(pd, pd.Bounds())
    sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
}

// Lets leave this one for posterity (ie old asm files)
// but I misunderstood the default screen dimensions :)
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

// TODO: separate textmode screen that only writes characters?
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
                b := nthBit(word, uint16(15-c))
                if b {
                    invrow := (511-row)
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

func upperIfShift(upper uint16, shift bool) uint16 {
    if shift {
        return upper
    }
    return upper + 32
}

func (k *SimpleKeyboard) RunKeyboard(win *pixelgl.Window) {
    shift :=  win.Pressed(pixelgl.KeyLeftShift) || win.Pressed(pixelgl.KeyRightShift)
    switch {
    case win.Pressed(pixelgl.KeyComma):
        if shift {
            k.reg.SendIn(0x3C)
        } else {
            k.reg.SendIn(0x2C)
        }
    case win.Pressed(pixelgl.KeyPeriod):
        if shift {
            k.reg.SendIn(0x3E)
        } else {
            k.reg.SendIn(0x2E)
        }
    case win.Pressed(pixelgl.Key0):
        if shift {
            k.reg.SendIn(0x29)
        } else {
            k.reg.SendIn(0x30)
        }
    case win.Pressed(pixelgl.Key1):
        if shift {
            k.reg.SendIn(0x21)
        } else {
            k.reg.SendIn(0x31)
        }
    case win.Pressed(pixelgl.Key2):
        if shift {
            k.reg.SendIn(0x40)
        } else {
            k.reg.SendIn(0x32)
        }
    case win.Pressed(pixelgl.Key3):
        k.reg.SendIn(0x33)
    case win.Pressed(pixelgl.Key4):
        k.reg.SendIn(0x34)
    case win.Pressed(pixelgl.Key5):
        k.reg.SendIn(0x35)
    case win.Pressed(pixelgl.Key6):
        k.reg.SendIn(0x36)
    case win.Pressed(pixelgl.Key7):
        k.reg.SendIn(0x37)
    case win.Pressed(pixelgl.Key8):
        if shift {
            k.reg.SendIn(0x2A)
        } else {
            k.reg.SendIn(0x38)
        }
    case win.Pressed(pixelgl.Key9):
        if shift {
            k.reg.SendIn(0x28)
        } else {
            k.reg.SendIn(0x39)
        }
    case win.Pressed(pixelgl.KeyMinus):
        k.reg.SendIn(0x2D)
    case win.Pressed(pixelgl.KeySemicolon):
        k.reg.SendIn(0x3B)
    case win.Pressed(pixelgl.KeyEqual):
        if shift {
            k.reg.SendIn(0x2B)
        } else {
            k.reg.SendIn(0x3D)
        }
    case win.Pressed(pixelgl.KeyA):
        k.reg.SendIn(upperIfShift(0x41, shift))
    case win.Pressed(pixelgl.KeyB):
        k.reg.SendIn(upperIfShift(0x42, shift))
    case win.Pressed(pixelgl.KeyC):
        k.reg.SendIn(upperIfShift(0x43, shift))
    case win.Pressed(pixelgl.KeyD):
        k.reg.SendIn(upperIfShift(0x44, shift))
    case win.Pressed(pixelgl.KeyE):
        k.reg.SendIn(upperIfShift(0x45, shift))
    case win.Pressed(pixelgl.KeyF):
        k.reg.SendIn(upperIfShift(0x46, shift))
    case win.Pressed(pixelgl.KeyG):
        k.reg.SendIn(upperIfShift(0x47, shift))
    case win.Pressed(pixelgl.KeyH):
        k.reg.SendIn(upperIfShift(0x48, shift))
    case win.Pressed(pixelgl.KeyI):
        k.reg.SendIn(upperIfShift(0x49, shift))
    case win.Pressed(pixelgl.KeyJ):
        k.reg.SendIn(upperIfShift(0x4A, shift))
    case win.Pressed(pixelgl.KeyK):
        k.reg.SendIn(upperIfShift(0x4B, shift))
    case win.Pressed(pixelgl.KeyL):
        k.reg.SendIn(upperIfShift(0x4C, shift))
    case win.Pressed(pixelgl.KeyM):
        k.reg.SendIn(upperIfShift(0x4D, shift))
    case win.Pressed(pixelgl.KeyN):
        k.reg.SendIn(upperIfShift(0x4E, shift))
    case win.Pressed(pixelgl.KeyO):
        k.reg.SendIn(upperIfShift(0x4F, shift))
    case win.Pressed(pixelgl.KeyP):
        k.reg.SendIn(upperIfShift(0x50, shift))
    case win.Pressed(pixelgl.KeyQ):
        k.reg.SendIn(upperIfShift(0x51, shift))
    case win.Pressed(pixelgl.KeyR):
        k.reg.SendIn(upperIfShift(0x52, shift))
    case win.Pressed(pixelgl.KeyS):
        k.reg.SendIn(upperIfShift(0x53, shift))
    case win.Pressed(pixelgl.KeyT):
        k.reg.SendIn(upperIfShift(0x54, shift))
    case win.Pressed(pixelgl.KeyU):
        k.reg.SendIn(upperIfShift(0x55, shift))
    case win.Pressed(pixelgl.KeyV):
        k.reg.SendIn(upperIfShift(0x56, shift))
    case win.Pressed(pixelgl.KeyW):
        k.reg.SendIn(upperIfShift(0x57, shift))
    case win.Pressed(pixelgl.KeyX):
        k.reg.SendIn(upperIfShift(0x58, shift))
    case win.Pressed(pixelgl.KeyY):
        k.reg.SendIn(upperIfShift(0x59, shift))
    case win.Pressed(pixelgl.KeyZ):
        k.reg.SendIn(upperIfShift(0x5A, shift))
    case win.Pressed(pixelgl.KeySpace):
        k.reg.SendIn(0x20)
    case win.Pressed(pixelgl.KeyEnter):
        k.reg.SendIn(0x80)
    case win.Pressed(pixelgl.KeyBackspace):
        k.reg.SendIn(0x08)
    default:
        k.reg.SendIn(0)
    }
}

type tapeReader struct {
    reg *BuiltinRegister
    scanner *bufio.Scanner
    eof chan bool
}

func NewTapeReader() *tapeReader {
    reg := NewBuiltinRegister()
    reg.SendLoad(true)
    return &tapeReader{
        reg: reg,
        eof: make(chan bool),
    }
}

func (tr *tapeReader) Out() uint16 {
    return tr.reg.Out()
}

func (tr *tapeReader) SendLoad(load bool) {
    if load {
        tr.reg.SendIn(0)
    }
}

func (tr *tapeReader) ClockTick() {
    if tr.scanner == nil || tr.reg.Out() != 0 {
        tr.reg.ClockTick()
        return
    }
    // Here's where we abstract heavily over how this would work exactly
    if tr.reg.Out() == 0 {
        if !tr.scanner.Scan() {
            tr.reg.SendIn(0x1C) // ascii File Separator control character
            tr.reg.ClockTick()
            tr.eof <- true
            return
        }
        char := tr.scanner.Text()[0]
        if char == '\n' {
            char = 0x80 // newlines are ENTER ascii values
        }
        tr.reg.SendIn(uint16(char))
    }
    tr.reg.ClockTick()
}

func (tr *tapeReader) LoadInputTape(filename string) error {
    return tr.LoadInputTapes([]string{filename})
}

// TODO: closing. using ReadCloser disallows strings.NewReader
// NOTE: currently opens all files at once
func (tr *tapeReader) LoadInputTapes(filenames []string) error {
    readers := []io.Reader{}
    for _, filename := range filenames {
        f, err := os.Open(filename)
        if err != nil {
            return err
        }
        readers = append(readers, f)
    }
    tr.LoadInputReaders(readers)
    return nil
}

func (tr *tapeReader) LoadInputReaders(readers []io.Reader) {
    go func() {
        for _, r := range readers {
            tr.scanner = bufio.NewScanner(r)
            tr.scanner.Split(bufio.ScanRunes)
            <-tr.eof
        }
    }()
}

type tapeWriter struct {
    reg *BuiltinRegister
    out int
    w io.Writer
}

func NewTapeWriter() *tapeWriter {
    reg := NewBuiltinRegister()
    return &tapeWriter{
        reg: reg,
        w:   os.Stdout,
    }
}

func (tr *tapeWriter) Out() uint16 {
    return tr.reg.Out()
}

func (tr *tapeWriter) SendIn(in uint16) {
    tr.reg.SendIn(in)
}

func (tr *tapeWriter) SendLoad(load bool) {
    tr.reg.SendLoad(load)
    if load {
        tr.out = -1
    }
}

func (tr *tapeWriter) ClockTick() {
    tr.reg.ClockTick()
    out := tr.reg.Out()
    if int(out) != tr.out {
        fmt.Fprintf(tr.w, "%04x\n", out)
        // this is of course cheating, should be represented by separate write bit
        tr.out = int(out)
    }
}

func (tr *tapeWriter) LoadOutputTape(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    tr.LoadOutputWriter(f)
    return nil
}

func (tr *tapeWriter) LoadOutputWriter(w io.Writer) {
    tr.w = w
}
