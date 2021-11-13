package main

type CPU interface {
    SendInM(uint16)
    SendInstr(uint16)
    SendReset(bool)
    OutM() uint16
    WriteM() bool
    AddressM() uint16
    PC() uint16
    ClockTick()
}

// components: registers A and D, ALU, counter
// instr decoding, instr execution, next instr fetching
type BuiltinCPU struct {
    inM uint16
    instr uint16
    a *BuiltinRegister
    d *BuiltinRegister
    pc *BuiltinCounter

}

func NewBuiltinCPU() *BuiltinCPU {
    return &BuiltinCPU{
        a: NewBuiltinRegister(),
        d: NewBuiltinRegister(),
        pc: NewBuiltinCounter(),
    }
}

func (cpu *BuiltinCPU) decode() (bit, [7]bit, [3]bit, [3]bit) {
    b := toBit16(cpu.instr)
    isC := b[0] // isA if false
    comp := [7]bit{b[3], b[4], b[5], b[6], b[7], b[8], b[9]}
    dest := [3]bit{b[10], b[11], b[12]}
    jump := [3]bit{b[13], b[14], b[15]}
    return isC, comp, dest, jump
}

func (b *BuiltinCPU) evalALU() (out uint16, zr, ng bit) {
    x := toBit16(b.d.Out())
    _, c, _, _ := b.decode()
    y := Mux16(toBit16(b.a.Out()), toBit16(b.inM), c[0])
    var o [16]bit
    o, zr, ng = Alu(x, y, c[1], c[2], c[3], c[4], c[5], c[6])
    out = fromBit16(o)
    return out, zr, ng
}

func (b *BuiltinCPU) SendInM(in uint16) {
    b.inM = in
}

func (b *BuiltinCPU) SendInstr(instr uint16) {
    b.instr = instr
}

func (b *BuiltinCPU) SendReset(reset bool) {
    b.pc.SendReset(reset)
}

func (b *BuiltinCPU) OutM() uint16 {
    outM, _, _ := b.evalALU()
    return outM
}

func (b *BuiltinCPU) WriteM() bool {
    isC, _, dest, _ := b.decode()
    return bool(And(isC, dest[2]))
}

func (b *BuiltinCPU) AddressM() uint16 {
    return b.a.Out()
}

func (b *BuiltinCPU) PC() uint16 {
    return b.pc.Out()
}

func (b *BuiltinCPU) ClockTick() {
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

// Hacking the unused instruction bits to support
// more efficient ROM reading: adds a PC register
// that can be used to jump back and forth
// We store the pc output in this register by default
// instr bit 1 set to 0 disables this update so the
// value grows stale; instr bit 2 set to 0 means
// jump to the value of the register (+1?)
type PCRegisterCPU struct {
    inM uint16
    instr uint16
    a *BuiltinRegister
    d *BuiltinRegister
    pc *BuiltinCounter
    // stores last pc, unless pcrl is false
    pcr *BuiltinRegister
    // controls whether pcr goes stale or not
    pcrl *BuiltinBit
}

func NewPCRegisterCPU() *PCRegisterCPU {
    return &PCRegisterCPU{
        a: NewBuiltinRegister(),
        d: NewBuiltinRegister(),
        pc: NewBuiltinCounter(),
        pcr: NewBuiltinRegister(),
        pcrl: NewBuiltinBit(),
    }
}

func (cpu *PCRegisterCPU) decode() (bit, bit, bit, [7]bit, [3]bit, [3]bit) {
    b := toBit16(cpu.instr)
    isC := b[0] // isA if false
    writePCR := b[1]
    jumpPCR := b[2]
    comp := [7]bit{b[3], b[4], b[5], b[6], b[7], b[8], b[9]}
    dest := [3]bit{b[10], b[11], b[12]}
    jump := [3]bit{b[13], b[14], b[15]}
    return isC, writePCR, jumpPCR, comp, dest, jump
}

func (b *PCRegisterCPU) evalALU() (out uint16, zr, ng bit) {
    _, _, jumpPCR, c, _, _ := b.decode()
    x := Mux16(toBit16(b.pcr.Out()), toBit16(b.d.Out()), jumpPCR)
    y := Mux16(toBit16(b.a.Out()), toBit16(b.inM), c[0])
    var o [16]bit
    o, zr, ng = Alu(x, y, c[1], c[2], c[3], c[4], c[5], c[6])
    out = fromBit16(o)
    return out, zr, ng
}

func (b *PCRegisterCPU) SendInM(in uint16) {
    b.inM = in
}

func (b *PCRegisterCPU) SendInstr(instr uint16) {
    b.instr = instr
}

func (b *PCRegisterCPU) SendReset(reset bool) {
    b.pc.SendReset(reset)
}

func (b *PCRegisterCPU) OutM() uint16 {
    outM, _, _ := b.evalALU()
    return outM
}

func (b *PCRegisterCPU) WriteM() bool {
    isC, _, _, _, dest, _ := b.decode()
    return bool(And(isC, dest[2]))
}

func (b *PCRegisterCPU) AddressM() uint16 {
    return b.a.Out()
}

func (b *PCRegisterCPU) PC() uint16 {
    return b.pc.Out()
}

func (b *PCRegisterCPU) ClockTick() {
    outM, isZero, isNeg := b.evalALU()
    isPos := Not(isNeg)
    outA := b.a.Out()
    isC, writePCR, jumpPCR, _, dest, jump := b.decode()
    pc := b.PC()
    pcrl := bit(b.pcrl.Out())

    // true if out<0 check matches j1 jump flag
    j1 := Not(Xor(jump[0], isNeg))
    // true if out=0 check matches j2 jump flag
    j2 := Not(Xor(jump[1], isZero))
    // true if out>0 check matches j3 jump flag
    j3 := Not(Xor(jump[2], isPos))
    // j1 and j3 need to match OR j2 matches and j2 = true
    // meaning: either we should be zero and are zero, or match pos/neg correctly
    // one of the three above needs to match
    // confusingly, jumpPCR=0 means jump
    shouldJump := bool(And(isC, Or(Not(jumpPCR), Or(And(j1, j3), And(j2, jump[1])))))

    //b.pc.SendIn(fromBit16(Mux16(toBit16(b.pcr.Out()), toBit16(outA), jumpPCR)))
    b.pc.SendIn(fromBit16(Mux16(toBit16(outM), toBit16(outA), jumpPCR)))
    // we either jump or incr pc, never both
    b.pc.SendInc(!shouldJump)
    b.pc.SendLoad(shouldJump)
    b.pc.ClockTick()
    b.pcr.SendIn(pc)
    b.pcr.SendLoad(bool(Not(pcrl)))
    b.pcr.ClockTick()
    // tick PC before A since it depends on it
    b.a.SendIn(fromBit16(Mux16(toBit16(b.instr), toBit16(outM), isC)))
    b.a.SendLoad(bool(Or(Not(isC), And(isC, dest[0]))))
    b.d.SendIn(outM)
    b.d.SendLoad(bool(And(isC, dest[1])))
    b.a.ClockTick()
    b.d.ClockTick()

    b.pcrl.SendIn(bool(Not(pcrl)))
    b.pcrl.SendLoad(bool(And(isC, Not(writePCR))))
    b.pcrl.ClockTick()
}
