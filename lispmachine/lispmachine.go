package main

// start by copying some builtins over from the nand2tetris builtin definitions

// instruction table change:
// We will have two M's: a car and a cdr of the pair
// We can read/write separately from/to those, but A is still the sole instruction register
// we will use the second bit to indicate bypassing the ALU (similar to how bitshiftCPU works)
// and we'll add lisp builtin machine instructions this way
// we will use a stack-based vm similar to the hack vm for the lisp machine
// third bit unused as of now

// As for types, we use the same as in deosjr/whistle:
// sexpression bool flags
// isExpression isAtom    isSymbol
// else Proc    else Pair else Primitive
// if Proc      isBuiltin
//              else user defined procedure
// So that's 3 bits of type info leaving 13 bits of actual data
// (or we extend the whole thing to work on more than 16-bit words)

// Example: SETCAR is a machine instruction, implemented using the CarM dest of an instr
// ISPAIR is a machine instruction that checks type data on the individual bit level
// SETCAR will use ISPAIR as part of its implementation
// The high level language is Lisp (or Scheme)

// CONS vm instruction uses SETCAR and SETCDR instrs and creates the pointer to return
// the vm level is also where 'free' lives (?!), we cant calculate pointer without it
// CAR/CDR vm level instructions
// ASSQ is a vm level instruction that takes a key and a pointer to a cons cell
// which is assumed to be an assoc list, and returns value associated with key or NIL

// renamed Computer
// TODO: could be an interface instead, see peripherals
type LispMachine struct {
    cpu *LispCPU
    instr_mem *ROM32K
    data_mem *LispMemory
}

