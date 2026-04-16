package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func assembleSAP1FromStrings(strs [16]string) ([16]uint8, error) {
	out := [16]uint8{}
	for i, s := range strs {
		instr, err := encodeASM1(s)
		if err != nil {
			return out, err
		}
		out[i] = instr
	}
	return out, nil
}

func encodeASM1(s string) (uint8, error) {
	if s == "" {
		return 0, nil
	}
	if unicode.IsDigit(rune(s[0])) {
		n, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(n), nil
	}
	if s == "OUT" {
		return 0b11100000, nil
	}
	if s == "HLT" {
		return 0b11110000, nil
	}
	split := strings.Split(s, " ")
	if len(split) != 2 {
		return 0, fmt.Errorf("invalid opcode format")
	}
	if len(split[1]) != 2 || split[1][0] != 'R' {
		return 0, fmt.Errorf("invalid opcode format")
	}
	dest, err := strconv.ParseUint(split[1][1:], 16, 8)
	if err != nil {
		return 0, err
	}
	switch split[0] {
	case "LDA":
		return 0b00000000 | uint8(dest), nil
	case "ADD":
		return 0b00010000 | uint8(dest), nil
	case "SUB":
		return 0b00100000 | uint8(dest), nil
	}
	return 0, fmt.Errorf("invalid opcode format")
}

func assembleSAP2FromStrings(strs []string) ([256]uint16, error) {
	out := [256]uint16{}
	for i, s := range strs {
		instr, err := encodeASM2(s)
		if err != nil {
			return out, err
		}
		out[i] = instr
	}
	return out, nil
}

func encodeASM2(s string) (uint16, error) {
	if s == "" {
		return 0, nil
	}
	if unicode.IsDigit(rune(s[0])) {
		n, err := strconv.ParseUint(s, 10, 12)
		if err != nil {
			return 0, err
		}
		return uint16(n), nil
	}
	opr := func(selectcode uint16) uint16 {
		return 0xF00 | selectcode<<4
	}
	switch s {
	case "NOP":
		return opr(0b0000), nil
	case "CLA":
		return opr(0b0001), nil
	case "XCH":
		return opr(0b0010), nil
	case "DEX":
		return opr(0b0011), nil
	case "INX":
		return opr(0b0100), nil
	case "CMA":
		return opr(0b0101), nil
	case "CMB":
		return opr(0b0110), nil
	case "IOR":
		return opr(0b0111), nil
	case "AND":
		return opr(0b1000), nil
	case "NOR":
		return opr(0b1001), nil
	case "NAN":
		return opr(0b1010), nil
	case "XOR":
		return opr(0b1011), nil
	case "BRB":
		return opr(0b1100), nil
	case "INP":
		return opr(0b1101), nil
	case "OUT":
		return opr(0b1110), nil
	case "HLT":
		return opr(0b1111), nil
	}
	split := strings.Split(s, " ")
	if len(split) != 2 {
		return 0, fmt.Errorf("invalid opcode format")
	}
	dest, err := strconv.ParseUint(split[1], 16, 16)
	if err != nil {
		return 0, err
	}
	// MRIs and Jumps
	switch split[0] {
	case "LDA":
		return 0b0000<<8 | uint16(dest), nil
	case "ADD":
		return 0b0001<<8 | uint16(dest), nil
	case "SUB":
		return 0b0010<<8 | uint16(dest), nil
	case "STA":
		return 0b0011<<8 | uint16(dest), nil
	case "LDB":
		return 0b0100<<8 | uint16(dest), nil
	case "LDX":
		return 0b0101<<8 | uint16(dest), nil
	case "JMP":
		return 0b0110<<8 | uint16(dest), nil
	case "JAM":
		return 0b0111<<8 | uint16(dest), nil
	case "JAZ":
		return 0b1000<<8 | uint16(dest), nil
	case "JIM":
		return 0b1001<<8 | uint16(dest), nil
	case "JIZ":
		return 0b1010<<8 | uint16(dest), nil
	case "JMS":
		return 0b1011<<8 | uint16(dest), nil
	}
	return 0, fmt.Errorf("invalid opcode format")
}

func assembleSAP3FromStrings(strs []string) ([4096]uint16, error) {
	out := [4096]uint16{}
	for i, s := range strs {
		instr, err := encodeASM3(s)
		if err != nil {
			return out, err
		}
		out[i] = instr
	}
	return out, nil
}

