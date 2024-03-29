package main

// these are the more optimised versions of the chips defined earlier
// where we abstract a little more over the clock

type Builtin interface {
    ClockTick()
}

type BuiltinBit struct {
    in bool
    out bool
    load bool
}

func NewBuiltinBit() *BuiltinBit {
    return &BuiltinBit{}
}

func (b *BuiltinBit) SendIn(in bool) {
    b.in = in
}

func (b *BuiltinBit) SendLoad(load bool) {
    b.load = load
}

func (b *BuiltinBit) Out() bool {
    return b.out
}

func (b *BuiltinBit) ClockTick() {
    if b.load {
        b.out = b.in
    }
}

type BuiltinRegister struct {
    in uint16
    out uint16
    load bool
}

func NewBuiltinRegister() *BuiltinRegister {
    return &BuiltinRegister{}
}

func (b *BuiltinRegister) SendIn(in uint16) {
    b.in = in
}

func (b *BuiltinRegister) SendLoad(load bool) {
    b.load = load
}

func (b *BuiltinRegister) Out() uint16 {
    return b.out
}

func (b *BuiltinRegister) ClockTick() {
    if b.load {
        b.out = b.in
    }
}

type BuiltinRAM16K struct {
    in uint16
    out uint16
    address uint16 // 14 bits in spec
    load bool
    mem [16384]uint16
}

func NewBuiltinRAM16K() *BuiltinRAM16K {
    return &BuiltinRAM16K{
        mem: [16384]uint16{},
    }
}

func (b *BuiltinRAM16K) SendIn(in uint16) {
    b.in = in
}

func (b *BuiltinRAM16K) SendLoad(load bool) {
    b.load = load
}

func (b *BuiltinRAM16K) SendAddress(address uint16) {
    b.address = address
    b.out = b.mem[b.address]
}

func (b *BuiltinRAM16K) Out() uint16 {
    return b.out
}

func (b *BuiltinRAM16K) ClockTick() {
    if b.load {
        b.mem[b.address] = b.in
    }
    b.out = b.mem[b.address]
}

type BuiltinCounter struct {
    in uint16
    out uint16
    inc bool
    load bool
    reset bool
}

func NewBuiltinCounter() *BuiltinCounter {
    return &BuiltinCounter{}
}

func (b *BuiltinCounter) SendIn(in uint16) {
    b.in = in
}

func (b *BuiltinCounter) SendInc(inc bool) {
    b.inc = inc
}

func (b *BuiltinCounter) SendLoad(load bool) {
    b.load = load
}

func (b *BuiltinCounter) SendReset(reset bool) {
    b.reset = reset
}

func (b *BuiltinCounter) Out() uint16 {
    return b.out
}

func (b *BuiltinCounter) ClockTick() {
    switch {
    case b.reset:
        b.out = 0
    case b.load:
        b.out = b.in
    case b.inc:
        b.out = b.out + 1
    }
}

type ROM32K struct {
    address uint16 // 15 bit in spec
    // out uint16 // from underlying ram16
    rams [2]*BuiltinRAM16K
}

func NewROM32K(program []uint16) *ROM32K {
    if len(program) > 32768 {
        panic("ROM32K OOM on load")
    }
    rams := [2]*BuiltinRAM16K{
            NewBuiltinRAM16K(), NewBuiltinRAM16K(),
    }
    for i, instr := range program {
        if i < 16384 {
            rams[0].mem[i] = instr
            continue
        }
        rams[1].mem[i-16384] = instr
    }
    return &ROM32K{
        rams: rams,
    }
}

func (b *ROM32K) SendAddress(addr uint16) {
    b.address = addr
    last14 := addr & 0x3fff
    b.rams[0].SendAddress(last14)
    b.rams[1].SendAddress(last14)
}

func (b *ROM32K) Out() uint16 {
    bit1 := 0
    if b.address & 0x4000 > 0 { bit1 = 1 }
    return b.rams[bit1].Out()
}

