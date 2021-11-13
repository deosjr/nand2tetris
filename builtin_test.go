package main

import "testing"

func TestBuiltinRegister(t *testing.T) {
    for _, tt := range []struct{
        sequence func(*BuiltinRegister)
        want uint16
    }{
        {
            sequence: func(b *BuiltinRegister) {
                b.SendIn(42)
            },
            want: 0,
        },
        {
            sequence: func(b *BuiltinRegister) {
                b.SendIn(42)
                b.SendLoad(true)
            },
            want: 0,
        },
        {
            sequence: func(b *BuiltinRegister) {
                b.SendIn(42)
                b.SendLoad(true)
                b.ClockTick()
            },
            want: 42,
        },
    }{
        reg := NewBuiltinRegister()
        tt.sequence(reg)
        got := reg.Out()
        if got != tt.want {
            t.Errorf("got %d but want %d", got, tt.want)
        }
    }
}

func TestCPU(t *testing.T) {
    for i, tt := range []struct{
        sequence func(*CPU)
        wantOutM uint16
        wantWriteM bool
        wantAddressM uint16
        wantPC uint16
    }{
        {
            sequence: func(b *CPU) {
                b.ClockTick()
                b.ClockTick()
            },
            wantPC: 2,
        },
        {
            sequence: func(b *CPU) {
                b.SendInstr(7) // 0000 0000 0000 0111 // @7
                b.ClockTick()
            },
            wantAddressM: 7,
            wantPC: 1,
        },
        {
            sequence: func(b *CPU) {
                b.SendInstr(0x0007) // 0000 0000 0000 0111 // @7
                b.ClockTick()
                b.SendInM(41) // contents of RAM[7]
                b.SendInstr(0xFDD8) // 1111 1101 1101 1000 // MD=M+1
                b.ClockTick()
            },
            wantOutM: 42,
            wantWriteM: true,
            wantAddressM: 7,
            wantPC: 2,
        },
    }{
        cpu := NewCPU()
        tt.sequence(cpu)
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
