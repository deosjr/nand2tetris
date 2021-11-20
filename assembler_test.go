package main

import (
    "reflect"
    "testing"
)

func TestAssembler(t *testing.T) {
    for i, tt := range []struct{
        input []uint16
        want []uint16
    }{
        {
            input: []uint16{0x0},
            want: []uint16{},
        },
        {
            // // COMMENT\n
            input: []uint16{0x2F, 0x2F, 0x43, 0x4F, 0x4D, 0x4D, 0x45, 0x4E, 0x54, 0x80},
            want: []uint16{},
        },
        {
            // @AD41\n
            input: []uint16{0x40, 0x41, 0x44, 0x34, 0x31, 0x80},
            want: []uint16{0xAD41},
        },
    }{
        cpu := NewBarrelShiftCPU()
        computer := NewComputer(cpu)
        computer.LoadProgram(NewROM32K(assembleStatement))
        computer.SendReset(true)
        computer.ClockTick()
        computer.SendReset(false)
        for n, ascii := range tt.input {
            computer.data_mem.ram.mem[0x1000+n] = ascii
        }
        var pprev, prev uint16
        for {
            computer.ClockTick()
            //t.Errorf("%d: %04x - R6: %04x, M0x2000: %04x", prev, cpu.instr, computer.data_mem.ram.mem[0x6], computer.data_mem.ram.mem[0x2000])
            // detect end loop
            if pprev == cpu.PC() {
                break
            }
            pprev = prev
            prev = cpu.PC()
        }
        output := []uint16{}
        n := 0
        for {
            v := computer.data_mem.ram.mem[0x2000+n]
            if v == 0 {
                break
            }
            output = append(output, v)
            n++
        }
        if !reflect.DeepEqual(output, tt.want) {
            t.Errorf("%d) got %d but want %d", i, output, tt.want)
        }
    }
}