func (b *ROM32K) ClockTick() {
    b.rams[0].ClockTick()
    b.rams[1].ClockTick()
}

type Memory struct {
    address uint16 // 15 bit in spec

    ram *BuiltinRAM16K
    screen Screen
    keyboard Keyboard
    reader TapeReader
    writer TapeWriter
}

func NewMemory() *Memory {
    return &Memory{
        ram: NewBuiltinRAM16K(),
        screen: NewScreen512x256(),
        keyboard: NewSimpleKeyboard(),
        reader: NewTapeReader(),
        writer: NewTapeWriter(),
    }
}

func (m *Memory) SendIn(in uint16) {
    m.ram.SendIn(in)
    m.screen.SendIn(in)
    m.writer.SendIn(in)
}

// TODO: is there a bug here where load remains high on screen
// when ram.SendLoad is called?
func (m *Memory) SendLoad(load bool) {
    if m.address & 0x4000 == 0 {    // checking bit1
        m.ram.SendLoad(load)
        return
    }
    // NOTE first two bits have been masked to 0 here already
    // ALSO NOTE bit0 is ignored so 2**15+1 is mapped to MEM[1]
    last14 := m.address & 0x3fff
    if last14 < 0x2000 { // 2**13 or 0x2000
        m.screen.SendLoad(load)
        return
    }
    switch last14 {
    case 8192:
        return // load to keyboard is ignored
    case 8193:
        m.reader.SendLoad(load)
    case 8194:
        m.writer.SendLoad(load)
    default:
        if load {
            panic("access memory beyond 0x6002")
        }
    }
}

func (m *Memory) SendAddress(address uint16) {
    m.address = address
    last14 := address & 0x3fff
    m.ram.SendAddress(last14)
    m.screen.SendAddress(last14)
}

func (m *Memory) Out() uint16 {
    if m.address & 0x4000 == 0 {      // check bit1
        return m.ram.Out()
    }
    address := m.address & 0x3fff
    if address < 0x2000 { // 2**13 or 0x2000
        return m.screen.Out()
    }
    switch address {
    case 8192:
        return m.keyboard.Out()
    case 8193:
        return m.reader.Out()
    case 8194:
        return m.writer.Out()
    default:
        return 0 // access beyond 0x6002 should never find anything
        // an actual read here should explode but we always read even when setting A=0x6003+
    }
}

func (m *Memory) ClockTick() {
    m.ram.ClockTick()
    // TODO: should screen/keyboard even be clocked?
    m.screen.ClockTick()
    m.keyboard.ClockTick()
    m.reader.ClockTick()
    m.writer.ClockTick()
}

type Computer struct {
    cpu CPU
    instr_mem *ROM32K
    data_mem *Memory
}

func NewComputer(cpu CPU) *Computer {
    datamem := NewMemory()
    return &Computer{
        cpu: cpu,
        data_mem: datamem,
    }
}

//When reset is 0, the program stored in the computer's
//ROM executes. When reset is 1, the execution of the
//program restarts. Thus, to start a program's
//execution, reset must be pushed "up" (1) and then
//"down" (0).
//From this point onward the user is at the mercy of
//the software. In particular, depending on the
//program's code, the screen may show some output and
//the user may be able to interact with the computer
//via the keyboard.
func (c *Computer) SendReset(reset bool) {
    c.cpu.SendReset(reset)
}

func (c *Computer) ClockTick() {
    if c.data_mem == nil {
        panic("no ROM loaded")
    }
    c.cpu.SendInstr(c.instr_mem.Out())
    c.data_mem.SendAddress(c.cpu.AddressM())
    c.cpu.SendInM(c.data_mem.Out())
    c.cpu.ClockTick()
    c.data_mem.SendLoad(c.cpu.WriteM())
    c.data_mem.SendIn(c.cpu.OutM())
    c.data_mem.ClockTick()
    c.instr_mem.SendAddress(c.cpu.PC())
    c.instr_mem.ClockTick()
}

func (c *Computer) LoadProgram(rom *ROM32K) {
    c.instr_mem = rom
}
