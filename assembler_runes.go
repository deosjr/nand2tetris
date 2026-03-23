package main

// Riffing on the fact that the word ALU has a meaning in Hack and runic magic
// And the fact that the Younger Futhark alphabet has exactly sixteen runes
// TODO: we can swap the instruction order of regular Hack such that dest/jump remain constant if unchanged
// This is achieved by swapping the unused bits into the last two nibbles. Won't work for extended CPU's I made.

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// take a .asm file and return machine language code
func AssembleRunes(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	contents := string(data)
	fset := token.NewFileSet()
	parsed, err := parse(fset, filename, contents)
	if err != nil {
		return nil, err
	}
	contents = strings.ReplaceAll(contents, "\n", "")
	return assembleRunes(fset, contents, parsed)
}

func assembleRunesFromString(s string) ([]string, error) {
	fset := token.NewFileSet()
	parsed, err := parse(fset, "string_input", s)
	if err != nil {
		return nil, err
	}
	s = strings.ReplaceAll(s, "\n", "")
	return assembleRunes(fset, s, parsed)
}

func assembleRunes(fset *token.FileSet, contents string, parsed *ast.File) ([]string, error) {
	// firstpass
	statements := []ast.Stmt{}
	labels := map[string]uint16{
		"SP":     0x0000,
		"LCL":    0x0001,
		"ARG":    0x0002,
		"THIS":   0x0003,
		"THAT":   0x0004,
		"R0":     0x0000,
		"R1":     0x0001,
		"R2":     0x0002,
		"R3":     0x0003,
		"R4":     0x0004,
		"R5":     0x0005,
		"R6":     0x0006,
		"R7":     0x0007,
		"R8":     0x0008,
		"R9":     0x0009,
		"R10":    0x000a,
		"R11":    0x000b,
		"R12":    0x000c,
		"R13":    0x000d,
		"R14":    0x000e,
		"R15":    0x000f,
		"SCREEN": 0x4000,
		"KBD":    0x6000,
	}
	for _, stmt := range parsed.Decls[0].(*ast.FuncDecl).Body.List {
		switch s := stmt.(type) {
		case *ast.ExprStmt:
			switch x := s.X.(type) {
			case *ast.ParenExpr:
				// label
				bl := x.X.(*ast.BasicLit)
				label := bl.Value
				if _, ok := labels[label]; ok {
					// syntax error
					posFrom := fset.Position(bl.ValuePos)
					return nil, fmt.Errorf("%s: label redeclaration: %s", posFrom.String(), contents[bl.ValuePos-1:bl.ValuePos+token.Pos(len(label))])
				}
				labels[label] = uint16(len(statements))
			case *ast.BasicLit:
				// AINSTR
				statements = append(statements, s)
			}
		case *ast.LabeledStmt:
			// CINSTR
			statements = append(statements, s)
		case *ast.BadStmt:
			// syntax error
			posFrom := fset.Position(s.From)
			return nil, fmt.Errorf("%s: syntax error: %s", posFrom.String(), contents[s.From-1:s.To+1])
		}
	}
	// secondpass
	program := []uint16{}
	vars := map[string]uint16{}
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.ExprStmt:
			// AINSTR
			bl := s.X.(*ast.BasicLit)
			switch bl.Kind {
			case token.STRING:
				if unicode.IsUpper(rune(bl.Value[0])) {
					// LABEL
					iptr, ok := labels[bl.Value]
					if !ok {
						posFrom := fset.Position(bl.ValuePos)
						return nil, fmt.Errorf("%s: label not found: %s", posFrom.String(), contents[bl.ValuePos:bl.ValuePos+token.Pos(len(bl.Value))])
					}
					program = append(program, iptr)
				} else {
					// var
					if strings.HasPrefix(bl.Value, "0x") {
						n, err := strconv.ParseInt(bl.Value[2:], 16, 16)
						if err != nil {
							posFrom := fset.Position(bl.ValuePos)
							return nil, fmt.Errorf("%s: cant parse: %s", posFrom.String(), contents[bl.ValuePos:bl.ValuePos+token.Pos(len(bl.Value))])
						}
						program = append(program, uint16(n))
						continue
					}
					if unicode.IsDigit(rune(bl.Value[0])) {
						posFrom := fset.Position(bl.ValuePos)
						return nil, fmt.Errorf("%s: var starting with digit: %s", posFrom.String(), contents[bl.ValuePos:bl.ValuePos+token.Pos(len(bl.Value))])
					}
					iptr, ok := vars[bl.Value]
					if !ok {
						iptr = uint16(16 + len(vars))
						vars[bl.Value] = iptr
					}
					program = append(program, iptr)
				}
			case token.INT:
				n, err := strconv.Atoi(bl.Value)
				if err != nil {
					posFrom := fset.Position(bl.ValuePos)
					return nil, fmt.Errorf("%s: cant parse: %s", posFrom.String(), contents[bl.ValuePos:bl.ValuePos+token.Pos(len(bl.Value))])
				}
				program = append(program, uint16(n))
			}
		case *ast.LabeledStmt:
			// CINSTR
			var instr uint16 = 0b1110000000000000
			switch s.Label {
			case LITJGT:
				instr |= 0b001
			case LITJEQ:
				instr |= 0b010
			case LITJGE:
				instr |= 0b011
			case LITJLT:
				instr |= 0b100
			case LITJNE:
				instr |= 0b101
			case LITJLE:
				instr |= 0b110
			case LITJMP:
				instr |= 0b111
			}
			as := s.Stmt.(*ast.AssignStmt)
			for _, d := range as.Lhs {
				switch d {
				case LITA:
					instr |= 0b100000
				case LITD:
					instr |= 0b010000
				case LITM:
					instr |= 0b001000
				}
			}
			switch t := as.Rhs[0].(type) {
			case *ast.Ident:
				switch t {
				case LIT0:
					instr |= 0b0000101010000000
				case LIT1:
					instr |= 0b0000111111000000
				case LITA:
					instr |= 0b0000110000000000
				case LITD:
					instr |= 0b0000001100000000
				case LITM:
					instr |= 0b0001110000000000
				}
			case *ast.UnaryExpr:
				switch t.Op {
				case token.SUB:
					switch t.X {
					case LIT1:
						instr |= 0b0000111010000000
					case LITA:
						instr |= 0b0000110011000000
					case LITD:
						instr |= 0b0000001111000000
					case LITM:
						instr |= 0b0001110011000000
					}
				case token.NOT:
					switch t.X {
					case LITA:
						instr |= 0b0000110001000000
					case LITD:
						instr |= 0b0000001101000000
					case LITM:
						instr |= 0b0001110001000000
					}
				}
			case *ast.BinaryExpr:
				switch t.Op {
				case token.ADD:
					switch {
					case t.X == LITA && t.Y == LIT1:
						instr |= 0b0000110111000000
					case t.X == LITM && t.Y == LIT1:
						instr |= 0b0001110111000000
					case t.X == LITD && t.Y == LIT1:
						instr |= 0b0000011111000000
					case t.X == LITD && t.Y == LITA:
						instr |= 0b0000000010000000
					case t.X == LITD && t.Y == LITM:
						instr |= 0b0001000010000000
					}
				case token.SUB:
					switch {
					case t.X == LITA && t.Y == LIT1:
						instr |= 0b0000110010000000
					case t.X == LITA && t.Y == LITD:
						instr |= 0b0000000111000000
					case t.X == LITM && t.Y == LIT1:
						instr |= 0b0001110010000000
					case t.X == LITM && t.Y == LITD:
						instr |= 0b0001000111000000
					case t.X == LITD && t.Y == LIT1:
						instr |= 0b0000001110000000
					case t.X == LITD && t.Y == LITA:
						instr |= 0b0000010011000000
					case t.X == LITD && t.Y == LITM:
						instr |= 0b0001010011000000
					}
				// AND and OR guaranteed to have X=LITD
				case token.AND:
					switch t.Y {
					//case LITA:
					//instr |= 0b0000000000000000
					case LITM:
						instr |= 0b0001000000000000
					}
				case token.OR:
					switch t.Y {
					case LITA:
						instr |= 0b0000010101000000
					case LITM:
						instr |= 0b0001010101000000
					}
				case token.SHL:
					instr &= 0b1101111111111111
					switch t.X {
					//case LITA:
					//instr |= 0b0000000000000000
					case LITD:
						instr |= 0b0000100000000000
					case LITM:
						instr |= 0b0001000000000000
					}
					bly := t.Y.(*ast.BasicLit)
					n, _ := strconv.ParseInt(bly.Value, 16, 16)
					instr |= uint16(n) << 7
				}
			}
			program = append(program, instr)
		}
	}
	runes := assembly2runes(program)
	return runes, nil
}

var futhark = []rune{'ᚠ', 'ᚢ', 'ᚦ', 'ᚬ', 'ᚱ', 'ᚴ', 'ᚼ', 'ᚾ', 'ᛁ', 'ᛅ', 'ᛋ', 'ᛏ', 'ᛒ', 'ᛘ', 'ᛚ', 'ᛦ'}

func assembly2runes(program []uint16) []string {
	runes := []string{}
	for _, a := range program {
		runes = append(runes, encodeRune(a))
	}
	return runes
}

func encodeRune(a uint16) string {
	return string([]rune{
		futhark[a>>12],
		futhark[a>>8&0xF],
		futhark[a>>4&0xF],
		futhark[a&0xF],
	})
}
