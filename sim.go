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
    for {
        if int(s.pc) >= len(program) {
            break
        }
        instr := program[s.pc]
        //fmt.Println(instr)
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
        // new in lispmachine
        case "SETCAR D": // so currently equivalent to M=D. Maybe not if we introduce another register though
            s.ramCar[s.a] = s.d
        case "SETCDR D":
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
        // bonus for debug
        case "RET":
            return
        default:
            panic(instr)
        }
        s.pc += 1
    }
}
