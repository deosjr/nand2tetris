package main

type SAP2 struct {
	PC   *BuiltinCounter8
	SC   *BuiltinCounter8
	A    *BuiltinRegister12
	B    *BuiltinRegister12
	MAR  *BuiltinRegister8
	MDR  *BuiltinRegister12
	RAM  *RAM256x12
	IR   *BuiltinRegister12
	X    *BuiltinCounter12
	I    *BuiltinRegister12
	O    *BuiltinRegister12
	// CONtrol unit
	// JMS flipflop: false = normal (PC), true = subroutine (SC)
	JMSFlag *BuiltinJK
	Ring    *RingCounter6
	Halt    bool
	// ALU
}

func NewSAP2() *SAP2 {
	return &SAP2{
		PC:   NewBuiltinCounter8(),
		SC:   NewBuiltinCounter8(),
		A:    NewBuiltinRegister12(),
		B:    NewBuiltinRegister12(),
		MAR:  NewBuiltinRegister8(),
		MDR:  NewBuiltinRegister12(),
		RAM:  NewRAM256x12(),
		IR:   NewBuiltinRegister12(),
		X:    NewBuiltinCounter12(),
		I:    NewBuiltinRegister12(),
		O:    NewBuiltinRegister12(),
		JMSFlag: NewBuiltinJK(),
		Ring: &RingCounter6{},
	}
}

func (c *SAP2) SendReset(reset bool) {
	c.PC.SendReset(reset)
}

func (c *SAP2) Halted() bool {
	return c.Halt
}