func encodeASM3(s string) (uint16, error) {
	if s == "" {
		return 0, nil
	}
	if unicode.IsDigit(rune(s[0])) {
		n, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(n), nil
	}
	opr := func(selectcode uint16) uint16 {
		return 0xF000 | selectcode<<8
	}
	mix := func(selectcode uint16) uint16 {
		return 0xE000 | selectcode<<8
	}
	// no argument
	switch s {
	case "NOP":
		return opr(0b0000), nil
	case "CLA":
		return opr(0b0001), nil
	case "CMA":
		return opr(0b0101), nil
	case "CMB":
		return opr(0b0110), nil
	case "IOR":
		return opr(0b0111), nil
	case "AND":
		return opr(0b1000), nil
	case "NOR":
		return opr(0b1001), nil
	case "NAN":
		return opr(0b1010), nil
	case "XOR":
		return opr(0b1011), nil
	case "BRB":
		return opr(0b1100), nil
	case "HLT":
		return opr(0b1111), nil
	case "SHL":
		return mix(0b0000), nil
	case "SHR":
		return mix(0b0001), nil
	case "RAL":
		return mix(0b0010), nil
	case "RAR":
		return mix(0b0011), nil
	case "LDM":
		return mix(0b0100), nil
	case "ADM":
		return mix(0b0101), nil
	case "SBM":
		return mix(0b0110), nil
	case "STM":
		return mix(0b0111), nil
	case "ORM":
		return mix(0b1000), nil
	case "ANM":
		return mix(0b1001), nil
	case "XNM":
		return mix(0b1010), nil
	}
	split := strings.Split(s, " ")
	if len(split) != 2 {
		return 0, fmt.Errorf("invalid opcode format")
	}
	x, err := strconv.ParseUint(split[1], 16, 16)
	if err != nil {
		// same-page instructions
		splitcomma := strings.Split(split[1], ",")
		if len(splitcomma) != 2 {
			return 0, err
		}
		x, err := strconv.ParseUint(splitcomma[0], 16, 4)
		if err != nil {
			return 0, err
		}
		idx := uint16(x)
		x, err = strconv.ParseUint(splitcomma[0], 16, 8)
		if err != nil {
			return 0, err
		}
		w := uint16(x)
		switch split[0] {
		case "LDX":
			return 0b0101<<12 | idx<<8 | w, nil
		case "JIM":
			return 0b1001<<12 | idx<<8 | w, nil
		case "JIZ":
			return 0b1010<<12 | idx<<8 | w, nil
		}
		return 0, fmt.Errorf("invalid opcode format")
	}
	arg := uint16(x)
	switch split[0] {
	case "XCH":
		return opr(0b0010) | arg<<4, nil
	case "DEX":
		return opr(0b0011) | arg<<4, nil
	case "INX":
		return opr(0b0100) | arg<<4, nil
	case "INP":
		return opr(0b1101) | arg<<4, nil
	case "OUT":
		return opr(0b1110) | arg<<4, nil
	case "LDN":
		return mix(0b1011) | arg<<4, nil
	case "ADN":
		return mix(0b1100) | arg<<4, nil
	case "SBN":
		return mix(0b1101) | arg<<4, nil
	case "STN":
		return mix(0b1110) | arg<<4, nil
	case "JSN":
		return mix(0b1111) | arg<<4, nil
	case "LDA":
		return 0b0000<<12 | arg, nil
	case "ADD":
		return 0b0001<<12 | arg, nil
	case "SUB":
		return 0b0010<<12 | arg, nil
	case "STA":
		return 0b0011<<12 | arg, nil
	case "LDB":
		return 0b0100<<12 | arg, nil
	case "JMP":
		return 0b0110<<12 | arg, nil
	case "JAM":
		return 0b0111<<12 | arg, nil
	case "JAZ":
		return 0b1000<<12 | arg, nil
	case "JMS":
		return 0b1011<<12 | arg, nil
	case "DSZ":
		return 0b1100<<12 | arg<<8, nil
	case "ISZ":
		return 0b1101<<12 | arg<<8, nil
	}
	return 0, fmt.Errorf("invalid opcode format")
}
