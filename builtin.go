package main

// these are the more optimised versions of the chips defined earlier
// where we abstract a little more over the clock

type Builtin interface {
    ClockTick()
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
        rams[1].mem[i] = instr
    }
    return &ROM32K{
        rams: rams,
    }
}

func splitaddr(address uint16) (bit1 int, last14 uint16) {
    // split b.address into bit0, bit1, last 14bits
    if address >= 32768 { // 2**15
        address -= 32768
    }
    if address >= 16384 { // 2**14 
        bit1 = 1 // otherwise it remains 0
        address -= 16384
    }
    last14 = address
    return
}

func (b *ROM32K) SendAddress(addr uint16) {
    b.address = addr
    _, addrs := splitaddr(b.address)
    b.rams[0].SendAddress(addrs)
    b.rams[1].SendAddress(addrs)
}

func (b *ROM32K) Out() uint16 {
    bit1, _ := splitaddr(b.address)
    return b.rams[bit1].Out()
}

func (b *ROM32K) ClockTick() {
    b.rams[0].ClockTick()
    b.rams[1].ClockTick()
}

// components: registers A and D, ALU, counter
// instr decoding, instr execution, next instr fetching
type CPU struct {
    // inputs
    inM uint16
    instr uint16
    //reset bool
    // outputs
    //outM uint16
    //writeM bool
    //addressM uint16 // 15 bit in spec
    //pc uint16 // 15 bit in spec
    // components
    a *BuiltinRegister
    d *BuiltinRegister
    pc *BuiltinCounter

}

func NewCPU() *CPU {
    return &CPU{
        a: NewBuiltinRegister(),
        d: NewBuiltinRegister(),
        pc: NewBuiltinCounter(),
    }
}

func (cpu *CPU) decode() (bit, [7]bit, [3]bit, [3]bit) {
    b := toBit16(cpu.instr)
    isC := b[0] // isA if false
    comp := [7]bit{b[3], b[4], b[5], b[6], b[7], b[8], b[9]}
    dest := [3]bit{b[10], b[11], b[12]}
    jump := [3]bit{b[13], b[14], b[15]}
    return isC, comp, dest, jump
}

func (b *CPU) evalALU() (out uint16, zr, ng bit) {
    x := toBit16(b.d.Out())
    _, c, _, _ := b.decode()
    y := Mux16(toBit16(b.a.Out()), toBit16(b.inM), c[0])
    var o [16]bit
    o, zr, ng = Alu(x, y, c[1], c[2], c[3], c[4], c[5], c[6])
    out = fromBit16(o)
    return out, zr, ng
}

func (b *CPU) SendInM(in uint16) {
    b.inM = in
}

func (b *CPU) SendInstr(instr uint16) {
    b.instr = instr
}

func (b *CPU) SendReset(reset bool) {
    b.pc.SendReset(reset)
}

func (b *CPU) OutM() uint16 {
    outM, _, _ := b.evalALU()
    return outM
}

func (b *CPU) WriteM() bool {
    isC, _, dest, _ := b.decode()
    return bool(And(isC, dest[2]))
}

func (b *CPU) AddressM() uint16 {
    return b.a.Out()
}

func (b *CPU) PC() uint16 {
    return b.pc.Out()
}

func (b *CPU) ClockTick() {
    outM, isZero, isNeg := b.evalALU()
    isPos := Not(isNeg)
    outA := b.a.Out()
    isC, _, dest, jump := b.decode()

    // true if out<0 check matches j1 jump flag
    j1 := Not(Xor(jump[0], isNeg))
    // true if out=0 check matches j2 jump flag
    j2 := Not(Xor(jump[1], isZero))
    // true if out>0 check matches j3 jump flag
    j3 := Not(Xor(jump[2], isPos))
    // j1 and j3 need to match OR j2 matches and j2 = true
    // meaning: either we should be zero and are zero, or match pos/neg correctly
    // one of the three above needs to match
    shouldJump := bool(And(isC, Or(And(j1, j3), And(j2, jump[1]))))

    b.pc.SendIn(outA)
    // we either jump or incr pc, never both
    b.pc.SendInc(!shouldJump)
    b.pc.SendLoad(shouldJump)
    b.pc.ClockTick()
    // tick PC before A since it depends on it
    b.a.SendIn(fromBit16(Mux16(toBit16(b.instr), toBit16(outM), isC)))
    b.a.SendLoad(bool(Or(Not(isC), And(isC, dest[0]))))
    b.d.SendIn(outM)
    b.d.SendLoad(bool(And(isC, dest[1])))
    b.a.ClockTick()
    b.d.ClockTick()
}

type Memory struct {
    address uint16 // 15 bit in spec

    ram *BuiltinRAM16K
    screen Screen
    keyboard Keyboard
}

func NewMemory() *Memory {
    return &Memory{
        ram: NewBuiltinRAM16K(),
        screen: NewScreen256x512(),
        keyboard: NewSimpleKeyboard(),
    }
}

func (m *Memory) SendIn(in uint16) {
    m.ram.SendIn(in)
    m.screen.SendIn(in)
}

func (m *Memory) SendLoad(load bool) {
    m.ram.SendLoad(load)
    m.screen.SendLoad(load)
}

func (m *Memory) SendAddress(address uint16) {
    m.address = address
    _, addr := splitaddr(address)
    m.ram.SendAddress(addr)
    m.screen.SendAddress(addr)
}

func (m *Memory) Out() uint16 {
    bit1, address := splitaddr(m.address)
    if bit1 == 0 {
        return m.ram.Out()
    }
    if address >= 8192 { // 2**13 
        if address > 0x6000 {
            panic("access memory beyond 0x6000")
        }
        return m.keyboard.Out()
    }
    return m.screen.Out()
}

func (m *Memory) ClockTick() {
    m.ram.ClockTick()
    // TODO: should screen/keyboard even be clocked?
    m.screen.ClockTick()
    m.keyboard.ClockTick()
}

type Computer struct {
    cpu *CPU
    instr_mem *ROM32K
    data_mem *Memory
}

func NewComputer() *Computer {
    datamem := NewMemory()
    return &Computer{
        cpu: NewCPU(),
        instr_mem: NewROM32K(nil),
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
    c.cpu.SendInstr(c.instr_mem.Out())
    c.cpu.ClockTick()
    c.data_mem.SendLoad(c.cpu.WriteM())
    c.data_mem.SendIn(c.cpu.OutM())
    c.data_mem.SendAddress(c.cpu.AddressM())
    c.data_mem.ClockTick()
    c.cpu.SendInM(c.data_mem.Out())
    c.instr_mem.SendAddress(c.cpu.PC())
    c.instr_mem.ClockTick()
}

func (c *Computer) LoadProgram(rom *ROM32K) {
    c.instr_mem = rom
}
