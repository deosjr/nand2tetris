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
