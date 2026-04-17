package main

import "math/bits"

// these are the more optimised versions of the chips defined earlier
// where we abstract a little more over the clock

type Builtin interface {
	ClockTick()
}

type BuiltinJK struct {
	j, k    bit
	q, qBar bit
}

func NewBuiltinJK() *BuiltinJK {
	return &BuiltinJK{qBar: true}
}

func (b *BuiltinJK) SendJ(j bit) {
	b.j = j
}

func (b *BuiltinJK) SendK(k bit) {
	b.k = k
}

func (b *BuiltinJK) OutQ() bit {
	return b.q
}

func (b *BuiltinJK) OutQBar() bit {
	return b.qBar
}

func (b *BuiltinJK) ClockTick() {
	if b.j && b.k {
		b.q, b.qBar = b.qBar, b.q
		return
	}
	if b.j {
		b.q = true // 1
	}
	if b.k {
		b.q = false // 0
	}
	b.qBar = !b.q
}

type BuiltinRegister4 struct {
	in   uint8
	out  uint8
	load bool
}

func NewBuiltinRegister4() *BuiltinRegister4 {
	return &BuiltinRegister4{}
}

func (b *BuiltinRegister4) SendIn(in uint8) {
	b.in = in
}

func (b *BuiltinRegister4) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinRegister4) Out() uint8 {
	return b.out
}

func (b *BuiltinRegister4) ClockTick() {
	if b.load {
		b.out = b.in & 0xF
	}
}

type BuiltinRegister8 struct {
	in   uint8
	out  uint8
	load bool
}

func NewBuiltinRegister8() *BuiltinRegister8 {
	return &BuiltinRegister8{}
}

func (b *BuiltinRegister8) SendIn(in uint8) {
	b.in = in
}

func (b *BuiltinRegister8) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinRegister8) Out() uint8 {
	return b.out
}

func (b *BuiltinRegister8) ClockTick() {
	if b.load {
		b.out = b.in
	}
}

type BuiltinRegister12 struct {
	in   uint16
	out  uint16
	load bool
}

func NewBuiltinRegister12() *BuiltinRegister12 {
	return &BuiltinRegister12{}
}

func (b *BuiltinRegister12) SendIn(in uint16) {
	b.in = in
}

func (b *BuiltinRegister12) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinRegister12) Out() uint16 {
	return b.out
}

func (b *BuiltinRegister12) ClockTick() {
	if b.load {
		b.out = b.in & 0xFFF
	}
}

type BuiltinRegister16 struct {
	in   uint16
	out  uint16
	load bool
}

func NewBuiltinRegister16() *BuiltinRegister16 {
	return &BuiltinRegister16{}
}

func (b *BuiltinRegister16) SendIn(in uint16) {
	b.in = in
}

func (b *BuiltinRegister16) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinRegister16) Out() uint16 {
	return b.out
}

func (b *BuiltinRegister16) ClockTick() {
	if b.load {
		b.out = b.in
	}
}

type ShiftRegister16 struct {
	BuiltinRegister16
	shl, shr, ral, rar bool
}

func NewShiftRegister16() *ShiftRegister16 {
	return &ShiftRegister16{}
}

func (b *ShiftRegister16) SendSHL(in bool) {
	b.shl = in
}
func (b *ShiftRegister16) SendSHR(in bool) {
	b.shr = in
}
func (b *ShiftRegister16) SendRAL(in bool) {
	b.ral = in
}
func (b *ShiftRegister16) SendRAR(in bool) {
	b.rar = in
}

func (b *ShiftRegister16) ClockTick() {
	switch {
	case b.load:
		b.out = b.in
	case b.shl:
		b.out = b.out << 1
	case b.shr:
		b.out = b.out >> 1
	case b.ral:
		b.out = bits.RotateLeft16(b.out, 1)
	case b.rar:
		b.out = bits.RotateLeft16(b.out, -1)
	}
}

type ROM16x8 struct {
	mem     [16]uint8
	address uint8 // 4 bit in spec
	out     uint8
}

func NewROM16x8() *ROM16x8 {
	return &ROM16x8{}
}

func (b *ROM16x8) SendAddress(addr uint8) {
	if addr >= 16 {
		panic("ROM addr too big")
	}
	b.address = addr
}

func (b *ROM16x8) Out() uint8 {
	return b.mem[b.address]
}

func (b *ROM16x8) ClockTick() {}

type BuiltinCounter4 struct {
	in    uint8
	out   uint8
	inc   bool
	load  bool
	reset bool
}

func NewBuiltinCounter4() *BuiltinCounter4 {
	return &BuiltinCounter4{}
}

