package main

import "testing"

func TestBuiltinCPU(t *testing.T) {
    for i, tt := range []struct{
        sequence func(*BuiltinCPU)
        wantOutM uint16
        wantWriteM bool
        wantAddressM uint16
        wantPC uint16
    }{
        {
            sequence: func(b *BuiltinCPU) {
                b.ClockTick()
                b.ClockTick()
            },
            wantPC: 2,
        },
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(7) // 0000 0000 0000 0111 // @7
                b.ClockTick()
            },
            wantAddressM: 7,
            wantPC: 1,
        },
        {
            sequence: func(b *BuiltinCPU) {
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
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(0)
                b.ClockTick()
                b.SendInstr(0xEC10)
                b.ClockTick()
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xE302)
                b.ClockTick()
            },
            wantPC: 7,
            wantAddressM: 7,
        },
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xEC10)
                b.ClockTick()
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xE302)
                b.ClockTick()
            },
            wantPC: 4,
            wantAddressM: 7,
            wantOutM: 7,
        },
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(0)
                b.ClockTick()
                b.SendInstr(0xEC10)
                b.ClockTick()
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xE305)
                b.ClockTick()
            },
            wantPC: 4,
            wantAddressM: 7,
        },
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xEC10)
                b.ClockTick()
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xE305)
                b.ClockTick()
            },
            wantPC: 7,
            wantAddressM: 7,
            wantOutM: 7,
        },
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xEC08) // M=A
                b.ClockTick()
                b.SendInM(7) // contents of RAM[7]
                b.SendInstr(0xFC98) // DM=M-1
                b.ClockTick()
            },
            wantPC: 3,
            wantAddressM: 7,
            wantOutM: 6,
            wantWriteM: true,
        },
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xEC10) // D=7
                b.ClockTick()
                b.SendInM(3) // contents of RAM[7]
                b.SendInstr(0xF098) // DM=D+M
                b.ClockTick()
            },
            wantPC: 3,
            wantOutM: 10,
            wantWriteM: true,
            wantAddressM: 7,
        },
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(7)
                b.ClockTick()
                b.SendInM(3) // contents of RAM[7]
                b.SendInstr(0xFCA8) // AM=M-1
                b.ClockTick()
            },
            wantPC: 2,
            wantOutM: 2,
            wantWriteM: true,
            wantAddressM: 2,
        },
        {
            sequence: func(b *BuiltinCPU) {
                b.SendInstr(0)
                b.ClockTick()
                b.SendInstr(0xEC10) // D=0
                b.ClockTick()
                b.SendInstr(7)
                b.ClockTick()
                b.SendInstr(0xE301) // D;JGT
                b.ClockTick()
            },
            wantPC: 4,
            wantOutM: 0,
            wantWriteM: false,
            wantAddressM: 7,
        },
    }{
        cpu := NewBuiltinCPU()
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

func TestPCRegisterCPU(t *testing.T) {
    for i, tt := range []struct{
        sequence func(*PCRegisterCPU)
        wantOutM uint16
        wantWriteM bool
        wantAddressM uint16
        wantPC uint16
        wantPCR uint16
    }{
        {
            sequence: func(b *PCRegisterCPU) {
                b.ClockTick()
                b.ClockTick()
            },
            wantPC: 2,
            wantPCR: 1,
        },
        {
            sequence: func(b *PCRegisterCPU) {
                b.SendInstr(7) // 0000 0000 0000 0111 // @7
                b.ClockTick()
            },
            wantAddressM: 7,
            wantPC: 1,
            wantPCR: 0,
        },
        {
            sequence: func(b *PCRegisterCPU) {
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
            wantPCR: 1,
        },
        {
            sequence: func(b *PCRegisterCPU) {
                b.SendInstr(0x0007) // @7
                b.ClockTick()
                b.SendInstr(0xAA87) // 0;JMP and switch pcrl
                b.ClockTick()
                b.SendInstr(0xFFFF) // store highest 16(!)bit value in A
                b.ClockTick()
                //b.SendInstr(0xEA80) // noop
                //b.ClockTick()
                //t.Error(b.PC(), b.pcr.Out(), b.pcrl.Out(), b.a.Out(), b.WriteM())
            },
            wantOutM: 1,
            wantWriteM: false,
            wantAddressM: 65535,
            wantPC: 2,
            wantPCR: 7,
        },
    }{
        cpu := NewPCRegisterCPU()
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
        if cpu.pcr.Out() != tt.wantPCR {
            t.Errorf("%d) got %d but wantPCR %d", i, cpu.pcr.Out(), tt.wantPCR)
        }
    }
}

func TestBarrelShiftCPU(t *testing.T) {
    for i, tt := range []struct{
        sequence func(*BarrelShiftCPU)
        wantOutM uint16
        wantWriteM bool
        wantAddressM uint16
        wantPC uint16
    }{
        {
            sequence: func(b *BarrelShiftCPU) {
                b.SendInstr(0x0007) // @7
                b.ClockTick()
                b.SendInstr(0xC280) // << 5
                b.ClockTick()
                //b.SendInstr(0xEA80) // noop
                //b.ClockTick()
                //t.Error(b.PC(), b.pcr.Out(), b.pcrl.Out(), b.a.Out(), b.WriteM())
            },
            wantOutM: 224,
            wantWriteM: false,
            wantAddressM: 7,
            wantPC: 2,
        },
        {
            sequence: func(b *BarrelShiftCPU) {
                b.SendInstr(0x0007) // @7
                b.ClockTick()
                b.SendInM(0x0004) // M[7] = 4
                b.SendInstr(0xD208) // << 4
                b.ClockTick()
            },
            wantOutM: 64,
            wantWriteM: true,
            wantAddressM: 7,
            wantPC: 2,
        },
    }{
        cpu := NewBarrelShiftCPU()
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
