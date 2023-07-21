package main

import "testing"

func TestSim(t *testing.T) {
    for i, tt := range []struct{
        program []string
        wantA uint16
        wantD uint16
    }{
        {
            // SET M TO 5 THROUGH D
            program: []string{
                "5",
                "D=A",
                "0",
                "M=D",
            },
            wantA: 0,
            wantD: 5,
        },
        {
            // CONS
            program: []string{
                "5",
                "D=A",
                "0", // destination should come from alloc(1)
                "SETCAR D",
                "7",
                "D=A",
                "0",
                "SETCDR D",
            },
            wantA: 0,
            wantD: 7,
        },
    }{
        sim := NewSim()
        sim.run(tt.program)
        if sim.a != tt.wantA {
            t.Errorf("%d) wantA %d got %d", i, tt.wantA, sim.a)
        }
        if sim.d != tt.wantD {
            t.Errorf("%d) wantD %d got %d", i, tt.wantD, sim.d)
        }
    }
}

func pair(n int) uint16 {
    return uint16(n) // | 0b0000000000000000
}
func symbol(n int) uint16 {
    return uint16(n) | 0b0110000000000000
}
func primitive(n int) uint16 {
    return uint16(n) | 0b0100000000000000
}

func TestSimASSQ(t *testing.T) {
    // TODO: type checks throughout?
    program := []string{
        // ASSQ K P where K is 3 and P is 1
        // meaning: find key 3 in assoc list starting at p1
        // we assume K is a symbol and P is a pair
        // for now we'll assume we can use @R11-@R13 as vars
        "24579", // symbol 3, or 0x6003
        "D=A",
        "11",  // @R11
        "M=D", // @R11 = K
        "1",   // pointer 1, or 0x0001
        "D=A",
        "12",  // @R12
        "M=D", // @R12 = P
        // (LOOP) : 8
        "12",
        "A=M", // A = P
        "A=M", // follow pointer to cell
        "D=M", // D = CAR cell
        "11",
        "D=EQLMD", // compare K and CAR cell, store true/false in D
        "23",      // label CONTINUE
        "D;JEQ",   // jump to label if K is not D
        // here K == D!
        "12",
        "A=M", // A = P
        "A=M", // follow pointer to cell
        "D=MCDR", // D = CDR cell
        "13",
        "M=D", // write found matching value to @R13 !
        "RET",
        // (CONTINUE) : 23
        "12",
        "A=M", // A = P
        "D=MCDR", // D = CDR cell
        // TODO: if D = NIL, we have failed to find!
        "12",  // @R12 follows cdr of list
        "M=D",
        "8",   // label LOOP
        "0;JMP",
    }
    sim := NewSim()
    // set assoc pair list of ((1 . 6) (2 . 7) (3 . 8) (4 . 9) (5. 10)) into ram
    // where key is a symbol and value is a primitive (number) using type flag
    // ram[1] contains head of the list as a cell of (p1 . p2) where p1 is pointer to car and p2 to cdr
    // ram[2] through ram[5] contain the rest of the cons cells of the assoc list
    // ram[6] through ram[10] contain the actual pairs of the assoc list
    for i:=1; i<6; i++ {
        sim.ramCar[i] = pair(i+5)
        sim.ramCdr[i] = pair(i+1)
        if i == 5 {
            sim.ramCdr[i] = pair(0) // NIL
        }
        sim.ramCar[i+5] = symbol(i)
        sim.ramCdr[i+5] = primitive(i+5)
    }
    sim.run(program)
    if sim.ramCar[13] != primitive(8) {
        t.Errorf("Expected assoc 3 to find 8, but got %x instead", sim.ramCar[13])
        t.Errorf("%x\n", sim.ramCar[:15])
        t.Errorf("%x\n", sim.ramCdr[:15])
    }
}
