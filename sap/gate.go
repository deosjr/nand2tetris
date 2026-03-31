package main

// primitive, uses builtin go
func Nand(a, b bit) bit {
	return !(a && b)
}

// rest is not primitive, derives from Nand
func Not(in bit) bit {
	return !in //Nand(in, in)
}

func And(a, b bit) bit {
	return a && b //Not(Nand(a, b))
}

func Or(a, b bit) bit {
	return a || b //Nand(Not(a), Not(b))
}

func Xor(a, b bit) bit {
	return (!a && b) || (a && !b)
	// without this declaration we would have to calculate nandab twice
	//nandab := Nand(a, b)
	//return Nand(Nand(a, nandab), Nand(b, nandab))
}

func Mux(a, b, sel bit) bit {
	if sel {
		return b
	}
	return a
	//return Or(And(a, Not(sel)), And(b, sel))
}

func DMux(in, sel bit) (bit, bit) {
	return in && !sel, in && sel
	//return And(in, Not(sel)), And(in, sel)
}

func And8(a, b [8]bit) (out [8]bit) {
	for i := 0; i < 8; i++ {
		out[i] = And(a[i], b[i])
	}
	return out
}

func Or8(a, b [8]bit) (out [8]bit) {
	for i := 0; i < 8; i++ {
		out[i] = Or(a[i], b[i])
	}
	return out
}

func Mux8(a, b [8]bit, sel bit) (out [8]bit) {
	if sel {
		return b
	}
	return a
}

func Not12(in [12]bit) (out [12]bit) {
	for i := 0; i < 12; i++ {
		out[i] = Not(in[i])
	}
	return out
}

func And12(a, b [12]bit) (out [12]bit) {
	for i := 0; i < 12; i++ {
		out[i] = And(a[i], b[i])
	}
	return out
}

func Or12(a, b [12]bit) (out [12]bit) {
	for i := 0; i < 12; i++ {
		out[i] = Or(a[i], b[i])
	}
	return out
}

func Xor12(a, b [12]bit) (out [12]bit) {
	for i := 0; i < 12; i++ {
		out[i] = Or(a[i], b[i])
	}
	return out
}

func Mux12(a, b [12]bit, sel bit) (out [12]bit) {
	if sel {
		return b
	}
	return a
}

func Or12Way(in [12]bit) bit {
	out := in[0]
	for i := 1; i < 12; i++ {
		out = Or(out, in[i])
	}
	return out
}