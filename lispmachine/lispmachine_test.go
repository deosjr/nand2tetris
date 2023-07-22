package main

import "testing"

func TestLispMachine(t *testing.T) {
    for i, tt := range []struct{
        program []uint16
        wantOutCarM uint16
        wantOutCdrM uint16
        wantWriteCarM bool
        wantWriteCdrM bool
        wantAddressM uint16
        wantPC uint16
        n int
    }{
        {
            program: []uint16{0x0, 0x0},
            n: 2,
            wantOutCarM: 0,
            wantWriteCarM: false,
            wantAddressM: 0,
            wantPC: 2,
        },
        {
            // load 5, move A to D, load 0, move D to M
            program: []uint16{0x5, 0xEC10, 0x0, 0xE308},
            n: 4,
            wantOutCarM: 5,
            wantWriteCarM: true,
            wantAddressM: 0,
            wantPC: 4,
        },
    }{
        cpu := NewLispCPU()
        computer := NewLispMachine(cpu)
        computer.LoadProgram(NewROM32K(tt.program))
        computer.SendReset(true)
        computer.ClockTick()
        computer.SendReset(false)
        for i:=0; i<tt.n; i++ {
            computer.ClockTick()
        }
        if cpu.OutCarM() != tt.wantOutCarM {
            t.Errorf("%d) got %d but wantOutCarM %d", i, cpu.OutCarM(), tt.wantOutCarM)
        }
        if cpu.WriteCarM() != tt.wantWriteCarM {
            t.Errorf("%d) got %t but wantWriteCarM %t", i, cpu.WriteCarM(), tt.wantWriteCarM)
        }
        if cpu.AddressM() != tt.wantAddressM {
            t.Errorf("%d) got %d but wantAddressM %d", i, cpu.AddressM(), tt.wantAddressM)
        }
        if cpu.PC() != tt.wantPC {
            t.Errorf("%d) got %d but wantPC %d", i, cpu.PC(), tt.wantPC)
        }
    }
}
