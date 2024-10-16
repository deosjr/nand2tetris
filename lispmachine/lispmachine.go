package main

import "fmt"

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
    if m.address & 0x4000 == 0 {    // checking bit1
        ram.SendLoad(load)
        return
    }
    if !sendToPeripherals {
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
            panic(fmt.Sprintf("access memory beyond 0x6002: %d", m.address))
        }
    }
}

func (m *LispMemory) SendAddress(address uint16) {
    m.address = address
    last14 := address & 0x3fff
    m.ramCar.SendAddress(last14)
    m.ramCdr.SendAddress(last14)
    m.screen.SendAddress(last14)
}

// we always return pairs, but only use paired 16K ram, not for peripherals.
// for now we just duplicate their output if needed
func (m *LispMemory) Out() (uint16, uint16) {
    if m.address & 0x4000 == 0 {    // checking bit1
        return m.ramCar.Out(), m.ramCdr.Out()
    }
    last14 := m.address & 0x3fff
    if last14 < 0x2000 { // 2**13 or 0x2000
        return m.screen.Out(), m.screen.Out()
    }
    switch last14 {
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
    // if useAlu, b[2]=0 means 'this is a leftshift operation'
    // if not, b[2]=1 means 'write to cdrM'
    comp := [7]bit{b[3], b[4], b[5], b[6], b[7], b[8], b[9]}
    dest := [3]bit{b[10], b[11], b[12]}
    jump := [3]bit{b[13], b[14], b[15]}
    return isC, useAlu, b[2], comp, dest, jump
}

// TODO: typecheck some of this to only happen on primitives?
func (b *LispCPU) evalALU() (outCar, outCdr uint16, zr, ng bit) {
    _, useALU, bit2, c, _, _ := b.decode()
    /*
    x := toBit16(b.d.Out())
    y := Mux16(toBit16(b.a.Out()), toBit16(b.inCarM), c[0])
    var o [16]bit
    o, zr, ng = Alu(x, y, c[1], c[2], c[3], c[4], c[5], c[6])
    */
    var x, y, o uint16
    x = b.d.Out()
    if c[0] { y = b.inCarM } else { y = b.a.Out() }
    o, zr, ng = AluInt16(x, y, c[1], c[2], c[3], c[4], c[5], c[6])
    //shiftOut := Shift(Mux16(y, x, c[1]), [4]bit{c[2], c[3], c[4], c[5]})
    car, cdr, _ := lispALU(toBit16(b.a.Out()), toBit16(b.d.Out()), toBit16(b.inCarM), toBit16(b.inCdrM), c[0], c[1], c[2], c[3], c[4], c[5], c[6])
    // We only allow ALU operations on car, so cdr always comes from lispALU
    // This means we have to be extra careful when to set WriteCdrM or we write garbage
    //outCar = fromBit16(Mux16(car, o, useALU))
    //outCar = fromBit16(Mux16(shiftOut, outCar, bit2))
    outCar = fromBit16(car)
    if useALU {
        if bit2 {
            outCar = o
        } else {
            outCar = fromBit16(Shift(Mux16(toBit16(y), toBit16(x), c[1]), [4]bit{c[2], c[3], c[4], c[5]}))
        }
    }
    outCdr = fromBit16(cdr)
    return outCar, outCdr, zr, ng
}

// type info of a 16 bit word: ISPROC ISATOM ISSYMBOL
// if not ISPROC, then ISEXPR
// for isproc, next bit defines whether it is special(1) or not(0)
// and the one after that defines builtin(1) or userdefined(0)
// special procs do not eval their arguments before calling
// special builtin: keywords like define and lambda
// special userdefined: compiled functions
// if ISEXPR bit not ISATOM then ISPAIR
// A pair is a pointer to a cons cell in memory (heap)!
// if ISPAIR then third bit defines emptylist if set
// if ISEXPR and ISATOM but not ISSYMBOL then ISPRIMITIVE
// NOTE: all types have cdr set to 0x0000 except PAIR, for pair thats an error
// without this we cant distinguish (n . 0) and n in memory!
// NOTE: above isnt true anymore?
// looks like 0x0 is emptylist and 0x0000 - 0x3fff are all valid pairs

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

func lispALU(regA, regD, inCarM, inCdrM [16]bit, a, b, c, d, e, f, g bit) (car, cdr [16]bit, typeOK bit) {
    true16 := toBit16(0xffff)
    false16 := toBit16(0x0000)
    sameType := And(Not(Xor(regD[0], inCarM[0])), And(Not(Xor(regD[1], inCarM[1])), Not(Xor(regD[2], inCarM[2]))))
    // TODO: switch on a-g bits using gates/mux
    selector := 0
    prefix := 0
    if g { selector = 0b1 }
    if f { selector |= 0b10 }
    if e { selector |= 0b100 }
    if d { selector |= 0b1000 }
    if c { selector |= 0b10000; prefix |= 0b1 }
    if b { selector |= 0b100000; prefix |= 0b10 }
    if a { selector |= 0b1000000; prefix |= 0b100 }
    if prefix == 0b000 {    // check type prefix against e/f/g
        x := inCdrM
        if d { x = inCarM } // d bit determines whether we check car or cdr of address
    // ISSYM: sets D to boolean true or boolean false based on typecheck of M
    // all of the typeinfo variants exist, so ISEXPR and ISATOM and so forth
    // all of them check type of pointers; to check type of cell in memory, more is needed
        eql := And(Not(Xor(e, x[0])), And(Not(Xor(f, x[1])), Not(Xor(g, x[2]))))
        out := Mux16(false16, true16, eql)
        return out, out, true
    }
    switch selector {
    // TODO: perhaps SETCAR/CDR should return the address set?
    // SETCAR is an alias for M=D from the normal ALU. setting M to D is implicit, no other options make sense
    // SETCDR: lispALU is only way to write to outCdrM, hence setcdr here.
    // setting both at the same time would need another register (or only write same value to both)
    // NOTE: SETCDR needs b2 to be set as well atm!
    case 0b0111111: // SETCDR
        return regD, regD, true
    // so CONS is a vm instruction that uses SETCAR/SETCDR for now
    // TODO: rename MCDR to DCDR because it writes to D
    // ACDR: set A to the cdr of register indexed by A
    // MCDR: set D to the cdr of register indexed by A
    // NOTE: MCDR needs dest set to D
    case 0b0011111: // ACDR/MCDR
        return inCdrM, inCdrM, true
    // EQL: used to check equality of lisp types. simple nested AND, 0xffff if true otherwise 0x0000
    // split into two actual instructions: EQLA and EQLM, both write to D
    case 0b0010111: // EQLM
        eql := sameType
        for i:=3; i<16; i++ {
            eql = And(eql, Not(Xor(regD[i], inCarM[i])))
        }
        out := Mux16(false16, true16, eql)
        return out, out, sameType
    case 0b1000000: // ISPROC
        out := Mux16(false16, true16, inCarM[0])
        return out, out, true
    case 0b1010000: // ISPAIR
        out := Mux16(false16, true16, And(Not(inCarM[0]), Not(inCarM[1])))
        return out, out, true
    case 0b0110000: // ISEMPTY
        carIsEmpty := Not(Or16Way(inCarM))
        out := Mux16(false16, true16, carIsEmpty)
        return out, out, true
    case 0b0100000: // EMPTYCDR
    // EMPTYCDR: sets D to boolean true or false based on whether cdr of M is emptylist
        cdrIsEmpty := Not(Or16Way(inCdrM))
        out := Mux16(false16, true16, cdrIsEmpty)
        return out, out, true
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
    //isC, useAlu, b2, _, dest, _ := b.decode()
    //return bool(And(dest[2], Or(And(isC, useAlu), And(And(isC, Not(useAlu)), Not(b2)))))
}

// only write to M cdr from a lisp instruction and if bit 2 is set
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
    isPos := !isNeg && !isZero
    outA := b.a.Out()
    isC, useAlu, _,  _, dest, jump := b.decode()

    /*
    // NOTE: currently if alu is bypassed, jump is not reliable and shouldnt ever be used
    jgt := And(And(Not(jump[0]), And(Not(jump[1]), jump[2])), isPos)
    jeq := And(And(Not(jump[0]), And(jump[1], Not(jump[2]))), isZero)
    jge := And(And(Not(jump[0]), And(jump[1], jump[2])), Or(isZero, isPos))
    jlt := And(And(jump[0], And(Not(jump[1]), Not(jump[2]))), isNeg)
    jne := And(And(jump[0], And(Not(jump[1]), jump[2])), Not(isZero))
    jle := And(And(jump[0], And(jump[1], Not(jump[2]))), Or(isZero, isNeg))
    jmp := And(jump[0], And(jump[1], jump[2]))
    shouldJump := bool(And(And(isC, useAlu), Or(jgt, Or(jeq, Or(jge, Or(jlt, Or(jne, Or(jle, jmp))))))))
    */
    jgt := (!jump[0] && (!jump[1] && jump[2])) && isPos
    jeq := (!jump[0] && (jump[1] && !jump[2])) && isZero
    jge := (!jump[0] && (jump[1] && jump[2])) && (isZero || isPos)
    jlt := (jump[0] && (!jump[1] && !jump[2])) && isNeg
    jne := (jump[0] && (!jump[1] && jump[2])) && !isZero
    jle := (jump[0] && (jump[1] && !jump[2])) && (isZero || isNeg)
    jmp := jump[0] && jump[1] && jump[2]
    shouldJump := bool(isC && useAlu && (jgt || jeq || jge || jlt || jne || jle || jmp))

    b.pc.SendIn(outA)
    // we either jump or incr pc, never both
    b.pc.SendInc(!shouldJump)
    b.pc.SendLoad(shouldJump)
    b.pc.ClockTick()
    // tick PC before A since it depends on it
    //b.a.SendIn(fromBit16(Mux16(toBit16(b.instr), toBit16(outCarM), isC)))
    if isC {
        b.a.SendIn(outCarM)
    } else {
        b.a.SendIn(b.instr)
    }

    //b.a.SendLoad(bool(Or(Not(isC), And(isC, dest[0]))))
    b.a.SendLoad(bool(!isC || (isC && dest[0])))
    b.d.SendIn(outCarM)
    b.d.SendLoad(bool(And(isC, dest[1])))
    b.a.ClockTick()
    b.d.ClockTick()
}
