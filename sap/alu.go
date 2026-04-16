package main

func HalfAdder(a, b bit) (carry, sum bit) {
	return And(a, b), Xor(a, b)
}

func FullAdder(a, b, c bit) (carry, sum bit) {
	c1, s1 := HalfAdder(a, b)
	c2, s2 := HalfAdder(s1, c)
	return Or(c1, c2), s2
}

func Add16(a, b [16]bit) (out [16]bit) {
	var c bit
	for i := 0; i < 16; i++ {
		carry, sum := FullAdder(a[15-i], b[15-i], c)
		out[15-i] = sum
		c = carry
	}
	return out
}

// SAP-1 ALU is just a 2's complement adder/subtractor
// Figure 4.9
func SAP1Alu(a, b [8]bit, sub bit) (out [8]bit) {
	c := sub
	for i := 0; i < 8; i++ {
		carry, sum := FullAdder(a[7-i], Xor(b[7-i], sub), c)
		out[7-i] = sum
		c = carry
	}
	return out
}

// SAP-2 ALU is more complicated and not fully described in the book
func SAP2Alu(a, b [12]bit, s3, s2, s1, s0 bit) [12]bit {
	n3, n2, n1, n0 := Not(s3), Not(s2), Not(s1), Not(s0)

	//isNULL := And(And(n3, n2), And(n1, n0)) // 0000
	isADD := And(And(n3, n2), And(n1, s0)) // 0001
	isSUB := And(And(n3, n2), And(s1, n0)) // 0010
	isCMA := And(And(n3, n2), And(s1, s0)) // 0011
	isCMB := And(And(n3, s2), And(n1, n0)) // 0100
	isIOR := And(And(n3, s2), And(n1, s0)) // 0101
	isAND := And(And(n3, s2), And(s1, n0)) // 0110
	isNOR := And(And(n3, s2), And(s1, s0)) // 0111
	isNAN := And(And(s3, n2), And(n1, n0)) // 1000
	isXOR := And(And(s3, n2), And(n1, s0)) // 1001

	// 2's complement adder/subtractor, 12-bits
	c := isSUB
	var out [12]bit
	for i := 0; i < 12; i++ {
		carry, sum := FullAdder(a[11-i], Xor(b[11-i], isSUB), c)
		out[11-i] = sum
		c = carry
	}

	// 6 possible instructions unused, we will default to isNULL
	var zero [12]bit
	return Mux12(
		Mux12(
			Mux12(
				Mux12(
					Mux12(
						Mux12(
							Mux12(
								Mux12(zero, Xor12(a, b), isXOR),
								Not12(And12(a, b)), isNAN),
							Not12(Or12(a, b)), isNOR),
						And12(a, b), isAND),
					Or12(a, b), isIOR),
				Not12(b), isCMB),
			Not12(a), isCMA),
		out, Or(isADD, isSUB))
}

// SAP-3 ALU is the same as SAP-2, but for 16 bits
func SAP3Alu(a, b [16]bit, s3, s2, s1, s0 bit) uint16 {
	n3, n2, n1, n0 := Not(s3), Not(s2), Not(s1), Not(s0)

	//isNULL := And(And(n3, n2), And(n1, n0)) // 0000
	isADD := And(And(n3, n2), And(n1, s0)) // 0001
	isSUB := And(And(n3, n2), And(s1, n0)) // 0010
	isCMA := And(And(n3, n2), And(s1, s0)) // 0011
	isCMB := And(And(n3, s2), And(n1, n0)) // 0100
	isIOR := And(And(n3, s2), And(n1, s0)) // 0101
	isAND := And(And(n3, s2), And(s1, n0)) // 0110
	isNOR := And(And(n3, s2), And(s1, s0)) // 0111
	isNAN := And(And(s3, n2), And(n1, n0)) // 1000
	isXOR := And(And(s3, n2), And(n1, s0)) // 1001

	// 2's complement adder/subtractor, 16-bits
	c := isSUB
	var out [16]bit
	for i := 0; i < 16; i++ {
		carry, sum := FullAdder(a[15-i], Xor(b[15-i], isSUB), c)
		out[15-i] = sum
		c = carry
	}

	// 6 possible instructions unused, we will default to isNULL
	var zero [16]bit
	return fromBit16(Mux16(
		Mux16(
			Mux16(
				Mux16(
					Mux16(
						Mux16(
							Mux16(
								Mux16(zero, Xor16(a, b), isXOR),
								Not16(And16(a, b)), isNAN),
							Not16(Or16(a, b)), isNOR),
						And16(a, b), isAND),
					Or16(a, b), isIOR),
				Not16(b), isCMB),
			Not16(a), isCMA),
		out, Or(isADD, isSUB)))
}
