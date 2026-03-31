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
	switch s {
	case "NOP":
		return 0b111100000000, nil
	case "CLA":
		return 0b111100010000, nil
	case "XCH":
		return 0b111100100000, nil
	case "DEX":
		return 0b111100110000, nil
	case "INX":
		return 0b111101000000, nil
	case "CMA":
		return 0b111101010000, nil
	case "CMB":
		return 0b111101100000, nil
	case "IOR":
		return 0b111101110000, nil
	case "AND":
		return 0b111110000000, nil
	case "NOR":
		return 0b111110010000, nil
	case "NAN":
		return 0b111110100000, nil
	case "XOR":
		return 0b111110110000, nil
	case "BRB":
		return 0b111111000000, nil
	case "INP":
		return 0b111111010000, nil
	case "OUT":
		return 0b111111100000, nil
	case "HLT":
		return 0b111111110000, nil
	}
	split := strings.Split(s, " ")
	if len(split) != 2 {
		return 0, fmt.Errorf("invalid opcode format")
	}
	dest, err := strconv.ParseUint(split[1], 16, 16)
	if err != nil {
		return 0, err
	}
	switch split[0] {
	case "LDA":
		return 0b000000000000 | uint16(dest), nil
	case "ADD":
		return 0b000100000000 | uint16(dest), nil
	case "SUB":
		return 0b001000000000 | uint16(dest), nil
	case "STA":
		return 0b001100000000 | uint16(dest), nil
	case "LDB":
		return 0b010000000000 | uint16(dest), nil
	case "LDX":
		return 0b010100000000 | uint16(dest), nil
	case "JMP":
		return 0b011000000000 | uint16(dest), nil
	case "JAM":
		return 0b011100000000 | uint16(dest), nil
	case "JAZ":
		return 0b100000000000 | uint16(dest), nil
	case "JIM":
		return 0b100100000000 | uint16(dest), nil
	case "JIZ":
		return 0b101000000000 | uint16(dest), nil
	case "JMS":
		return 0b101100000000 | uint16(dest), nil
	}
	return 0, fmt.Errorf("invalid opcode format")
}