func (c *SAP2) ClockTick() {
	t := c.Ring.T()
	
	// Figures 9.13 & 9.23: Instruction Decoder
	instr := toBit16(c.IR.Out() << 4) // 12 bit value!
	op0, op1, op2, op3 := instr[0], instr[1], instr[2], instr[3]
	n0, n1, n2, n3 := Not(op0), Not(op1), Not(op2), Not(op3)
	lda := And(And(n0, n1), And(n2, n3))
	add := And(And(n0, n1), And(n2, op3))
	sub := And(And(n0, n1), And(op2, n3))
	sta := And(And(n0, n1), And(op2, op3))
	ldb := And(And(n0, op1), And(n2, n3))
	ldx := And(And(n0, op1), And(n2, op3))
	jmp := And(And(n0, op1), And(op2, n3))
	jam := And(And(n0, op1), And(op2, op3))
	jaz := And(And(op0, n1), And(n2, n3))
	jim := And(And(op0, n1), And(n2, op3))
	jiz := And(And(op0, n1), And(op2, n3))
	jms := And(And(op0, n1), And(op2, op3))
	// 1100, 1101 and 1110 unused
	opr := And(And(op0, op1), And(op2, op3))

	x0, x1, x2, x3 := instr[4], instr[5], instr[6], instr[7]
	n0, n1, n2, n3 = Not(x0), Not(x1), Not(x2), Not(x3)
	//nop := And(opr, And(And(n0, n1), And(n2, n3)))
	cla := And(opr, And(And(n0, n1), And(n2, x3)))
	xch := And(opr, And(And(n0, n1), And(x2, n3)))
	dex := And(opr, And(And(n0, n1), And(x2, x3)))
	inx := And(opr, And(And(n0, x1), And(n2, n3)))
	cma := And(opr, And(And(n0, x1), And(n2, x3)))
	cmb := And(opr, And(And(n0, x1), And(x2, n3)))
	ior := And(opr, And(And(n0, x1), And(x2, x3)))
	and := And(opr, And(And(x0, n1), And(n2, n3)))
	nor := And(opr, And(And(x0, n1), And(n2, x3)))
	nan := And(opr, And(And(x0, n1), And(x2, n3)))
	xor := And(opr, And(And(x0, n1), And(x2, x3)))
	brb := And(opr, And(And(x0, x1), And(n2, n3)))
	inp := And(opr, And(And(x0, x1), And(n2, x3)))
	out := And(opr, And(And(x0, x1), And(x2, n3)))
	hlt := And(opr, And(And(x0, x1), And(x2, x3)))

	c.Halt = bool(hlt)
	if c.Halt {
		return
	}

	// Figures 9.14 & 9.15
	a := toBit12(c.A.Out())
	aZero := Not(Or12Way(a))
	aMinus := a[0]

	x := toBit12(c.X.Out())
	xZero := Not(Or12Way(x))
	xMinus := x[0]

	// Figure 9.22: SAP-2 control unit

	// Figure 9.16: jump logic
	jOR := Or(And(jmp, t[3]),
		Or(And(jam, And(aMinus, t[3])),
			Or(And(jaz, And(aZero, t[3])),
				Or(And(jim, And(xMinus, t[3])),
					And(jiz, And(xZero, t[3])),
			))))

	// Figure 9.17: JMS flag
	c.JMSFlag.SendJ(And(jms, t[3]))
	c.JMSFlag.SendK(And(brb, t[3]))


	// T0-2: as SAP-1. ME replaces ER. Primes are JMS toggled
	epPrime := t[0]
	lm0 := t[0]
	li := t[1]
	cpPrime := t[2]

	// ram control: ME is normal (MAR addressed) out, WE+ME is write MDR _into_ RAM
	me0 := t[1]

	// Figure 9.18: jump and subroutine control
	q := c.JMSFlag.OutQ()
	qBar := c.JMSFlag.OutQBar()
	jOR = Or(And(jms, t[4]), jOR)

	ep := And(qBar, epPrime)
	es := And(q, epPrime)
	cp := And(qBar, cpPrime)
	cs := And(q, cpPrime)
	lp := And(qBar, jOR)
	ls := And(q, jOR)

	// Figures 9.19, 9.20 & 9.21: rest of control matrix
	// outputs per instruction, then OR-joined at the end
	// LDA:
	ei1 := And(lda, t[3])
	lm1 := And(lda, t[3])
	me1 := And(lda, t[4])
	la1 := And(lda, t[4])
	// ADD:
	ei2 := And(add, t[3])
	lm2 := And(add, t[3])
	me2 := And(add, t[4])
	lb1 := And(add, t[4])
	eu1 := And(add, t[5])
	la2 := And(add, t[5])
	s01 := And(add, t[5])
	// SUB:
	ei3 := And(sub, t[3])
	lm3 := And(sub, t[3])
	me3 := And(sub, t[4])
	lb2 := And(sub, t[4])
	eu2 := And(sub, t[5])
	la3 := And(sub, t[5])
	s11 := And(sub, t[5])
	// STA:
	ei4 := And(sta, t[3])
	lm4 := And(sta, t[3])
	ea1 := And(sta, t[4])
	ld1 := And(sta, t[4])
	we := And(sta, t[5])
	me4 := And(sta, t[5])
	// LDB:
	ei5 := And(ldb, t[3])
	lm5 := And(ldb, t[3])
	me5 := And(ldb, t[4])
	lb3 := And(ldb, t[4])
	// LDX:
	ei6 := And(ldx, t[3])
	lm6 := And(ldx, t[3])
	me6 := And(ldx, t[4])
	lx1 := And(ldx, t[4])
	// jumps: lp/ls already covered in 9.18
	// JMP:
	ei7 := And(jmp, t[3])
	// JAM:
	ei8 := And(jam, t[3])
	// JAZ:
	ei9 := And(jaz, t[3])
	// JIM:
	ei10 := And(jim, t[3])
	// JIZ:
	ei11 := And(jiz, t[3])
	// JMS:
	ei12 := And(jms, t[4])
	// NOP
	// CLA:
	eu3 := And(cla, t[3])
	la4 := And(cla, t[3])
	// XCH:
	ea2 := And(xch, t[3])
	ld2 := And(xch, t[3])
	ex := And(xch, t[4])
	la5 := And(xch, t[4])
	ed := And(xch, t[5])
	lx2 := And(xch, t[5])
	// DEX
	// INX
	// CMA:
	eu4 := And(cma, t[3])
	la6 := And(cma, t[3])
	s12 := And(cma, t[3])
	s02 := And(cma, t[3])
	// CMB:
	eu5 := And(cmb, t[3])
	lb4 := And(cmb, t[3])
	s21 := And(cmb, t[3])
	// IOR:
	eu6 := And(ior, t[3])
	la7 := And(ior, t[3])
	s22 := And(ior, t[3])
	s03 := And(ior, t[3])
	// AND:
	eu7 := And(and, t[3])
	la8 := And(and, t[3])
	s23 := And(and, t[3])
	s13 := And(and, t[3])
	// NOR:
	eu8 := And(nor, t[3])
	la9 := And(nor, t[3])
	s24 := And(nor, t[3])
	s14 := And(nor, t[3])
	s04 := And(nor, t[3])
	// NAN:
	eu9 := And(nan, t[3])
	la10 := And(nan, t[3])
	s31 := And(nan, t[3])
	// XOR:
	eu10 := And(xor, t[3])
	la11 := And(xor, t[3])
	s32 := And(xor, t[3])
	s05 := And(xor, t[3])
	// BRB
	// INP:
	ln := And(inp, t[3])
	en := And(inp, t[3])
	la12 := And(inp, t[4])
	// OUT:
	ea3 := And(out, t[3])
	lo := And(out, t[3])

	orJoin := func(bits ...bit) bit {
		var out bit
		for _, b := range bits {
			out = Or(out, b)
		}
		return out
	}

	ei := orJoin(ei1, ei2, ei3, ei4, ei5, ei6, ei7, ei8, ei9, ei10, ei11, ei12)
	lm := orJoin(lm0, lm1, lm2, lm3, lm4, lm5, lm6)
	me := orJoin(me0, me1, me2, me3, me4, me5, me6)
	la := orJoin(la1, la2, la3, la4, la5, la6, la7, la8, la9, la10, la11, la12)
	lb := orJoin(lb1, lb2, lb3, lb4)
	eu := orJoin(eu1, eu2, eu3, eu4, eu5, eu6, eu7, eu8, eu9, eu10)
	s0 := orJoin(s01, s02, s03, s04, s05)
	s1 := orJoin(s11, s12, s13, s14)
	s2 := orJoin(s21, s22, s23, s24)
	s3 := orJoin(s31, s32)
	ld := orJoin(ld1, ld2)
	lx := orJoin(lx1, lx2)
	ea := orJoin(ea1, ea2, ea3)

	alu := SAP2Alu(toBit12(c.A.Out()), toBit12(c.B.Out()), s3, s2, s1, s0)

	c.RAM.SendAddress(c.MAR.Out())

	// 12-bit bus W
	low := toBit12(0)
	outSC := Mux12(low, toBit12(uint16(c.SC.Out())), es)
	outPC := Mux12(low, toBit12(uint16(c.PC.Out())), ep)
	outRAM := Mux12(low, toBit12(c.RAM.Out()), me)
	outMDR := Mux12(low, toBit12(c.MDR.Out()), ed)
	outIR := Mux12(low, toBit12(c.IR.Out()), ei)
	outI := Mux12(low, toBit12(c.I.Out()), en)
	outA := Mux12(low, toBit12(c.A.Out()), ea)
	outALU := Mux12(low, alu, eu)
	outX := Mux12(low, toBit12(c.X.Out()), ex)

	outIR = toBit12(fromBit12(outIR) & 0xFFF)
	w := fromBit12(Or12(outSC, Or12(outPC, Or12(outRAM, Or12(outMDR, Or12(outIR, Or12(outI, Or12(outA, Or12(outALU, outX)))))))))
	wLSB := uint8(w & 0xFF)

	// Clock everything at the end
	c.SC.SendIn(wLSB)
	c.SC.SendInc(bool(cs))
	c.SC.SendLoad(bool(ls))
	c.PC.SendIn(wLSB)
	c.PC.SendInc(bool(cp))
	c.PC.SendLoad(bool(lp))
	c.MAR.SendIn(wLSB)
	c.MAR.SendLoad(bool(lm))
	c.RAM.SendLoad(bool(And(we, me)))
	c.RAM.SendIn(c.MDR.Out())
	c.MDR.SendIn(w)
	c.MDR.SendLoad(bool(ld))
	c.IR.SendIn(w)
	c.IR.SendLoad(bool(li))
	// c.I.SendIn(fromCircuitInterface??)
	c.I.SendLoad(bool(ln))
	c.A.SendIn(w)
	c.A.SendLoad(bool(la))
	c.B.SendIn(w)
	c.B.SendLoad(bool(lb))
	c.X.SendIn(w)
	c.X.SendLoad(bool(lx))
	c.X.SendIncr(bool(And(inx, t[3])))
	c.X.SendDecr(bool(And(dex, t[3])))
	c.O.SendIn(w)
	c.O.SendLoad(bool(lo))

	c.SC.ClockTick()
	c.PC.ClockTick()
	c.MAR.ClockTick()
	c.RAM.ClockTick()
	c.MDR.ClockTick()
	c.IR.ClockTick()
	c.I.ClockTick()
	c.A.ClockTick()
	c.B.ClockTick()
	c.X.ClockTick()
	c.O.ClockTick()
	c.JMSFlag.ClockTick()
	c.Ring.ClockTick()
}

func (c *SAP2) LoadProgram(program [256]uint16) {
	c.RAM.mem = program
}