package main

// cheating, i.e. writing an assembler in go
// instead of in Hack assembler :)
// also an excuse to play with go/ast
// abusing go/ast def to leverage builtin funcs such as ast.Print

import (
    "bufio"
    "fmt"
    "go/ast"
    "go/token"
    "os"
    "strconv"
    "strings"
    "unicode"
)

// take a .asm file and return machine language code
func Assemble(filename string) ([]uint16, error) {
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
    return assemble(fset, contents, parsed)
}

func assembleFromString(s string) ([]uint16, error) {
    fset := token.NewFileSet()
    parsed, err := parse(fset, "string_input", s)
    if err != nil {
        return nil, err
    }
    s = strings.ReplaceAll(s, "\n", "")
    return assemble(fset, s, parsed)
}

func parse(fset *token.FileSet, filename, contents string) (*ast.File, error) {
    // we're going to add everything into one virtual block statement
    // comments are completely filtered out
    block := &ast.BlockStmt{}
    astfile := &ast.File{
        Decls: []ast.Decl{ &ast.FuncDecl{ Body: block } },
    }
	scanner := bufio.NewScanner(strings.NewReader(contents))
    i := 0
    lineOffsets := []int{}
    for scanner.Scan() {
        lineOffsets = append(lineOffsets, i)
        line := scanner.Text()
        parseLine(block, line, i)
        i += len(line)
    }
    fset.AddFile(filename, 1, i).SetLines(lineOffsets)
    //ast.Print(fset, astfile)
    return astfile, nil
}

var (
    LIT0 = ast.NewIdent("0")
    LIT1 = ast.NewIdent("1")
    LITA = ast.NewIdent("A")
    LITD = ast.NewIdent("D")
    LITM = ast.NewIdent("M")
    LITJGT = ast.NewIdent("JGT")
    LITJEQ = ast.NewIdent("JEQ")
    LITJGE = ast.NewIdent("JGE")
    LITJLT = ast.NewIdent("JLT")
    LITJNE = ast.NewIdent("JNE")
    LITJLE = ast.NewIdent("JLE")
    LITJMP = ast.NewIdent("JMP")
)

// crude mapping into go/ast types
// LABEL  = (VALUE) -> ParenExpr(BasicLit:STRING)
// AINSTR = @VALUE<label, int>  -> BasicLit(STRING/INT)
// CINSTR = DEST=COMP;JMP -> LabeledStmt(Label:JMP, Stmt:AssignStmt(Lhs:DEST, Rhs:Expr(COMP))

func parseLine(block *ast.BlockStmt, line string, offset int) {
    pos := token.Pos(offset) + 1
    trimLeft := strings.TrimLeft(line, " \t")
    pos += token.Pos(len(line)-len(trimLeft))
    split := strings.Split(trimLeft, " ")
    line = strings.TrimSpace(split[0])
    // lispinstructions
    switch line {
    case "SETCAR", "SETCDR", "EQLA", "EQLM", "MCDR", "ISSYMB", "ISPRIM", "EMPTYCDR":
        block.List = append(block.List, &ast.ExprStmt{X: &ast.BasicLit{Kind: token.IMAG, Value:line, ValuePos:pos}})
        return
    }
    // TODO: bad stmt if split[1] exists but is not a comment
    var stmt ast.Stmt
    switch line[0] {
    case '/':
        return
    case '(':
        stmt = parseLabel(line, pos)
    case '@':
        stmt = parseAInstr(line, pos)
    default:
        stmt = parseCInstr(line, pos)
    }
    block.List = append(block.List, stmt)
}

