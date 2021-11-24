package main

import (
    "reflect"
    "testing"
)

func TestFirstPassAssembler(t *testing.T) {
    for i, tt := range []struct{
        input string
        want []uint16
    }{
        {
            input: "",
            want: []uint16{},
        },
        {
            input: "//COMMENT\n",
            want: []uint16{},
        },
        {
            input: "@AD41\n",
            want: []uint16{0xAD41},
        },
        {
            input: "D=A\n",
            want: []uint16{0xEC10},
        },
        {
            input: "M=0\n",
            want: []uint16{0xEA88},
        },
        {
            input: "A=1\n",
            want: []uint16{0xEFE0},
        },
        {
            input: "AMD=A\n",
            want: []uint16{0xEC38},
        },
        {
            input: "A\n",
            want: []uint16{0xEC00},
        },
        {
            input: "-1\n",
            want: []uint16{0xEE80},
        },
        {
            input: "!M\n",
            want: []uint16{0xFC40},
        },
        {
            input: "D=D+A\n",
            want: []uint16{0xE090},
        },
        {
            input: "M=M+1\n",
            want: []uint16{0xFDC8},
        },
        {
            input: "D;JEQ\n",
            want: []uint16{0xE302},
        },
        {
            input: "0;JMP\n",
            want: []uint16{0xEA87},
        },
        {
            input: "D=D<<3\n",
            want: []uint16{0xC990},
        },
        {
            input: "@0006\nD=M\n@0002\nM=M+1\nA=M-1\nM=D\n",
            want: []uint16{0x6, 0xFC10, 0x2, 0xFDC8, 0xFCA0, 0xE308},
        },
    }{
        input := make([]uint16, len(tt.input))
        for i, r := range tt.input {
            if r == '\n' { r = 0x80 }
            input[i] = uint16(r)
        }
        cpu := NewBarrelShiftCPU()
        computer := NewComputer(cpu)
        computer.LoadProgram(NewROM32K(assembleStatement))
        computer.SendReset(true)
        computer.ClockTick()
        computer.SendReset(false)
        for n, ascii := range input {
            computer.data_mem.ram.mem[0x1000+n] = ascii
        }
        var pprev, prev uint16
        for {
            computer.ClockTick()
            //if i == 15 {
            //t.Errorf("%d: %04x - R6: %04x, M0x2000: %04x", prev, cpu.instr, computer.data_mem.ram.mem[0x6], computer.data_mem.ram.mem[0x2000])
            //t.Errorf("%d: %04x - R6: %04x, A: %04x, D: %04x", prev, cpu.instr, computer.data_mem.ram.mem[0x6], cpu.a.Out(), cpu.d.Out())
            //}
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

