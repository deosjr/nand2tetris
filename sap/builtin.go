package main

// these are the more optimised versions of the chips defined earlier
// where we abstract a little more over the clock

type Builtin interface {
	ClockTick()
}

type BuiltinJK struct {
	j, k   bit
	q, qBar  bit
}

func NewBuiltinJK() *BuiltinJK {
	return &BuiltinJK{qBar:true}
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
		b.q = true  // 1
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
 
func (r *RAM256x12) SendIn(in uint16) { r.in = in }
func (r *RAM256x12) SendLoad(load bool) { r.load = load }
func (r *RAM256x12) Out() uint16        { return r.out }
 
func (r *RAM256x12) ClockTick() {
	if r.load {
		r.mem[r.address] = r.in & 0xFFF
	}
	r.out = r.mem[r.address]
}