func parseLabel(line string, pos token.Pos) ast.Stmt {
    label := line[1:len(line)-1]
    for _, c := range label {
        if !unicode.IsUpper(c) {
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
    }
    return &ast.ExprStmt{X: &ast.ParenExpr{X:&ast.BasicLit{Kind: token.STRING, Value:label, ValuePos:pos+1}}}
}

func parseAInstr(line string, pos token.Pos) ast.Stmt {
    value := line[1:]
    kind := token.INT
    for _, c := range value {
        if !unicode.IsDigit(c) {
            kind = token.STRING
            break
        }
    }
    return &ast.ExprStmt{X: &ast.BasicLit{Kind: kind, Value:line[1:], ValuePos:pos}}
}

func parseCInstr(line string, pos token.Pos) ast.Stmt {
    var dest []ast.Expr
    var comp ast.Expr
    var jmp *ast.Ident
    split := strings.Split(line, "=")
    switch len(split) {
    case 1:
        break
    case 2:
        for _, c := range split[0] {
            switch c {
            case 'A':
                dest = append(dest, LITA)
            case 'D':
                dest = append(dest, LITD)
            case 'M':
                dest = append(dest, LITM)
            default:
                return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
            }
        }
        line = split[1]
    default:
        return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
    }
    split = strings.Split(line, ";")
    switch len(split) {
    case 1:
        break
    case 2:
        switch split[1] {
        case "JGT":
            jmp = LITJGT
        case "JEQ":
            jmp = LITJEQ
        case "JGE":
            jmp = LITJGE
        case "JLT":
            jmp = LITJLT
        case "JNE":
            jmp = LITJNE
        case "JLE":
            jmp = LITJLE
        case "JMP":
            jmp = LITJMP
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        line = split[0]
    default:
        return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
    }
    switch len(line) {
    case 1:
        switch line {
        case "0":
            comp = LIT0
        case "1":
            comp = LIT1
        case "A":
            comp = LITA
        case "D":
            comp = LITD
        case "M":
            comp = LITM
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
    case 2:
        expr := &ast.UnaryExpr{}
        switch line[0] {
        case '-':
            expr.Op = token.SUB
        case '!':
            expr.Op = token.NOT
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        switch line[1] {
        case '1':
            if expr.Op == token.NOT {
                return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
            }
            expr.X = LIT1
        case 'A':
            expr.X = LITA
        case 'D':
            expr.X = LITD
        case 'M':
            expr.X = LITM
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        comp = expr
    case 3:
        expr := &ast.BinaryExpr{}
        switch line[0] {
        case 'A':
            expr.X = LITA
        case 'D':
            expr.X = LITD
        case 'M':
            expr.X = LITM
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        switch line[1] {
        case '+':
            expr.Op = token.ADD
        case '-':
            expr.Op = token.SUB
        case '&':
            expr.Op = token.AND
        case '|':
            expr.Op = token.OR
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        switch line[2] {
        case '1':
            if expr.Op == token.AND || expr.Op == token.OR {
                return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
            }
            expr.Y = LIT1
        case 'A':
            expr.Y = LITA
        case 'D':
            if expr.Op == token.ADD || expr.Op == token.AND || expr.Op == token.OR {
                return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
            }
            expr.Y = LITD
        case 'M':
            expr.Y = LITM
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        if expr.X == expr.Y {
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        if (expr.Op == token.AND || expr.Op == token.OR) && (expr.X != LITD) {
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        comp = expr
    case 4:
        if line[1:3] != "<<" {
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        expr := &ast.BinaryExpr{Op:token.SHL}
        switch line[0] {
        case 'A':
            expr.X = LITA
        case 'D':
            expr.X = LITD
        case 'M':
            expr.X = LITM
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        switch line[3] {
        case '1','2','3','4','5','6','7','8','9':
            expr.Y = &ast.BasicLit{Kind: token.INT, Value: string(line[3])}
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        comp = expr
    case 5:
        if line[1:3] != "<<" {
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        expr := &ast.BinaryExpr{Op:token.SHL}
        switch line[0] {
        case 'A':
            expr.X = LITA
        case 'D':
            expr.X = LITD
        case 'M':
            expr.X = LITM
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        switch line[3:5] {
        case "10","11","12","13","14","15":
            expr.Y = &ast.BasicLit{Kind: token.INT, Value: line[3:5]}
        default:
            return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
        }
        comp = expr
    default:
        return &ast.BadStmt{From:pos, To:pos+token.Pos(len(line))}
    }
    return &ast.LabeledStmt{Label: jmp, Stmt:&ast.AssignStmt{Lhs:dest, Rhs:[]ast.Expr{comp}}}
}

func assemble(fset *token.FileSet, contents string, parsed *ast.File) ([]uint16, error) {
    // firstpass
    statements := []ast.Stmt{}
    labels := map[string]uint16{
        "SP":       0x0000,
        "LCL":      0x0001,
        "ARG":      0x0002,
        "THIS":     0x0003,
        "THAT":     0x0004,
        "R0":       0x0000,
        "R1":       0x0001,
        "R2":       0x0002,
        "R3":       0x0003,
        "R4":       0x0004,
        "R5":       0x0005,
        "R6":       0x0006,
        "R7":       0x0007,
        "R8":       0x0008,
        "R9":       0x0009,
        "R10":      0x000a,
        "R11":      0x000b,
        "R12":      0x000c,
        "R13":      0x000d,
        "R14":      0x000e,
        "R15":      0x000f,
        "SCREEN":   0x4000,
        "KBD":      0x6000,
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
            case token.IMAG:
                // LISPINSTR
                switch bl.Value {
                case "SETCAR": // same as M=D
                    program = append(program, 0b1111001100001000)
                case "SETCDR":
                    program = append(program, 0b1010111111000000)
                case "EQLA":
                    program = append(program, 0b0)
                case "EQLM":
                    program = append(program, 0b0)
                case "MCDR":
                    program = append(program, 0b1000011111010000)
                case "ISSYMB":
                    program = append(program, 0b0)
                case "ISPRIM":
                    program = append(program, 0b0)
                case "EMPRYCDR":
                    program = append(program, 0b0)
                default:
                    posFrom := fset.Position(bl.ValuePos)
                    return nil, fmt.Errorf("%s: invalid instr: %s", posFrom.String(), contents[bl.ValuePos:bl.ValuePos+token.Pos(len(bl.Value))])
                }
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
    return program, nil
}
