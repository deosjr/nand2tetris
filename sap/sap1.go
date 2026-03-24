package main

// start by copying some builtins over from the nand2tetris builtin definitions

// renamed Computer
// TODO: could be an interface instead, see peripherals
type SAP1 struct {
	PC   *BuiltinCounter
	A    *BuiltinRegister
	B    *BuiltinRegister
	MAR  *BuiltinRegister
	PROM *ROM16x8
	IR   *BuiltinRegister
	O    *BuiltinRegister
	// D, 8 LEDs showing output of O
	// CONtrol unit
	Ring *RingCounter6
	Halt bool
	// ALU
}

func NewSAP1() *SAP1 {
	return &SAP1{
		PC:   NewBuiltinCounter(),
		A:    NewBuiltinRegister(),
		B:    NewBuiltinRegister(),
		MAR:  NewBuiltinRegister(),
		PROM: NewROM16x8(),
		IR:   NewBuiltinRegister(),
		O:    NewBuiltinRegister(),
		Ring: &RingCounter6{},
	}
}

// When reset is 0, the program stored in the computer's
// ROM executes. When reset is 1, the execution of the
// program restarts. Thus, to start a program's
// execution, reset must be pushed "up" (1) and then
// "down" (0).
// From this point onward the user is at the mercy of
// the software. In particular, depending on the
// program's code, the screen may show some output and
// the user may be able to interact with the computer
// via the keyboard.
func (c *SAP1) SendReset(reset bool) {
	c.PC.SendReset(reset)
}

func (c *SAP1) ClockTick() {
	// TODO: at start of execution, CON sends CLR to IR and PC
	// Control Unit sends clock signal to all registers
	// CON outputs a 12-bit word, based on instruction and T state (ringcounter)
	// we implement the control world explicitly, though it could be skipped entirely
	t := c.Ring.T()
	lda, add, sub, out, hlt := c.instructionDecoder()
	c.Halt = bool(hlt)
	if c.Halt {
		return
	}

	// Fig 8.11: control matrix
	cp := t[2]
	ep := t[0]
	lm := Or(Or(t[0], And(lda, t[3])), Or(And(add, t[3]), And(sub, t[3])))
	er := Or(Or(t[1], And(lda, t[4])), Or(And(add, t[4]), And(sub, t[4])))
	li := t[1]
	ei := Or(Or(And(lda, t[3]), And(add, t[3])), And(sub, t[3]))
	la := Or(Or(And(lda, t[4]), And(add, t[5])), And(sub, t[5]))
	ea := And(out, t[3])
	su := And(sub, t[5])
	eu := Or(And(add, t[5]), And(sub, t[5]))
	lb := Or(And(add, t[4]), And(sub, t[4]))
	lo := And(out, t[3])

	alu := SAP1Alu(toBit8(c.A.Out()), toBit8(c.B.Out()), su)

	c.PROM.SendAddress(c.MAR.Out())

	// 8-bit bus W
	low := toBit8(0)
	outPC := Mux8(low, toBit8(c.PC.Out()), ep)
	outPROM := Mux8(low, toBit8(c.PROM.Out()), er)
	outIR := Mux8(low, toBit8(c.IR.Out()), ei)
	outA := Mux8(low, toBit8(c.A.Out()), ea)
	outALU := Mux8(low, alu, eu)

	// bus is an OR of outputs; architecture ensures only one writes at a time
	// NOTE: PC and IR only write 4 bits each. PC handles 4 bit numbers in a uint8
	// but IR needs to mask off the 4 MSB that end up going to CON first
	outIR[0] = false
	outIR[1] = false
	outIR[2] = false
	outIR[3] = false
	w := fromBit8(Or8(Or8(outPC, outPROM), Or8(outIR, Or8(outA, outALU))))

	// NOTE: MAR only takes 4 bits from the bus. Handled by masking outputs, see above
	c.MAR.SendIn(w)
	c.MAR.SendLoad(bool(lm))
	c.IR.SendIn(w)
	c.IR.SendLoad(bool(li))
	c.A.SendIn(w)
	c.A.SendLoad(bool(la))
	c.B.SendIn(w)
	c.B.SendLoad(bool(lb))
	c.O.SendIn(w)
	c.O.SendLoad(bool(lo))
	c.PC.SendInc(bool(cp))

	c.MAR.ClockTick()
	c.IR.ClockTick()
	c.A.ClockTick()
	c.B.ClockTick()
	c.O.ClockTick()
	c.PC.ClockTick()
	c.Ring.ClockTick()
}

func (c *SAP1) LoadProgram(program [16]uint8) {
	c.PROM.mem = program
}

// Figure 8.10
func (c *SAP1) instructionDecoder() (bit, bit, bit, bit, bit) {
	instr := toBit8(c.IR.Out())
	// Malvino counts from lowest bit
	i7, i6, i5, i4 := instr[0], instr[1], instr[2], instr[3]
	n7, n6, n5, n4 := Not(i7), Not(i6), Not(i5), Not(i4)
	lda := And(And(n7, n6), And(n5, n4))
	add := And(And(n7, n6), And(n5, i4))
	sub := And(And(n7, n6), And(i5, n4))
	out := And(And(i7, i6), And(i5, n4))
	hlt := And(And(i7, i6), And(i5, i4))
	return lda, add, sub, out, hlt
}

type RingCounter6 struct {
	t uint8 // ticks from 0 to 5 in a loop
}

func (r *RingCounter6) ClockTick() {
	r.t = (r.t + 1) % 6
}

// NOTE: ring counter runs in reverse order to the one in the book
func (r *RingCounter6) T() [6]bit {
	t := [6]bit{}
	t[r.t] = true
	return t
}
