package main

import (
    "io"
    "strings"
    "testing"
)

func TestDecimal(t *testing.T) {
    program, err := Assemble("asm/decimal.asm")
    if err != nil {
        t.Fatal(err)
    }
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    computer.LoadProgram(NewROM32K(program))
    r := strings.NewReader("128\n")
    w := &strings.Builder{}
    computer.data_mem.reader.LoadInputReaders([]io.Reader{r})
    computer.data_mem.writer.LoadOutputWriter(w)
    run(computer)
    if w.String() != "0080\n" {
        t.Errorf("expected %s but got %s", "0080", w.String())
    }
}
