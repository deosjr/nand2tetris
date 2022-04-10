package main

import (
    "io"
    "strconv"
    "strings"
    "testing"
)

type hexCapture struct {
    out []uint16
}

func (h *hexCapture) Write(p []byte) (int, error) {
    // some big assumptions here on how tapeWriter writes
    x, err := strconv.ParseInt(string(p)[:4], 16, 16)
    if err != nil {
        return 0, err
    }
    h.out = append(h.out, uint16(x))
    return 2, nil
}

func TestDecimal(t *testing.T) {
    program, err := Assemble("asm/decimal.asm")
    if err != nil {
        t.Fatal(err)
    }
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    computer.LoadProgram(NewROM32K(program))
    r := strings.NewReader("128\n")
    w := &hexCapture{}
    computer.data_mem.reader.LoadInputReaders([]io.Reader{r})
    computer.data_mem.writer.LoadOutputWriter(w)
    run(computer, nil)
    if len(w.out) != 1 {
        t.Fatalf("expected w.out of len 1, but got %v", w.out)
    }
    if w.out[0] != 0x0080 {
        t.Errorf("expected %04x but got %04x", 0x0080, w.out[0])
    }
}

type testDebugger struct {
    t *testing.T
    steps int
    w *hexCapture
}

func (*testDebugger) BeforeLoop() {}
func (*testDebugger) BeforeTick(c *Computer) {}
func (td *testDebugger) AfterTick(c *Computer) {
/*
    sp := c.data_mem.ram.mem[0]
    stack := [20]uint16{}
    for i:=0; i<20; i++ {
        stack[i] = c.data_mem.ram.mem[256+i]
    }
    td.t.Logf("SP: %d STACK: %v\n", sp, stack)
    td.steps++
    if td.steps > 500 {
        td.t.Fatalf("loop: %v", td.w.out)
    }
    */
}

func TestVMMult(t *testing.T) {
    program, err := Assemble("asm/vm_mult.asm")
    if err != nil {
        t.Fatal(err)
    }
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    computer.LoadProgram(NewROM32K(program))
    w := &hexCapture{}
    computer.data_mem.writer.LoadOutputWriter(w)
    run(computer, &testDebugger{t:t, w:w})
    if len(w.out) != 1 {
        t.Fatalf("expected w.out of len 1, but got %v", w.out)
    }
    if w.out[0] != 6 {
        t.Errorf("expected %04x but got %04x", 6, w.out[0])
    }
}