func NewLispMachine(cpu *LispCPU) *LispMachine {
    datamem := NewLispMemory()
    return &LispMachine{
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
func (c *LispMachine) SendReset(reset bool) {
    c.cpu.SendReset(reset)
}

func (c *LispMachine) ClockTick() {
    if c.data_mem == nil {
        panic("no ROM loaded")
    }
    c.cpu.SendInstr(c.instr_mem.Out())
    c.data_mem.SendAddress(c.cpu.AddressM())
    car, cdr := c.data_mem.Out()
    c.cpu.SendInCarM(car)
    c.cpu.SendInCdrM(cdr)
    c.cpu.ClockTick()
    c.data_mem.SendLoad(c.cpu.WriteCarM(), c.cpu.WriteCdrM())
    c.data_mem.SendIn(c.cpu.OutCarM(), c.cpu.OutCdrM())
    c.data_mem.ClockTick()
    c.instr_mem.SendAddress(c.cpu.PC())
    c.instr_mem.ClockTick()
}

func (c *LispMachine) LoadProgram(rom *ROM32K) {
    c.instr_mem = rom
}

// renamed Memory
// added another RAM16K; we will store pairs
type LispMemory struct {
    address uint16 // 15 bit in spec

    ramCar *BuiltinRAM16K
    ramCdr *BuiltinRAM16K
    screen Screen
    keyboard Keyboard
    reader TapeReader
    writer TapeWriter
}

func NewLispMemory() *LispMemory {
    return &LispMemory{
        ramCar: NewBuiltinRAM16K(),
        ramCdr: NewBuiltinRAM16K(),
        screen: NewScreen512x256(),
        keyboard: NewSimpleKeyboard(),
        reader: NewTapeReader(),
        writer: NewTapeWriter(),
    }
}

func (m *LispMemory) SendIn(inCar, inCdr uint16) {
    m.ramCar.SendIn(inCar)
    m.ramCdr.SendIn(inCdr)
    m.screen.SendIn(inCar)
    m.writer.SendIn(inCar)
}

// TODO: is there a bug here where load remains high on screen
// when ram.SendLoad is called?
func (m *LispMemory) SendLoad(loadCar, loadCdr bool) {
    m.sendLoad(m.ramCar, loadCar, true)
    m.sendLoad(m.ramCdr, loadCdr, false)
}

// NOTE: peripherals are called twice atm!
func (m *LispMemory) sendLoad(ram *BuiltinRAM16K, load, sendToPeripherals bool) {
    bit1, address := splitaddr(m.address)
    if bit1 == 0 {
        ram.SendLoad(load)
        return
    }
    if !sendToPeripherals {
        return
    }
    // NOTE first two bits have been masked to 0 here already
    // ALSO NOTE bit0 is ignored so 2**15+1 is mapped to MEM[1]
    if address < 8192 { // 2**13 or 0x2000
        m.screen.SendLoad(load)
        return
    }
    switch address {
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

func (m *LispMemory) SendAddress(address uint16) {
    m.address = address
    _, addr := splitaddr(address)
    m.ramCar.SendAddress(addr)
    m.ramCdr.SendAddress(addr)
    m.screen.SendAddress(addr)
}

// we always return pairs, but only use paired 16K ram, not for peripherals.
// for now we just duplicate their output if needed
func (m *LispMemory) Out() (uint16, uint16) {
    bit1, address := splitaddr(m.address)
    if bit1 == 0 {
        return m.ramCar.Out(), m.ramCdr.Out()
    }
    if address < 8192 { // 2**13 or 0x2000
        return m.screen.Out(), m.screen.Out()
    }
    switch address {
    case 8192:
        return m.keyboard.Out(), m.keyboard.Out()
    case 8193:
        return m.reader.Out(), m.reader.Out()
    case 8194:
        return m.writer.Out(), m.writer.Out()
    default:
        return 0, 0 // access beyond 0x6002 should never find anything
        // an actual read here should explode but we always read even when setting A=0x6003+
    }
}

func (m *LispMemory) ClockTick() {
    m.ramCar.ClockTick()
    m.ramCdr.ClockTick()
    // TODO: should screen/keyboard even be clocked?
    m.screen.ClockTick()
    m.keyboard.ClockTick()
    m.reader.ClockTick()
    m.writer.ClockTick()
}

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

type LispCPU struct {
    inCarM uint16
    inCdrM uint16
    instr uint16
    outCarM uint16
    outCdrM uint16
    a *BuiltinRegister
    d *BuiltinRegister
    pc *BuiltinCounter

}

func NewLispCPU() *LispCPU {
    return &LispCPU{
        a: NewBuiltinRegister(),
        d: NewBuiltinRegister(),
        pc: NewBuiltinCounter(),
    }
}

func (cpu *LispCPU) decode() (bit, bit, bit, [7]bit, [3]bit, [3]bit) {
    b := toBit16(cpu.instr)
    isC := b[0]    // isA if false
    useAlu := b[1] // if false, bypass ALU and use lisp machine instructions
    // b[2] is still free! lets use it as 'write to mcdr' for now
    comp := [7]bit{b[3], b[4], b[5], b[6], b[7], b[8], b[9]}
    dest := [3]bit{b[10], b[11], b[12]}
    jump := [3]bit{b[13], b[14], b[15]}
    return isC, useAlu, b[2], comp, dest, jump
}

// TODO: typecheck some of this to only happen on primitives?
func (b *LispCPU) evalALU() (outCar, outCdr uint16, zr, ng bit) {
    _, useALU, _, c, _, _ := b.decode()
    x := toBit16(b.d.Out())
    y := Mux16(toBit16(b.a.Out()), toBit16(b.inCarM), c[0])
    var o [16]bit
    o, zr, ng = Alu(x, y, c[1], c[2], c[3], c[4], c[5], c[6])
    car, cdr, _ := lispALU(toBit16(b.a.Out()), toBit16(b.d.Out()), toBit16(b.inCarM), toBit16(b.inCdrM), c[0], c[1], c[2], c[3], c[4], c[5], c[6])
    // We only allow ALU operations on car, so cdr always comes from lispALU
    // This means we have to be extra careful when to set WriteCdrM or we write garbage
    outCar = fromBit16(Mux16(car, o, useALU))
    outCdr = fromBit16(cdr)
    return outCar, outCdr, zr, ng
}

// type info of a 16 bit word: ISPROC ISATOM ISSYMBOL
// if not ISPROC, then ISEXPR
// for isproc, next bit defines whether it is special(1) or not(0)
// and the one after that defines builtin(1) or userdefined(0)
// special procs do not eval their arguments before calling
// if ISEXPR bit not ISATOM then ISPAIR
// A pair is a pointer to a cons cell in memory (heap)!
// if ISPAIR then third bit defines emptylist if set
// if ISEXPR and ISATOM but not ISSYMBOL then ISPRIMITIVE
// NOTE: all types have cdr set to 0x0000 except PAIR, for pair thats an error
// without this we cant distinguish (n . 0) and n in memory!
// this func only checks the prefix of CAR
// NOTE: this means EQL checks can be fast (1 instr) on pointers
// but comparing cons cells means comparing both CAR and CDR of each
func typeInfo(x [16]bit) (isExpr, isAtom, isSymb, isProc, isSpecial, isBuiltin, isPair, isEmptyList, isPrim bit) {
    isExpr = Not(x[0])
    isAtom = And(isExpr, x[1])
    isSymb = And(isAtom, x[2])
    isProc = x[0]
    isSpecial = And(isProc, x[1])
    isBuiltin = And(isProc, x[2])
    isPair = And(isExpr, Not(x[1]))
    isEmptyList = And(isPair, x[2])
    isPrim = And(isAtom, Not(x[2]))
    return
}

// TODO: g introduced later, weird order!
func lispALU(regA, regD, inCarM, inCdrM [16]bit, g, a, b, c, d, e, f bit) (car, cdr [16]bit, typeOK bit) {
    true16 := toBit16(0xffff)
    false16 := toBit16(0x0000)
    sameType := And(Not(Xor(regD[0], inCarM[0])), And(Not(Xor(regD[1], inCarM[1])), Not(Xor(regD[2], inCarM[2]))))
    // TODO: switch on a-f bits using gates/mux
    switch {
    // SETCAR is an alias for M=D from the normal ALU. setting M to D is implicit, no other options make sense
    // SETCDR: lispALU is only way to write to outCdrM, hence setcdr here.
    // setting both at the same time would need another register (or only write same value to both)
    // NOTE: SETCDR needs b2 to be set as well atm!
    case bool(a) && bool(b) && bool(c) && bool(d) && bool(e) && bool(f): // SETCDR
        return regD, regD, true
    // so CONS is a vm instruction that uses SETCAR/SETCDR for now
    // MCDR: set D to the cdr of register indexed by A
    // NOTE: MCDR needs dest set to D
    case !bool(a) && bool(b) && bool(c) && bool(d) && bool(e) && bool(f): // MCDR
        return inCdrM, inCdrM, true
    // ISPAIR: sets D to boolean true or boolean false based on typecheck of M
    // all of the typeinfo variants exist, so ISEXPR and ISATOM and so forth
    // all of them check type of pointers; to check type of cell in memory, more is needed
    // NOTE: only symbol/primitive/emptylist works like this, other checks are masks!
    // ie ISPROC only checks first bit
    case !bool(a) && !bool(b) && bool(c): // check type prefix of car against d/e/f
        eql := And(Not(Xor(d, inCarM[0])), And(Not(Xor(e, inCarM[1])), Not(Xor(f, inCarM[2]))))
        out := Mux16(false16, true16, eql)
        return out, out, true
    // EMPTYCDR: sets D to boolean true or false based on whether cdr of M is emptylist
    case !bool(a) && !bool(b) && !bool(c): // check type prefix of cdr against d/e/f
        eql := And(Not(Xor(d, inCdrM[0])), And(Not(Xor(e, inCdrM[1])), Not(Xor(f, inCdrM[2]))))
        out := Mux16(false16, true16, eql)
        return out, out, true
    // EQL: used to check equality of lisp types. simple nested AND, 0xffff if true otherwise 0x0000
    // split into two actual instructions: EQLA and EQLM, both write to D
    case !bool(a) && bool(b) && !bool(c) && bool(d) && bool(e) && bool(f): // EQLM
        eql := sameType
        for i:=3; i<16; i++ {
            eql = And(eql, Not(Xor(regD[i], inCarM[i])))
        }
        out := Mux16(false16, true16, eql)
        return out, out, sameType
    }
    // default: return an error
    return inCarM, inCdrM, false
}

func (b *LispCPU) SendInCarM(in uint16) {
    b.inCarM = in
}

func (b *LispCPU) SendInCdrM(in uint16) {
    b.inCdrM = in
}

func (b *LispCPU) SendInstr(instr uint16) {
    b.instr = instr
}

func (b *LispCPU) SendReset(reset bool) {
    b.pc.SendReset(reset)
}

func (b *LispCPU) OutCarM() uint16 {
    return b.outCarM
}

func (b *LispCPU) OutCdrM() uint16 {
    return b.outCdrM
}

func (b *LispCPU) WriteCarM() bool {
    isC, _, _, _, dest, _ := b.decode()
    return bool(And(isC, dest[2]))
}

func (b *LispCPU) WriteCdrM() bool {
    isC, useAlu, b2, _, _, _ := b.decode()
    return bool(And(And(isC, Not(useAlu)), b2))
}

func (b *LispCPU) AddressM() uint16 {
    return b.a.Out()
}

func (b *LispCPU) PC() uint16 {
    return b.pc.Out()
}

func (b *LispCPU) ClockTick() {
    outCarM, outCdrM, isZero, isNeg := b.evalALU()
    b.outCarM = outCarM
    b.outCdrM = outCdrM
    isPos := And(Not(isNeg), Not(isZero))
    outA := b.a.Out()
    isC, useAlu, _,  _, dest, jump := b.decode()

    // NOTE: currently if alu is bypassed, jump is not reliable and shouldnt ever be used
    jgt := And(And(Not(jump[0]), And(Not(jump[1]), jump[2])), isPos)
    jeq := And(And(Not(jump[0]), And(jump[1], Not(jump[2]))), isZero)
    jge := And(And(Not(jump[0]), And(jump[1], jump[2])), Or(isZero, isPos))
    jlt := And(And(jump[0], And(Not(jump[1]), Not(jump[2]))), isNeg)
    jne := And(And(jump[0], And(Not(jump[1]), jump[2])), Not(isZero))
    jle := And(And(jump[0], And(jump[1], Not(jump[2]))), Or(isZero, isNeg))
    jmp := And(jump[0], And(jump[1], jump[2]))
    shouldJump := bool(And(And(isC, useAlu), Or(jgt, Or(jeq, Or(jge, Or(jlt, Or(jne, Or(jle, jmp))))))))

    b.pc.SendIn(outA)
    // we either jump or incr pc, never both
    b.pc.SendInc(!shouldJump)
    b.pc.SendLoad(shouldJump)
    b.pc.ClockTick()
    // tick PC before A since it depends on it
    b.a.SendIn(fromBit16(Mux16(toBit16(b.instr), toBit16(outCarM), isC)))
    b.a.SendLoad(bool(Or(Not(isC), And(isC, dest[0]))))
    b.d.SendIn(outCarM)
    b.d.SendLoad(bool(And(isC, dest[1])))
    b.a.ClockTick()
    b.d.ClockTick()
}
