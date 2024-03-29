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
    outM uint16
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
    return b.outM
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
    b.outM = outM
    isPos := And(Not(isNeg), Not(isZero))
    outA := b.a.Out()
    isC, _, dest, jump := b.decode()

    jgt := And(And(Not(jump[0]), And(Not(jump[1]), jump[2])), isPos)
    jeq := And(And(Not(jump[0]), And(jump[1], Not(jump[2]))), isZero)
    jge := And(And(Not(jump[0]), And(jump[1], jump[2])), Or(isZero, isPos))
    jlt := And(And(jump[0], And(Not(jump[1]), Not(jump[2]))), isNeg)
    jne := And(And(jump[0], And(Not(jump[1]), jump[2])), Not(isZero))
    jle := And(And(jump[0], And(jump[1], Not(jump[2]))), Or(isZero, isNeg))
    jmp := And(jump[0], And(jump[1], jump[2]))
    shouldJump := bool(And(isC, Or(jgt, Or(jeq, Or(jge, Or(jlt, Or(jne, Or(jle, jmp))))))))

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
// value grows stale; next instruction now is read into
// A without executing and jumps back to where we came from
type PCRegisterCPU struct {
    inM uint16
    instr uint16
    outM uint16
    a *BuiltinRegister
    d *BuiltinRegister
    pc *BuiltinCounter
    // stores last pc
    pcr *BuiltinRegister
    // controls whether we jump to pc
    pcrl *BuiltinBit
    // switches off writes when pcrl reads A
    writebit *BuiltinBit
}

func NewPCRegisterCPU() *PCRegisterCPU {
    return &PCRegisterCPU{
        a: NewBuiltinRegister(),
        d: NewBuiltinRegister(),
        pc: NewBuiltinCounter(),
        pcr: NewBuiltinRegister(),
        pcrl: NewBuiltinBit(),
        writebit: NewBuiltinBit(),
    }
}

func (cpu *PCRegisterCPU) decode() (bit, bit, [7]bit, [3]bit, [3]bit) {
    b := toBit16(cpu.instr)
    isC := b[0] // isA if false
    pcrl := b[1]
    comp := [7]bit{b[3], b[4], b[5], b[6], b[7], b[8], b[9]}
    dest := [3]bit{b[10], b[11], b[12]}
    jump := [3]bit{b[13], b[14], b[15]}
    return isC, pcrl, comp, dest, jump
}

func (b *PCRegisterCPU) evalALU() (out uint16, zr, ng bit) {
    _, _, c, _, _ := b.decode()
    x := toBit16(b.d.Out())
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
    return b.outM
}

func (b *PCRegisterCPU) WriteM() bool {
    isC, _, _, dest, _ := b.decode()
    return bool(And(isC, And(Not(bit(b.writebit.Out())), dest[2])))
}

func (b *PCRegisterCPU) AddressM() uint16 {
    return b.a.Out()
}

func (b *PCRegisterCPU) PC() uint16 {
    return b.pc.Out()
}

func (b *PCRegisterCPU) ClockTick() {
    outM, isZero, isNeg := b.evalALU()
    b.outM = outM
    isPos := And(Not(isNeg), Not(isZero))
    outA := b.a.Out()
    isC, pcrl, _, dest, jump := b.decode()
    pc := b.PC()
    pcrout := b.pcr.Out()
    pcrlout := bit(b.pcrl.Out())

    jgt := And(And(Not(jump[0]), And(Not(jump[1]), jump[2])), isPos)
    jeq := And(And(Not(jump[0]), And(jump[1], Not(jump[2]))), isZero)
    jge := And(And(Not(jump[0]), And(jump[1], jump[2])), Or(isZero, isPos))
    jlt := And(And(jump[0], And(Not(jump[1]), Not(jump[2]))), isNeg)
    jne := And(And(jump[0], And(Not(jump[1]), jump[2])), Not(isZero))
    jle := And(And(jump[0], And(jump[1], Not(jump[2]))), Or(isZero, isNeg))
    jmp := And(jump[0], And(jump[1], jump[2]))
    shouldJump := bool(Or(pcrlout, And(isC, Or(jgt, Or(jeq, Or(jge, Or(jlt, Or(jne, Or(jle, jmp)))))))))

    pcrPlus1 := Add16(toBit16(pcrout), toBit16(1))
    b.pc.SendIn(fromBit16(Mux16(toBit16(outA), pcrPlus1, pcrlout)))
    // we either jump or incr pc, never both
    b.pc.SendInc(!shouldJump)
    b.pc.SendLoad(shouldJump)
    b.pc.ClockTick()
    // tick PC before A since it depends on it
    b.a.SendIn(fromBit16(Mux16(toBit16(b.instr), toBit16(outM), And(isC, Not(pcrlout)))))
    b.a.SendLoad(bool(Or(pcrlout, Or(Not(isC), And(isC, dest[0])))))
    b.d.SendIn(outM)
    b.d.SendLoad(bool(And(isC, dest[1])))
    b.a.ClockTick()
    b.d.ClockTick()

    b.pcr.SendIn(pc)
    b.pcr.SendLoad(true)
    b.pcr.ClockTick()

    // basically buffer the pcrlout one clock step
    b.writebit.SendIn(bool(pcrlout))
    b.writebit.SendLoad(true)
    b.writebit.ClockTick()

    // if instr starts with 10, we flip the pcrl bit
    // next clocktick it should automatically flip back
    b.pcrl.SendIn(bool(Not(pcrlout)))
    b.pcrl.SendLoad(bool(Or(pcrlout, And(isC, Not(pcrl)))))
    b.pcrl.ClockTick()
}

