package main

import (
    "fmt"
    "io"
    "os"
    "strconv"
    "strings"
    "testing"
)

type hexCapture struct {
    out []uint16
}

func (h *hexCapture) Write(p []byte) (int, error) {
    // some big assumptions here on how tapeWriter writes
    x, err := strconv.ParseInt(string(p)[:4], 16, 16)
    if err != nil {
        return 0, err
    }
    h.out = append(h.out, uint16(x))
    return 2, nil
}

func TestDecimal(t *testing.T) {
    program, err := Assemble("asm/decimal.asm")
    if err != nil {
        t.Fatal(err)
    }
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    computer.LoadProgram(NewROM32K(program))
    r := strings.NewReader("128\n")
    w := &hexCapture{}
    computer.data_mem.reader.LoadInputReaders([]io.Reader{r})
    computer.data_mem.writer.LoadOutputWriter(w)
    run(computer, nil)
    if len(w.out) != 1 {
        t.Fatalf("expected w.out of len 1, but got %v", w.out)
    }
    if w.out[0] != 0x0080 {
        t.Errorf("expected %04x but got %04x", 0x0080, w.out[0])
    }
}

type testDebugger struct {
    t *testing.T
    steps int
    w *hexCapture
}

func (*testDebugger) BeforeLoop() {}
func (*testDebugger) BeforeTick(c *Computer) {}
func (td *testDebugger) AfterTick(c *Computer) {
    td.steps++
/*
    sp := c.data_mem.ram.mem[0]
    stack := [20]uint16{}
    for i:=0; i<20; i++ {
        stack[i] = c.data_mem.ram.mem[256+i]
    }
    td.t.Logf("SP: %d STACK: %v\n", sp, stack)
    if td.steps > 500 {
        td.t.Fatalf("loop: %v", td.w.out)
    }
    */
}

func TestVMMult(t *testing.T) {
    program, err := Assemble("asm/vm_mult.asm")
    if err != nil {
        t.Fatal(err)
    }
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    computer.LoadProgram(NewROM32K(program))
    w := &hexCapture{}
    computer.data_mem.writer.LoadOutputWriter(w)
    run(computer, &testDebugger{t:t, w:w})
    if len(w.out) != 1 {
        t.Fatalf("expected w.out of len 1, but got %v", w.out)
    }
    if w.out[0] != 6 {
        t.Errorf("expected %04x but got %04x", 6, w.out[0])
    }
}

func TestJackMult(t *testing.T) {
    main := `// test main
    func main() {
        x := mult.mult(42, 36)
        print(x)
        return
    }`
    mult := `// naive multiplication
    func mult(x int, y int) int {
        sum := 0
        for j:=0; j<y; j++ {
            sum = sum + x
        }
        return sum
    }`
    testlib := []string{
        "jack/memory.jack",
        "jack/array.jack",
        "jack/list.jack",
    }
    filenames := []string{"main.jack", "mult.jack"}
    contents := []string{main, mult}
    for _, lib := range testlib {
        data, err := os.ReadFile(lib)
        if err != nil {
            t.Fatal(err)
        }
        filenames = append(filenames, lib)
        contents = append(contents, string(data))
    }
    program, err := compile(filenames, contents)
    if err != nil {
        t.Fatal(err)
    }
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    computer.LoadProgram(NewROM32K(program))
    w := &hexCapture{}
    computer.data_mem.writer.LoadOutputWriter(w)
    td := &testDebugger{t:t, w:w}
    run(computer, td)
    if len(w.out) != 1 {
        t.Fatalf("expected w.out of len 1, but got %v", w.out)
    }
    if w.out[0] != 1512 {
        t.Errorf("expected %04x but got %04x", 1512, w.out[0])
    }
    t.Fatal(td.steps) // 4521
}

func TestQuarterSquare(t *testing.T) {
    table := []string{}
    for i:=0; i<100; i++ {
        sq := (i*i)/4
        line := fmt.Sprintf("t[%d] = %d", i, sq)
        table = append(table, line)
    }
    main := fmt.Sprintf(`// test main
    func main() {
        t := 100
        %s
        x := mult.mult(42, 36)
        print(x)
        return
    }`, strings.Join(table, "\n"))
    mult := `// quarter square multiplication
    func mult(x int, y int) int {
        t := 100
        sum := x + y
        dif := x - y
        return t[sum] - t[dif]
    }`
    testlib := []string{
        "jack/memory.jack",
        "jack/array.jack",
        "jack/list.jack",
    }
    filenames := []string{"main.jack", "mult.jack"}
    contents := []string{main, mult}
    for _, lib := range testlib {
        data, err := os.ReadFile(lib)
        if err != nil {
            t.Fatal(err)
        }
        filenames = append(filenames, lib)
        contents = append(contents, string(data))
    }
    program, err := compile(filenames, contents)
    if err != nil {
        t.Fatal(err)
    }
    cpu := NewBarrelShiftCPU()
    computer := NewComputer(cpu)
    computer.LoadProgram(NewROM32K(program))
    w := &hexCapture{}
    computer.data_mem.writer.LoadOutputWriter(w)
    td := &testDebugger{t:t, w:w}
    run(computer, td)
    if len(w.out) != 1 {
        t.Fatalf("expected w.out of len 1, but got %v", w.out)
    }
    if w.out[0] != 1512 {
        t.Errorf("expected %04x but got %04x", 1512, w.out[0])
    }
    t.Fatal(td.steps) // 4526 including setting up the lookup table!
}
