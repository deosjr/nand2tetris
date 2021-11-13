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
                //t.Error(b.PC(), b.pcr.Out(), b.pcrl.Out())
                b.SendInstr(0xAA87) // 0;JMP and switch pcrl
                b.ClockTick()
                b.SendInstr(0xEA80) // noop, calculate 0 and dont store nor jump
                b.ClockTick()
                b.SendInstr(0x7FFF) // store highest 15bit value in A
                b.ClockTick()
                b.SendInstr(0xAA80) // noop switching pcrl (could toggle in previous if it wasnt A instr)
                b.ClockTick()
                b.SendInstr(0x87C7) // PCR+1;JMPPCR //1000011111000111
                b.ClockTick()
                b.SendInstr(0xEA80) // noop
                b.ClockTick()
            },
            wantOutM: 0,
            wantWriteM: false,
            wantAddressM: 32767,
            wantPC: 3,
            wantPCR: 10,
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