func (b *BuiltinCounter4) SendIn(in uint8) {
	b.in = in & 0xF
}

func (b *BuiltinCounter4) SendInc(inc bool) {
	b.inc = inc
}

func (b *BuiltinCounter4) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinCounter4) SendReset(reset bool) {
	b.reset = reset
}

func (b *BuiltinCounter4) Out() uint8 {
	return b.out
}

func (b *BuiltinCounter4) ClockTick() {
	switch {
	case b.reset:
		b.out = 0
	case b.load:
		b.out = b.in
	case b.inc:
		b.out = (b.out + 1) & 0xF
	}
}

type BuiltinCounter8 struct {
	in    uint8
	out   uint8
	inc   bool
	load  bool
	reset bool
}

func NewBuiltinCounter8() *BuiltinCounter8 {
	return &BuiltinCounter8{}
}

func (b *BuiltinCounter8) SendIn(in uint8) {
	b.in = in
}

func (b *BuiltinCounter8) SendInc(inc bool) {
	b.inc = inc
}

func (b *BuiltinCounter8) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinCounter8) SendReset(reset bool) {
	b.reset = reset
}

func (b *BuiltinCounter8) Out() uint8 {
	return b.out
}

func (b *BuiltinCounter8) ClockTick() {
	switch {
	case b.reset:
		b.out = 0
	case b.load:
		b.out = b.in
	case b.inc:
		b.out = b.out + 1
	}
}

type BuiltinCounter12 struct {
	in   uint16
	out  uint16
	inc  bool
	dec  bool
	load bool
}

func NewBuiltinCounter12() *BuiltinCounter12 {
	return &BuiltinCounter12{}
}

func (b *BuiltinCounter12) SendIn(in uint16) {
	b.in = in
}

func (b *BuiltinCounter12) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinCounter12) SendIncr(inc bool) {
	b.inc = inc
}

func (b *BuiltinCounter12) SendDecr(dec bool) {
	b.dec = dec
}

func (b *BuiltinCounter12) Out() uint16 {
	return b.out
}

func (b *BuiltinCounter12) ClockTick() {
	switch {
	case b.load:
		b.out = b.in & 0xFFF
	case b.inc:
		b.out = (b.out + 1) & 0xFFF
	case b.dec:
		b.out = (b.out - 1) & 0xFFF
	}
}

type BuiltinCounter16 struct {
	in   uint16
	out  uint16
	inc  bool
	dec  bool
	load bool
}

func NewBuiltinCounter16() *BuiltinCounter16 {
	return &BuiltinCounter16{}
}

func (b *BuiltinCounter16) SendIn(in uint16) {
	b.in = in
}

func (b *BuiltinCounter16) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinCounter16) SendIncr(inc bool) {
	b.inc = inc
}

func (b *BuiltinCounter16) SendDecr(dec bool) {
	b.dec = dec
}

func (b *BuiltinCounter16) Out() uint16 {
	return b.out
}

func (b *BuiltinCounter16) ClockTick() {
	switch {
	case b.load:
		b.out = b.in
	case b.inc:
		b.out = (b.out + 1)
	case b.dec:
		b.out = (b.out - 1)
	}
}

type RAM256x12 struct {
	address uint8
	in      uint16
	out     uint16
	load    bool
	mem     [256]uint16
}

func NewRAM256x12() *RAM256x12 {
	return &RAM256x12{}
}

func (r *RAM256x12) SendAddress(addr uint8) {
	r.address = addr
	r.out = r.mem[r.address]
}

func (r *RAM256x12) SendIn(in uint16)   { r.in = in }
func (r *RAM256x12) SendLoad(load bool) { r.load = load }
func (r *RAM256x12) Out() uint16        { return r.out }

func (r *RAM256x12) ClockTick() {
	if r.load {
		r.mem[r.address] = r.in & 0xFFF
	}
	r.out = r.mem[r.address]
}

type RAM4096x16 struct {
	address uint16
	in      uint16
	out     uint16
	load    bool
	mem     [4096]uint16
}

func NewRAM4096x16() *RAM4096x16 {
	return &RAM4096x16{}
}

func (r *RAM4096x16) SendAddress(addr uint16) {
	r.address = addr & 0xFFF
	r.out = r.mem[r.address]
}

func (r *RAM4096x16) SendIn(in uint16)   { r.in = in }
func (r *RAM4096x16) SendLoad(load bool) { r.load = load }
func (r *RAM4096x16) Out() uint16        { return r.out }

func (r *RAM4096x16) ClockTick() {
	if r.load {
		r.mem[r.address] = r.in
	}
	r.out = r.mem[r.address]
}

type IPLCounter12 struct {
	BuiltinCounter12
	ipl bool
}

