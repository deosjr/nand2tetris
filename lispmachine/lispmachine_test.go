package main

import (
    "reflect"
    "strconv"
    "testing"
)

type testCapturer struct {
    out []uint16
}

func (tc *testCapturer) Write(p []byte) (int, error) {
    x, err := strconv.ParseInt(string(p)[:4], 16, 16)
    if err != nil {
        return 0, err
    }
    tc.out = append(tc.out, uint16(x))
    return len(p), nil
}

type testDebugger struct {
    t *testing.T
    i int
    index int
}

func (*testDebugger) BeforeLoop() {}
func (td *testDebugger) BeforeTick(c *LispMachine) {
    td.i++
}
func (td *testDebugger) AfterTick(c *LispMachine) {
    if td.i > 10000 {
        td.t.Fatalf("%d): took too long", td.index)
    }
}

func TestLispMachine(t *testing.T) {
    for i, tt := range []struct{
        in string
        want []uint16
    }{
        {
            in: "5",
            want: []uint16{0x4005},
        },
        {
            in: "x",
            want: []uint16{0x0669}, // symbol not found
        },
        {
            in: "(+ 1 2)",
            want: []uint16{0x4003},
        },
        {
            in: "(- 5 3)",
            want: []uint16{0x4002},
        },
        {
            in: "(define x 42) x",
            want: []uint16{0x0, 0x402a}, // define returns NIL
        },
        {
            in: "(quote lambda)",
            want: []uint16{0x6004},     // symbol 4 = lambda, as per hardcoded symbol table
        },
        {
            in: "(if (> 3 2) 1 fail)",
            want: []uint16{0x4001},
        },
        {
            in: "(define identity (lambda (x) x)) (identity 42)",
            want: []uint16{0x0, 0x402a}, // define returns NIL
        },
        {
            in: "(begin (+ 1 2) (+ 3 4))",
            want: []uint16{0x4007},
        },
    }{
        out, err := compileFromString(tt.in)
        if err != nil {
            t.Fatal(err)
        }
    
        //asm, err := Translate([]string{"vm/eval.vm"}, out)
        asm, err := Translate([]string{}, out)
        if err != nil {
            t.Fatal(err)
        }
    
        program, err := assembleFromString(asm)
        if err != nil {
            t.Fatal(err)
        }
    
        cpu := NewLispCPU()
        computer := NewLispMachine(cpu)
        computer.LoadProgram(NewROM32K(program))
        tc := &testCapturer{}
        computer.data_mem.writer.LoadOutputWriter(tc)
    
        td := &testDebugger{t:t, index:i}
        run(computer, td)

        if !reflect.DeepEqual(tc.out, tt.want) {
            for _, v := range tt.want {
                t.Logf("0x%04x", v)
            }
            for _, v := range tc.out {
                t.Logf("0x%04x", v)
            }
            t.Errorf("%d): want %v got %v", i, tt.want, tc.out)
        }
    }
}
