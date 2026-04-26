package main

import (
	"fmt"
	"os"
	"regexp"
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

// labelRe matches a line that is nothing but a label declaration, e.g. "LOOP:".
// Leading/trailing whitespace on the line is tolerated.
var labelRe = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*:$`)

// placeholderRe matches a braced symbolic reference inside an instruction.
//
//	{FOO}   — 12-bit absolute address of label FOO
//	{:FOO}  — 8-bit page-relative offset to FOO, must be on the current page
var placeholderRe = regexp.MustCompile(`\{(:?)([A-Z_][A-Z0-9_]*)\}`)

// parseLabels walks src and extracts label-only lines ("LOOP:") into a label
// table, returning the source with those lines removed. Each label records
// the absolute address (base + position-in-output) of the next emitted word.
// Consecutive labels share that address; a label at the end points one past
// the final emitted word.
func parseLabels(base uint16, src []string) ([]string, map[string]uint16, error) {
	labels := map[string]uint16{}
	clean := make([]string, 0, len(src))
	for i, line := range src {
		trimmed := strings.TrimSpace(line)
		if labelRe.MatchString(trimmed) {
			name := trimmed[:len(trimmed)-1]
			if _, exists := labels[name]; exists {
				return nil, nil, fmt.Errorf("line %d: duplicate label %q", i, name)
			}
			labels[name] = base + uint16(len(clean))
			continue
		}
		clean = append(clean, line)
	}
	return clean, labels, nil
}

// resolve substitutes {FOO} and {:FOO} placeholders in each line with the
// resolved address, producing concrete assembly text that encodeASM3 accepts.
// Labels in `labels` (defined within the block) take precedence over
// `externals` (sysvars, fixed-address symbols). Page-relative references
// assert that the target lives on the same 256-word page as the referring
// instruction.
func resolve(base uint16, src []string, labels, externals map[string]uint16) ([]string, error) {
	lookup := func(name string) (uint16, bool) {
		if v, ok := labels[name]; ok {
			return v, true
		}
		v, ok := externals[name]
		return v, ok
	}
	out := make([]string, len(src))
	for i, line := range src {
		addr := base + uint16(i)
		var loopErr error
		replaced := placeholderRe.ReplaceAllStringFunc(line, func(m string) string {
			if loopErr != nil {
				return m
			}
			parts := placeholderRe.FindStringSubmatch(m)
			pageRel := parts[1] == ":"
			name := parts[2]
			target, ok := lookup(name)
			if !ok {
				loopErr = fmt.Errorf("line %d (addr 0x%X): unresolved label %q in %q", i, addr, name, line)
				return m
			}
			if pageRel {
				if target&0xFF00 != addr&0xFF00 {
					loopErr = fmt.Errorf("line %d (addr 0x%X): page-relative {:%s} targets 0x%X which is on a different page", i, addr, name, target)
					return m
				}
				return fmt.Sprintf("%X", target&0xFF)
			}
			return fmt.Sprintf("%X", target&0xFFF)
		})
		if loopErr != nil {
			return nil, loopErr
		}
		out[i] = replaced
	}
	return out, nil
}

// assembleSAP3Labeled assembles a labeled source block into words placed at
// base..base+N. The source may contain:
//
//   - label-only lines like "LOOP:" that define a symbol pointing at the next
//     emitted word and emit nothing themselves;
//   - instructions or data words with {FOO} / {:FOO} placeholders for
//     symbolic references (absolute vs. same-page respectively);
//   - ordinary lines in the format encodeASM3 already accepts.
//
// `externals` supplies addresses for symbols the block references but does
// not define (sysvars, fixed-address outputs). Returns the emitted words and
// the label table resolved during this pass.
func assembleSAP3Labeled(base uint16, src []string, externals map[string]uint16) ([]uint16, map[string]uint16, error) {
	clean, labels, err := parseLabels(base, src)
	if err != nil {
		return nil, nil, err
	}
	resolved, err := resolve(base, clean, labels, externals)
	if err != nil {
		return nil, labels, err
	}
	out := make([]uint16, len(resolved))
	for i, line := range resolved {
		w, err := encodeASM3(line)
		if err != nil {
			return nil, labels, fmt.Errorf("line %d (addr 0x%X) %q: %w", i, base+uint16(i), line, err)
		}
		out[i] = w
	}
	return out, labels, nil
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
	if len(s) > 2 && s[0] == '0' && s[1] == 'x' {
		n, err := strconv.ParseUint(s[2:], 16, 16)
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
		return 0, fmt.Errorf("invalid opcode format %s", s)
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
		x, err = strconv.ParseUint(splitcomma[1], 16, 8)
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
		return 0, fmt.Errorf("invalid opcode format %s", s)
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
	return 0, fmt.Errorf("invalid opcode format %s", s)
}

// readAsmFile reads a SAP3 assembly source file and returns a []string
// compatible with assembleSAP3Labeled. Rules:
//   - Each non-empty line becomes one element.
//   - Anything from ';' to end-of-line is a comment and is stripped.
//   - Blank lines (and comment-only lines) are skipped entirely.
//   - Leading/trailing whitespace is trimmed.
func readAsmFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.Index(line, ";"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
