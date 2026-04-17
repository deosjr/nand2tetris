package main

type SAP3 struct {
	S   *Stack4x12
	A   *ShiftRegister16
	B   *BuiltinRegister16
	MAR *IPLCounter12 // counting part (incr/decr) ignored
	MDR *BuiltinRegister16
	RAM *RAM4096x16
	IR  *BuiltinRegister16
	X   [10]*BuiltinCounter16
	I   *BuiltinRegister16
	O   *BuiltinRegister16
	P   [16]*BuiltinRegister16
	// CONtrol unit
	Ring *RingCounter6
	Halt bool
	// ALU
}

func NewSAP3() *SAP3 {
	xRegisters := [10]*BuiltinCounter16{}
	for i := range xRegisters {
		xRegisters[i] = NewBuiltinCounter16()
	}
	pRegisters := [16]*BuiltinRegister16{}
	for i := range pRegisters {
		pRegisters[i] = NewBuiltinRegister16()
	}
	return &SAP3{
		S:    NewBuiltinStack4x12(),
		A:    NewShiftRegister16(),
		B:    NewBuiltinRegister16(),
		MAR:  NewIPLCounter12(),
		MDR:  NewBuiltinRegister16(),
		RAM:  NewRAM4096x16(),
		IR:   NewBuiltinRegister16(),
		X:    xRegisters,
		I:    NewBuiltinRegister16(),
		O:    NewBuiltinRegister16(),
		P:    pRegisters,
		Ring: &RingCounter6{},
	}
}

func (c *SAP3) SendReset(reset bool) {
	c.S.SendReset(reset)
}

func (c *SAP3) Halted() bool {
	return c.Halt
}

