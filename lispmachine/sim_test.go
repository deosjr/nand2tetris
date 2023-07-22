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
            sim.ramCdr[i] = emptylist()
        }
        sim.ramCar[i+5] = symbol(i)
        sim.ramCdr[i+5] = primitive(i+5)
    }
    sim.run(program)
    if sim.ramCar[13] != primitive(8) {
        t.Errorf("Expected assoc 3 to find 8, but got %04x instead", sim.ramCar[13])
        t.Errorf("%x\n", sim.ramCar[:15])
        t.Errorf("%x\n", sim.ramCdr[:15])
    }
}

func TestSimpleEval(t *testing.T) {
    // one global env as assoc list, no nested envs yet
    // parsed sexpression starts in memory, we will eval it
    // env is list of (symbol . sexpression) pairs,
    // where most sexpressions will be builtin procs which mean
    // jumps in instruction memory
    program := []string{
        "100", // start of env
        "D=A",
        "2",   // ENV
        "M=D",
        "111", // start of expression e to eval
        "D=A",
        "3",   // ARG
        "M=D",
        // (EVAL)
        "3",
        "A=M",
        "D=ISSYM",
        "19", // label EVALSYMBOL
        "D;JNE",
        "3",
        "A=M",
        "D=ISPRIM",
        "30", // label EVALPRIM
        "D;JNE",
        // TODO fallthrough: procedure call
        // which includes if/define and other special funcs checked first
        "RET",

        // (EVALSYMBOL) : 19
        // TODO return ASSQ ARG ENV or error if not found
        "2",
        "D=M",
        "12",
        "M=D",
        "3",
        "A=M",
        "D=M",
        "11",
        "M=D",
        "36", // label ASSQ
        "0;JMP",

        // (EVALPRIM) : 30
        "3",
        "A=M",
        "D=M",
        "13",
        "M=D",
        "RET",

        // (ASSQ) : 36
        // assume @R11 = K and @R12 = P
        "12",
        "A=M", // A = P
        "A=M", // follow pointer to cell
        "D=M", // D = CAR cell
        "11",
        "D=EQLMD", // compare K and CAR cell, store true/false in D
        "51",      // label ASSQCONTINUE
        "D;JEQ",   // jump to label if K is not D
        // here K == D!
        "12",
        "A=M", // A = P
        "A=M", // follow pointer to cell
        "D=MCDR", // D = CDR cell
        "13",
        "M=D", // write found matching value to @R13 !
        "RET",
        // (ASSQCONTINUE) : 51
        "12",
        "A=M", // A = P
        // TODO: if D = emptylist, we have failed to find!
        "D=CDRISEMPTY",
        "63", // (FAILTOFIND)
        "D;JNE",
        "12",
        "A=M", // A = P
        "D=MCDR", // D = CDR cell
        "12",  // @R12 follows cdr of list
        "M=D",
        "36",   // label ASSQ
        "0;JMP",
        // (FAILTOFIND) : 63
        "RET",
        // (PLUS) : 
    }
    sim := NewSim()
    // set assoc pair list of ((1 . 6) (2 . 7) (3 . 8) (4 . 9) (5 . plus)) into ram
    // where key is a symbol and value is a primitive (number) except plus, which is a builtin proc
    // you can imagine symboltable including ("x" . s2) ("+" . s5) pairs
    // this symboltable should start with some special symbols like "if", because for those
    // we will not eval their args first
    // ram[100] contains head of the list as a cell of (p1 . p2) where p1 is pointer to car and p2 to cdr
    // ram[102] through ram[105] contain the rest of the cons cells of the assoc list
    // ram[106] through ram[110] contain the actual pairs of the assoc list
    for i:=100; i<105; i++ {
        sim.ramCar[i] = pair(i+5)
        sim.ramCdr[i] = pair(i+1)
        sim.ramCar[i+5] = symbol(i-99)
        sim.ramCdr[i+5] = primitive(i-94)
        if i == 104 {
            sim.ramCdr[i] = emptylist()
            sim.ramCdr[i+5] = builtin(0) // PLUS
        }
    }
    // ram[111] and onwards contains the sexpression to eval
    // first lets eval a number, like 42
    sim.ramCar[111] = primitive(42)
    sim.run(program)
    if sim.ramCar[13] != primitive(42) {
        t.Errorf("Expected 42, but got %04x instead", sim.ramCar[13])
        t.Errorf("%x\n", sim.ramCar[100:114])
        t.Errorf("%x\n", sim.ramCdr[100:114])
        t.Error(sim.pc)
    }
    // then lets eval x
    sim.ramCar[111] = symbol(2)
    sim.pc = 0
    sim.ramCar[13] = 0
    sim.run(program)
    if sim.ramCar[13] != primitive(7) {
        t.Errorf("Expected 7, but got %04x instead", sim.ramCar[13])
        t.Errorf("%x\n", sim.ramCar[:16])
        t.Errorf("%x\n", sim.ramCar[100:114])
        t.Errorf("%x\n", sim.ramCdr[100:114])
        t.Error(sim.pc)
    }
    // now we will eval (+ x 35)
    /*
    sim.ramCar[111], sim.ramCdr[111] = symbol(5), pair(112)
    sim.ramCar[112], sim.ramCdr[112] = symbol(2), pair(113)
    sim.ramCar[113], sim.ramCdr[113] = primitive(35), emptylist()
    sim.run(program)
    if sim.ramCar[13] != primitive(42) {
        t.Errorf("Expected 42, but got %x instead", sim.ramCar[13])
        t.Errorf("%x\n", sim.ramCar[100:114])
        t.Errorf("%x\n", sim.ramCdr[100:114])
    }
    */
}