// This CPU also supports the leftshift operation
// if instr[2]==0, we ignore ALU output and use barrelshift out instead
// the first bit after instr[2] (so the a bit in the Hack spec)
// the a bit still switches A/M, the next bit toggles that or using D
// as input to shift and the next 4 bits encode how much to shift
type BarrelShiftCPU struct {
    inM uint16
    instr uint16
    outM uint16
    a *BuiltinRegister
    d *BuiltinRegister
    pc *BuiltinCounter
    // stores last pc
    pcr *BuiltinRegister
    // controls whether we jump to pc
    pcrl *BuiltinBit
    // switches off writes when pcrl reads A
    writebit *BuiltinBit
}

func NewBarrelShiftCPU() *BarrelShiftCPU {
    return &BarrelShiftCPU{
        a: NewBuiltinRegister(),
        d: NewBuiltinRegister(),
        pc: NewBuiltinCounter(),
        pcr: NewBuiltinRegister(),
        pcrl: NewBuiltinBit(),
        writebit: NewBuiltinBit(),
    }
}

func (cpu *BarrelShiftCPU) decode() (bit, bit, bit, [7]bit, [3]bit, [3]bit) {
    b := toBit16(cpu.instr)
    isC := b[0] // isA if false
    pcrl := b[1]
    shift := b[2]
    comp := [7]bit{b[3], b[4], b[5], b[6], b[7], b[8], b[9]}
    dest := [3]bit{b[10], b[11], b[12]}
    jump := [3]bit{b[13], b[14], b[15]}
    return isC, pcrl, shift, comp, dest, jump
}

// NOTE: I don't think jumps will work when using the shift instruction
// since zr/ng are still based on alu output (which will be nonsensical)
func (b *BarrelShiftCPU) evalALU() (out uint16, zr, ng bit) {
    _, _, shift, c, _, _ := b.decode()
    x := toBit16(b.d.Out())
    y := Mux16(toBit16(b.a.Out()), toBit16(b.inM), c[0])
    var o [16]bit
    o, zr, ng = Alu(x, y, c[1], c[2], c[3], c[4], c[5], c[6])
    shiftOut := Shift(Mux16(y, x, c[1]), [4]bit{c[2], c[3], c[4], c[5]})
    out = fromBit16(Mux16(shiftOut, o, shift))
    return out, zr, ng
}

func (b *BarrelShiftCPU) SendInM(in uint16) {
    b.inM = in
}

func (b *BarrelShiftCPU) SendInstr(instr uint16) {
    b.instr = instr
}

func (b *BarrelShiftCPU) SendReset(reset bool) {
    b.pc.SendReset(reset)
}

func (b *BarrelShiftCPU) OutM() uint16 {
    return b.outM
}

func (b *BarrelShiftCPU) WriteM() bool {
    isC, _, _, _, dest, _ := b.decode()
    return bool(And(isC, And(Not(bit(b.writebit.Out())), dest[2])))
}

func (b *BarrelShiftCPU) AddressM() uint16 {
    return b.a.Out()
}

func (b *BarrelShiftCPU) PC() uint16 {
    return b.pc.Out()
}

func (b *BarrelShiftCPU) ClockTick() {
    outM, isZero, isNeg := b.evalALU()
    b.outM = outM
    isPos := And(Not(isNeg), Not(isZero))
    outA := b.a.Out()
    isC, pcrl, _, _, dest, jump := b.decode()
    pc := b.PC()
    pcrout := b.pcr.Out()
    pcrlout := bit(b.pcrl.Out())

    jgt := And(And(Not(jump[0]), And(Not(jump[1]), jump[2])), isPos)
    jeq := And(And(Not(jump[0]), And(jump[1], Not(jump[2]))), isZero)
    jge := And(And(Not(jump[0]), And(jump[1], jump[2])), Or(isZero, isPos))
    jlt := And(And(jump[0], And(Not(jump[1]), Not(jump[2]))), isNeg)
    jne := And(And(jump[0], And(Not(jump[1]), jump[2])), Not(isZero))
    jle := And(And(jump[0], And(jump[1], Not(jump[2]))), Or(isZero, isNeg))
    jmp := And(jump[0], And(jump[1], jump[2]))
    shouldJump := bool(Or(pcrlout, And(isC, Or(jgt, Or(jeq, Or(jge, Or(jlt, Or(jne, Or(jle, jmp)))))))))

    pcrPlus1 := Add16(toBit16(pcrout), toBit16(1))
    b.pc.SendIn(fromBit16(Mux16(toBit16(outA), pcrPlus1, pcrlout)))
    // we either jump or incr pc, never both
    b.pc.SendInc(!shouldJump)
    b.pc.SendLoad(shouldJump)
    b.pc.ClockTick()
    // tick PC before A since it depends on it
    b.a.SendIn(fromBit16(Mux16(toBit16(b.instr), toBit16(outM), And(isC, Not(pcrlout)))))
    b.a.SendLoad(bool(Or(pcrlout, Or(Not(isC), And(isC, dest[0])))))
    b.d.SendIn(outM)
    b.d.SendLoad(bool(And(isC, dest[1])))
    b.a.ClockTick()
    b.d.ClockTick()

    b.pcr.SendIn(pc)
    b.pcr.SendLoad(true)
    b.pcr.ClockTick()

    // basically buffer the pcrlout one clock step
    b.writebit.SendIn(bool(pcrlout))
    b.writebit.SendLoad(true)
    b.writebit.ClockTick()

    // if instr starts with 10, we flip the pcrl bit
    // next clocktick it should automatically flip back
    b.pcrl.SendIn(bool(Not(pcrlout)))
    b.pcrl.SendLoad(bool(Or(pcrlout, And(isC, Not(pcrl)))))
    b.pcrl.ClockTick()
}