func NewIPLCounter12() *IPLCounter12 {
	return &IPLCounter12{}
}

func (c *IPLCounter12) SendIPL(ipl bool) {
	c.ipl = ipl
}

func (c *IPLCounter12) ClockTick() {
	switch {
	case c.load:
		in := c.in & 0xFFF
		if c.ipl {
			// page bits stay unchanged
			in = in & 0xFF
			page := c.out & 0xF00
			in = in | page
		}
		c.out = in
	case c.inc:
		c.out = (c.out + 1) & 0xFFF
	case c.dec:
		c.out = (c.out - 1) & 0xFFF
	}
}

// Figure 10-7: Stack registers and internal bus
type Stack4x12 struct {
	PC   *IPLCounter12
	SC1  *IPLCounter12
	SC2  *IPLCounter12
	SC3  *IPLCounter12
	in   uint16
	load bool
	inc  bool
	emit bool
	ipl  bool
	out  uint16
	// 2-bit up/down counter
	q0, q1 bit
	pU, pD bool
}

func NewBuiltinStack4x12() *Stack4x12 {
	return &Stack4x12{
		PC:  NewIPLCounter12(),
		SC1: NewIPLCounter12(),
		SC2: NewIPLCounter12(),
		SC3: NewIPLCounter12(),
	}
}

func (s *Stack4x12) SendReset(r bool) {
	// unimplemented
}

func (s *Stack4x12) SendPU(in bit) {
	s.pU = bool(in)
}

func (s *Stack4x12) SendPD(in bit) {
	s.pD = bool(in)
}

func (s *Stack4x12) SendIn(in uint16) {
	s.in = in
}

func (s *Stack4x12) SendLoad(load bool) {
	s.load = load
}

func (s *Stack4x12) SendInc(inc bool) {
	s.inc = inc
}

func (s *Stack4x12) SendEmit(e bool) {
	s.emit = e
}

func (s *Stack4x12) SendIPL(ipl bool) {
	s.ipl = ipl
}

func (s *Stack4x12) Out() uint16 {
	if !s.emit {
		return 0
	}
	// Figure 10-9: stack pointer
	p0 := And(Not(s.q0), Not(s.q1))
	p1 := And(s.q0, Not(s.q1))
	p2 := And(Not(s.q0), s.q1)
	p3 := And(s.q0, s.q1)
	low := toBit12(0)
	outPC := Mux12(low, toBit12(s.PC.Out()), p0)
	outSC1 := Mux12(low, toBit12(s.SC1.Out()), p1)
	outSC2 := Mux12(low, toBit12(s.SC2.Out()), p2)
	outSC3 := Mux12(low, toBit12(s.SC3.Out()), p3)
	return fromBit12(Or12(outPC, Or12(outSC1, Or12(outSC2, outSC3))))
}

// TODO: IPL, inhibit page load
func (s *Stack4x12) ClockTick() {
	// note: none of this guarantees overflow protection
	if s.pU {
		// 2 bit incr
		if s.q0 {
			s.q1 = !s.q1
		}
		s.q0 = !s.q0
	}
	if s.pD {
		// 2 bit decr
		if !s.q0 {
			s.q1 = !s.q1
		}
		s.q0 = !s.q0
	}
	// Figure 10-9: stack pointer
	p0 := And(Not(s.q0), Not(s.q1))
	p1 := And(s.q0, Not(s.q1))
	p2 := And(Not(s.q0), s.q1)
	p3 := And(s.q0, s.q1)

	// 12-bit bus W
	w := fromBit12(Or12(toBit12(s.in), toBit12(s.Out())))

	// Figure 10-10: stack demultiplexer
	l, c := bit(s.load), bit(s.inc)
	s.PC.SendIn(w)
	s.PC.SendIPL(s.ipl)
	s.PC.SendLoad(bool(And(l, p0)))
	s.PC.SendIncr(bool(And(c, p0)))
	s.SC1.SendIn(w)
	s.SC1.SendIPL(s.ipl)
	s.SC1.SendLoad(bool(And(l, p1)))
	s.SC1.SendIncr(bool(And(c, p1)))
	s.SC2.SendIn(w)
	s.SC2.SendIPL(s.ipl)
	s.SC2.SendLoad(bool(And(l, p2)))
	s.SC2.SendIncr(bool(And(c, p2)))
	s.SC3.SendIn(w)
	s.SC3.SendIPL(s.ipl)
	s.SC3.SendLoad(bool(And(l, p3)))
	s.SC3.SendIncr(bool(And(c, p3)))

	s.PC.ClockTick()
	s.SC1.ClockTick()
	s.SC2.ClockTick()
	s.SC3.ClockTick()
}
