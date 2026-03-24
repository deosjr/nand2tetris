package main

// these are the more optimised versions of the chips defined earlier
// where we abstract a little more over the clock

type Builtin interface {
	ClockTick()
}

type BuiltinBit struct {
	in   bool
	out  bool
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
	in   uint8
	out  uint8
	load bool
}

func NewBuiltinRegister() *BuiltinRegister {
	return &BuiltinRegister{}
}

func (b *BuiltinRegister) SendIn(in uint8) {
	b.in = in
}

func (b *BuiltinRegister) SendLoad(load bool) {
	b.load = load
}

func (b *BuiltinRegister) Out() uint8 {
	return b.out
}

func (b *BuiltinRegister) ClockTick() {
	if b.load {
		b.out = b.in
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

func (b *ROM16x8) ClockTick() {
}

type BuiltinCounter struct {
	// unused in SAP-1, but its a 4-bit register
	in    uint8
	out   uint8
	inc   bool
	load  bool
	reset bool
}

func NewBuiltinCounter() *BuiltinCounter {
	return &BuiltinCounter{}
}

func (b *BuiltinCounter) SendIn(in uint8) {
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

func (b *BuiltinCounter) Out() uint8 {
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
