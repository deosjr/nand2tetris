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