func (c *SAP3) ClockTick() {
	t := c.Ring.T()

	outIR := c.IR.Out()
	instr := decodeSAP3instr(outIR)

	if instr == HLT {
		c.Halt = true
		return
	}

	// Table 11-1. Microroutines
	// These are defined for a microprocessor
	nop := conWord{}
	table := map[sap3instr][3]conWord{
		// mnemonic: T3, T4, T5 phase
		LDA: {{ei: true, lm: true}, {me: true, la: true}, nop},
		ADD: {{ei: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, s0: true}},
		SUB: {{ei: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, s0: true}},
		STA: {{ei: true, lm: true}, {ea: true, ld: true}, {we: true, me: true}},
		LDB: {{ei: true, lm: true}, {me: true, lb: true}, nop},
		LDX: {{ei: true, lm: true, ipl: true}, {me: true, lx: true}, nop},
		JMP: {{ei: true, lkun: true}, nop, nop},
		JAM: {{ei: true, lkam: true}, nop, nop},
		JAZ: {{ei: true, lkaz: true}, nop, nop},
		JIM: {{ei: true, lkim: true, ipl: true}, nop, nop},
		JIZ: {{ei: true, lkiz: true, ipl: true}, nop, nop},
		JMS: {{pu: true}, {ei: true, lkun: true}, nop},
		DSZ: {{dx: true}, {ckiz: true}, nop},
		ISZ: {{ix: true}, {ckiz: true}, nop},
		SHL: {{shl: true}, nop, nop},
		SHR: {{shr: true}, nop, nop},
		RAL: {{ral: true}, nop, nop},
		RAR: {{rar: true}, nop, nop},
		LDM: {{ek: true, lm: true}, {me: true, la: true}, {ckun: true}},
		ADM: {{ek: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, ckun: true, s0: true}},
		SBM: {{ek: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, ckun: true, s1: true}},
		STM: {{ek: true, lm: true}, {ea: true, ld: true}, {we: true, me: true, ckun: true}},
		ORM: {{ek: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, ckun: true, s2: true, s0: true}},
		ANM: {{ek: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, ckun: true, s2: true, s1: true}},
		XNM: {{ek: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, ckun: true, s3: true, s1: true}},
		LDN: {{ex: true, lm: true}, {me: true, la: true}, nop},
		ADN: {{ex: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, s0: true}},
		SBN: {{ex: true, lm: true}, {me: true, lb: true}, {eu: true, la: true, s1: true}},
		STN: {{ex: true, lm: true}, {ea: true, ld: true}, {we: true, me: true}},
		JSN: {{pu: true}, {ex: true, lkun: true}, nop},
		NOP: {nop, nop, nop},
		CLA: {{eu: true, la: true}, nop, nop},
		XCH: {{ea: true, ld: true}, {ex: true, la: true}, {ed: true, lx: true}},
		DEX: {{dx: true}, nop, nop},
		INX: {{ix: true}, nop, nop},
		CMA: {{eu: true, la: true, s1: true, s0: true}, nop, nop},
		CMB: {{eu: true, lb: true, s2: true}, nop, nop},
		IOR: {{eu: true, la: true, s2: true, s0: true}, nop, nop},
		AND: {{eu: true, la: true, s2: true, s1: true}, nop, nop},
		NOR: {{eu: true, la: true, s2: true, s1: true, s0: true}, nop, nop},
		NAN: {{eu: true, la: true, s3: true}, nop, nop},
		XOR: {{eu: true, la: true, s3: true, s0: true}, nop, nop},
		BRB: {{pd: true}, nop, nop},
		INP: {{eport: true, ln: true}, {en: true, la: true}, nop},
		OUT: {{ea: true, l0: true}, {e0: true, lport: true}, nop},
		HLT: {nop, nop, {hlt: true}},
	}

	var con conWord
	switch {
	case bool(t[0]):
		con = conWord{ek: true, lm: true}
	case bool(t[1]):
		con = conWord{li: true, me: true}
	case bool(t[2]):
		con = conWord{ck: true}
	case bool(t[3]):
		con = table[instr][0]
	case bool(t[4]):
		con = table[instr][1]
	case bool(t[5]):
		con = table[instr][2]
	}

	// port / x register select bits depending on instruction
	var s uint16
	switch instr {
	case DSZ, ISZ, LDX, JIM, JIZ:
		// first 4 bits have already been sliced off by IR
		s = (outIR & 0xF00) >> 8
	case XCH, DEX, INX, INP, OUT, LDN, ADN, SBN, STN, JSN:
		// first 4 bits of remaining 12 taken by opr select bits
		s = (outIR & 0xF0) >> 4
	}
	// will default to x[0] for other instructions, but ignored by eX=false
	xOut := c.X[s].Out()

	// Figure 10-14. SAP-3 architecture
	alu := SAP3Alu(toBit16(c.A.Out()), toBit16(c.B.Out()), bit(con.s3), bit(con.s2), bit(con.s1), bit(con.s0))

	c.RAM.SendAddress(c.MAR.Out())
	c.S.SendEmit(con.ek)

	// 16-bit bus W
	var w uint16
	lowMux := func(b bool, v uint16) uint16 {
		if b {
			return v
		}
		return 0
	}
	w |= lowMux(con.en, c.I.Out())
	w |= lowMux(con.ek, c.S.Out()) // 12-bit out handled internally
	w |= lowMux(con.me, c.RAM.Out())
	w |= lowMux(con.ed, c.MDR.Out())
	w |= lowMux(con.ei, c.IR.Out()&0xFFF) // 12-bit out
	w |= lowMux(con.ea, c.A.Out())
	w |= lowMux(con.eu, alu)
	w |= lowMux(con.ex, xOut)

	wLSB := w & 0xFFF

	// todo: 16-bit I/O bus

	// jump logic: conditional count/load
	a := toBit12(c.A.Out())
	x := toBit12(xOut)
	var lk bool
	ck := con.ck
	switch {
	case con.lkun:
		lk = true // load unconditionally
	case con.lkam:
		lk = bool(a[0]) // load if accumulator minus
	case con.lkaz:
		lk = bool(Not(Or12Way(a))) // load if accumulator zero
	case con.lkim:
		lk = bool(x[0]) // load if index minus
	case con.lkiz:
		lk = bool(Not(Or12Way(x))) // load if index zero
	case con.ckun:
		ck = true // count unconditional
	case con.ckiz:
		ck = bool(Not(Or12Way(x))) // count if index zero
	}

	// send inputs, tick clock
	//c.I.SendIn(ioBus)
	c.I.SendLoad(bool(con.ln))
	c.S.SendIn(wLSB)
	c.S.SendPU(bit(con.pu))
	c.S.SendPD(bit(con.pd))
	c.S.SendInc(ck)
	c.S.SendLoad(lk)
	c.S.SendIPL(con.ipl)
	c.MAR.SendIn(w)
	c.MAR.SendLoad(con.lm)
	c.MAR.SendIPL(con.ipl)
	c.RAM.SendIn(c.MDR.Out())
	c.RAM.SendLoad(con.we && con.me)
	c.MDR.SendIn(w)
	c.MDR.SendLoad(con.ld)
	c.IR.SendIn(w)
	c.IR.SendLoad(con.li)
	c.O.SendIn(w)
	c.O.SendLoad(con.lport)
	c.A.SendIn(w)
	c.A.SendLoad(con.la)
	c.A.SendSHL(con.shl)
	c.A.SendSHR(con.shr)
	c.A.SendRAL(con.ral)
	c.A.SendRAR(con.rar)
	c.B.SendIn(w)
	c.B.SendLoad(con.lb)

	// x register inputs
	for i, x := range c.X {
		x.SendIn(w)
		if i != int(s) {
			x.SendLoad(false)
			x.SendIncr(false)
			x.SendDecr(false)
			continue
		}
		x.SendLoad(con.lx)
		x.SendIncr(con.ix)
		x.SendDecr(con.dx)
	}

	// todo: port registers

	c.S.ClockTick()
	c.MAR.ClockTick()
	c.RAM.ClockTick()
	c.MDR.ClockTick()
	c.IR.ClockTick()
	c.I.ClockTick()
	c.A.ClockTick()
	c.B.ClockTick()
	c.O.ClockTick()
	for _, x := range c.X {
		x.ClockTick()
	}
	for _, p := range c.P {
		p.ClockTick()
	}
	c.Ring.ClockTick()
}

