package main

import (
    //"fmt"
    "strconv"
)

// simulate the nand2tetris machine one abstraction level higher
// so we can more easily adapt it to a lispmachine and run experiments
type sim struct {
    a uint16
    d uint16
    pc uint16
    ramCar [16384]uint16
    ramCdr [16384]uint16
}

func NewSim() *sim {
    return &sim{
        ramCar: [16384]uint16{},
        ramCdr: [16384]uint16{},
    }
}

func (s *sim) run(program []string) {
    i := 0
    for {
        i++
        if i > 10000 {
            // prevent infinite loops
            break
        }
        if int(s.pc) >= len(program) {
            break
        }
        instr := program[s.pc]
        //fmt.Println(s.pc, instr)
        n, err := strconv.Atoi(instr)
        if err == nil {
            s.a = uint16(n)
            s.pc += 1
            continue
        }
        switch instr {
        case "A=M": 
            s.a = s.ramCar[s.a]
        case "D=A": 
            s.d = s.a
        case "M=D": 
            s.ramCar[s.a] = s.d
        case "D=M": 
            s.d = s.ramCar[s.a]
        case "0;JMP":
            s.pc = s.a
            continue
        case "D;JEQ":
            if s.d == 0 {
                s.pc = s.a
                continue
            }
        case "D;JNE":
            if s.d != 0 {
                s.pc = s.a
                continue
            }
        // new in lispmachine
        // some of these might overlap; still playing around with which ones
        // i would need in order to build the simplest eval
        case "SETCAR D": // so currently equivalent to M=D. Maybe not if we introduce another register though
            s.ramCar[s.a] = s.d
        case "SETCDR D":
            // with the added dest in Cinstr, this could just be MCDR=D
            s.ramCdr[s.a] = s.d
        case "D=EQLAD":
            if s.a == s.d {
                s.d = 0xffff
            } else {
                s.d = 0
            }
        case "D=EQLMD":
            if s.ramCar[s.a] == s.d {
                s.d = 0xffff
            } else {
                s.d = 0
            }
        case "D=MCDR":
            s.d = s.ramCdr[s.a]
        case "D=ISSYM":
            if s.ramCdr[s.a] != 0 {
                s.d = 0
                s.pc += 1
                continue
            }
            // bit prefix should be 011
            m := s.ramCar[s.a]
            if !nthBit(m, 15) && nthBit(m, 14) && nthBit(m, 13) {
                s.d = 0xffff
            } else {
                s.d = 0
            }
        case "D=ISPRIM":
            if s.ramCdr[s.a] != 0 {
                s.d = 0
                s.pc += 1
                continue
            }
            // bit prefix should be 010
            m := s.ramCar[s.a]
            if !nthBit(m, 15) && nthBit(m, 14) && !nthBit(m, 13) {
                s.d = 0xffff
            } else {
                s.d = 0
            }
        case "D=CDRISEMPTY":
            // bit prefix should be 001
            m := s.ramCdr[s.a]
            if !nthBit(m, 15) && !nthBit(m, 14) && nthBit(m, 13) {
                s.d = 0xffff
            } else {
                s.d = 0
            }
        /*
        // JMP if M (or D) is type:
        // this will only work (or save instr) if we would have another register
        case "JMPISSYM":
        case "JMPISPRIM":
        case "JMPISEMPTY":
        */
        // bonus for debug
        case "RET":
            return
        default:
            panic(instr)
        }
        s.pc += 1
    }
}

func pair(n int) uint16 {
    return uint16(n) // | 0b0000000000000000
}
func emptylist() uint16 {
    return uint16(0b0010000000000000)
}
func symbol(n int) uint16 {
    return uint16(n) | 0b0110000000000000
}
func primitive(n int) uint16 {
    return uint16(n) | 0b0100000000000000
}
func special(n int) uint16 {
    return uint16(n) | 0b1110000000000000
}
func builtin(n int) uint16 {
    return uint16(n) | 0b1010000000000000
}
