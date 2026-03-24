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
		instr, err := encodeASM(s)
		if err != nil {
			return out, err
		}
		out[i] = instr
	}
	return out, nil
}

func encodeASM(s string) (uint8, error) {
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
		return uint8(0b00000000) | uint8(dest), nil
	case "ADD":
		return uint8(0b00010000) | uint8(dest), nil
	case "SUB":
		return uint8(0b00100000) | uint8(dest), nil
	}
	return 0, fmt.Errorf("invalid opcode format")
}