func (c *SAP3) LoadProgram(program [4096]uint16) {
	c.RAM.mem = program
}

type conWord struct {
	// SAP3 has 86 bit CON output, with x/port arguments x10
	// everything defaults to false, only true needs to be set
	e0, ea, ed, ei, ek, en, eu, ex bool
	l0, la, lb, ld, lm, ln, lx, li bool
	me, we                         bool
	s0, s1, s2, s3                 bool
	ck, ipl                        bool
	pu, pd                         bool
	lkun, lkam, lkaz, lkim, lkiz   bool
	ckun, ckiz                     bool
	dx, ix                         bool
	shl, shr, ral, rar             bool
	eport, lport                   bool
	hlt                            bool
}

type sap3instr uint8

const (
	LDA sap3instr = iota
	ADD
	SUB
	STA
	LDB
	LDX
	JMP
	JAM
	JAZ
	JIM
	JIZ
	JMS
	DSZ
	ISZ
	SHL
	SHR
	RAL
	RAR
	LDM
	ADM
	SBM
	STM
	ORM
	ANM
	XNM
	LDN
	ADN
	SBN
	STN
	JSN
	NOP
	CLA
	XCH
	DEX
	INX
	CMA
	CMB
	IOR
	AND
	NOR
	NAN
	XOR
	BRB
	INP
	OUT
	HLT
)

// could be wired like the SAP2 instruction decoder code, but this is easier
func decodeSAP3instr(ir uint16) sap3instr {
	switch ir >> 12 & 0x000F {
	case 0b0000:
		return LDA
	case 0b0001:
		return ADD
	case 0b0010:
		return SUB
	case 0b0011:
		return STA
	case 0b0100:
		return LDB
	case 0b0101:
		return LDX
	case 0b0110:
		return JMP
	case 0b0111:
		return JAM
	case 0b1000:
		return JAZ
	case 0b1001:
		return JIM
	case 0b1010:
		return JIZ
	case 0b1011:
		return JMS
	case 0b1100:
		return DSZ
	case 0b1101:
		return ISZ
	case 0b1110:
		return decodeSAP3Mix(ir >> 8 & 0x000F)
	case 0b1111:
		return decodeSAP3Opr(ir >> 8 & 0x000F)
	}
	panic("unreachable code")
}

func decodeSAP3Mix(selectcode uint16) sap3instr {
	switch selectcode {
	case 0b0000:
		return SHL
	case 0b0001:
		return SHR
	case 0b0010:
		return RAL
	case 0b0011:
		return RAR
	case 0b0100:
		return LDM
	case 0b0101:
		return ADM
	case 0b0110:
		return SBM
	case 0b0111:
		return STM
	case 0b1000:
		return ORM
	case 0b1001:
		return ANM
	case 0b1010:
		return XNM
	case 0b1011:
		return LDN
	case 0b1100:
		return ADN
	case 0b1101:
		return SBN
	case 0b1110:
		return STN
	case 0b1111:
		return JSN
	}
	panic("unreachable code")
}

func decodeSAP3Opr(selectcode uint16) sap3instr {
	switch selectcode {
	case 0b0000:
		return NOP
	case 0b0001:
		return CLA
	case 0b0010:
		return XCH
	case 0b0011:
		return DEX
	case 0b0100:
		return INX
	case 0b0101:
		return CMA
	case 0b0110:
		return CMB
	case 0b0111:
		return IOR
	case 0b1000:
		return AND
	case 0b1001:
		return NOR
	case 0b1010:
		return NAN
	case 0b1011:
		return XOR
	case 0b1100:
		return BRB
	case 0b1101:
		return INP
	case 0b1110:
		return OUT
	case 0b1111:
		return HLT
	}
	panic("unreachable code")
}
