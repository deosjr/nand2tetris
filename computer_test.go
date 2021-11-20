package main

import "testing"

func TestComputer(t *testing.T) {
    for i, tt := range []struct{
        program []uint16
        wantOutM uint16
        wantWriteM bool
        wantAddressM uint16
        wantPC uint16
        n int
    }{
        {
            program: []uint16{0x0, 0x0},
            n: 2,
            wantOutM: 0,
            wantWriteM: false,
            wantAddressM: 0,
            wantPC: 2,
        },
        {
            program: []uint16{0x5, 0xEC10, 0x0, 0xE308},
            n: 4,
            wantOutM: 5,
            wantWriteM: true,
            wantAddressM: 0,
            wantPC: 4,
        },
    }{
        cpu := NewBuiltinCPU()
        computer := NewComputer(cpu)
        computer.LoadProgram(NewROM32K(tt.program))
        computer.SendReset(true)
        computer.ClockTick()
        computer.SendReset(false)
        for i:=0; i<tt.n; i++ {
            computer.ClockTick()
        }
        if cpu.OutM() != tt.wantOutM {
            t.Errorf("%d) got %d but wantOutM %d", i, cpu.OutM(), tt.wantOutM)
        }
        if cpu.WriteM() != tt.wantWriteM {
            t.Errorf("%d) got %t but wantWriteM %t", i, cpu.WriteM(), tt.wantWriteM)
        }
        if cpu.AddressM() != tt.wantAddressM {
            t.Errorf("%d) got %d but wantAddressM %d", i, cpu.AddressM(), tt.wantAddressM)
        }
        if cpu.PC() != tt.wantPC {
            t.Errorf("%d) got %d but wantPC %d", i, cpu.PC(), tt.wantPC)
        }
    }
}
