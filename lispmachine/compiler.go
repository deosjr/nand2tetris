package main

import (
    "fmt"

    "github.com/deosjr/whistle/lisp"
)

var symbolTable = map[string]int{
    // TODO: create vm/sys.vm using compiled symbols
    // later we could create an actual symbol table in memory
    // for use in 'read'
    "if": 0,
    "define":1,
    "quote":2,
    "set!":3,
    "lambda":4,
    "begin":5,
    "eval":6,
    "+": 7,
    "-": 8,
    ">": 9,
    "null?": 10,
    "read-char": 11,
    "display": 12,
    "cons": 13,
    "car": 14,
    "cdr": 15,
    "=": 16,
    "write-char": 17,
    "peek-char": 18,
    "error": 19,
    "or": 20,
    "and": 21,
    "not": 22,
    "num->symbol": 23,
    "<<": 24,
    // maybe there should be a separate asm/unsafe.asm
    "&": 25,
    "pair?": 26,
}

// lisp -> vm
func compile2vm(filenames ...string) (string, error) {
    s := "function main\n"
    for _, filename := range filenames {
        sexps, err := lisp.ParseFile(filename)
        if err != nil {
            return "", err
        }
        for _, sexp := range sexps {
            s += "\tpush environment\n"
            out, err := compileSExp(sexp)
            if err != nil {
                return "", err
            }
            s += out
            s += "\tcall eval.eval\n"
            s += "\twrite\n"
        }
    }
    s += "\tpush constant 0\n"
    s += "\treturn\n"
    return s, nil
}

func compileFromString(in string) (string, error) {
    s := "function main\n"
    sexps, err := lisp.Multiparse(in)
    if err != nil {
        return "", err
    }
    for _, sexp := range sexps {
        s += "\tpush environment\n"
        out, err := compileSExp(sexp)
        if err != nil {
            return "", err
        }
        s += out
        s += "\tcall eval.eval\n"
        s += "\twrite\n"
    }
    s += "\tpush constant 0\n"
    s += "\treturn\n"
    return s, nil
}

func compileSExp(sexp lisp.SExpression) (string, error) {
    if sexp.IsPrimitive() {
        n := int(sexp.AsNumber())
        if n < 0 || n >= 0x4000 {
            return "", fmt.Errorf("invalid primitive %d", n)
        }
        n += 0x4000
        return fmt.Sprintf("\tpush constant %d\n", n), nil
    }
    if sexp.IsSymbol() {
        sym := string(sexp.AsSymbol())
        if sym == "#f" {
            return "\tpush constant 0\n", nil
        }
        if len(sym) > 2 && sym[0] == '#' && sym[1] == '\\' {
            return compileChar(sym[2:])
        }
        n := len(symbolTable)
        if i, ok := symbolTable[sym]; ok {
            n = i
        } else {
            symbolTable[sym] = n
        }
        n += 0x6000
        return fmt.Sprintf("\tpush constant %d\n", n), nil
    }
    // guaranteed to be pair!
    s := ""
    list, err := lisp.UnpackConsList(sexp)
    if err != nil {
        return "", err
    }
    for _, e := range list {
        out, err := compileSExp(e)
        if err != nil {
            return "", err
        }
        s += out
    }
    s += "\tpush constant 0\n"   // 0x0 = emptylist
    for range list {
        s += "\tcons\n"
    }
    return s, nil
}

// a symbol such as #\a represents a character in scheme
// we will compile them to their ascii value and represent as primitive
func compileChar(s string) (string, error) {
    n := 0
    switch s {
    case "newline":
        n = 0x0A
    case "space":
        n = 0x20
    case "tab":
        n = 0x09
    case "eof":
        n = 0x1C
    default:
        if len(s) != 1 {
            return "", fmt.Errorf("unknown char #\\%s", s)
        }
        n = int(s[0])
    }
    n += 0x4000
    return fmt.Sprintf("\tpush constant %d\n", n), nil
}

// lisp -> asm
func compile2asm(filenames ...string) (string, error) {
    s := "(MAINMAIN)\n"
    for _, filename := range filenames {
        sexps, err := lisp.ParseFile(filename)
        if err != nil {
            return "", err
        }
        for _, sexp := range sexps {
            // push environment
            s += "\t@ENV\n" 
            s += "\tA=M\n" 
            s += "\tD=M\n" 
            s += "\t@SP\n" 
            s += "\tM=M+1\n" 
            s += "\tA=M-1\n" 
            s += "\tM=D\n" 
            out, err := compileSExp2asm(sexp)
            if err != nil {
                return "", err
            }
            s += out
            // push D onto the stack
            s += "\t@SP\n" 
            s += "\tM=M+1\n" 
            s += "\tA=M-1\n" 
            s += "\tM=D\n" 
            // call eval.eval
            s += "\t@EVALEVAL\n" 
            s += "\tD=A\n" 
            s += "\t@R13\n" 
            s += "\tM=D\n" 
            label := genLabel()
            s += "\t@"+label+"\n" 
            s += "\tD=A\n" 
            s += "\t@R15\n" 
            s += "\tM=D\n" 
            s += "\t@SYSCALL\n" 
            s += "\t0;JMP\n" 
            s += "("+label+")\n" 
            // write
            s += "\t@SP\n" 
            s += "\tAM=M-1\n" 
            s += "\tD=M\n" 
            s += "\t@0x6002\n" 
            s += "\tM=D\n" 
        }
    }
    // push constant 0
    s += "\t@SP\n" 
    s += "\tM=M+1\n" 
    s += "\tA=M-1\n" 
    s += "\tM=0\n" 
    // return
    s += "\t@SYSRETURN\n" 
    s += "\t0;JMP" 
    return s, nil
}

// difference with vm: we no longer push
// this function only sets D to value
func compileSExp2asm(sexp lisp.SExpression) (string, error) {
    if sexp.IsPrimitive() {
        n := int(sexp.AsNumber())
        if n < 0 || n >= 0x4000 {
            return "", fmt.Errorf("invalid primitive %d", n)
        }
        n += 0x4000
        return fmt.Sprintf("\t@%d\n\tD=A\n", n), nil
    }
    if sexp.IsSymbol() {
        sym := string(sexp.AsSymbol())
        if sym == "#f" {
            return "\tD=0\n", nil
        }
        if len(sym) > 2 && sym[0] == '#' && sym[1] == '\\' {
            return compileChar2asm(sym[2:])
        }
        n := len(symbolTable)
        if i, ok := symbolTable[sym]; ok {
            n = i
        } else {
            symbolTable[sym] = n
        }
        n += 0x6000
        return fmt.Sprintf("\t@%d\n\tD=A\n", n), nil
    }
    // guaranteed to be pair!
    s := "\tD=0\n"      // 0x0 = emptylist
    s += "\t@FREE\n"
    s += "\tA=M\n"
    s += "\tSETCDR\n"
    list, err := lisp.UnpackConsList(sexp)
    if err != nil {
        return "", err
    }
    // traverse list in reverse order
    // if not last (ie first in traversal), and pair (meaning we recurse),
    // use the stack to save previous value
    for i:=len(list)-1; i >=0; i-- {
        e := list[i]
        useStack := i < len(list)-1 && e.IsPair()
        if useStack {
            s += "\t@SP\n"
            s += "\tM=M+1\n"
            s += "\tA=M-1\n"
            s += "\tM=D\n"
        }
        out, err := compileSExp2asm(e)
        if err != nil {
            return "", err
        }
        s += out    // D = e
        if useStack {
            s += "\t@FREE\n"
            s += "\tA=M\n"
            s += "\tSETCAR\n"
            s += "\t@SP\n"
            s += "\tAM=M-1\n"
            s += "\tD=M\n"
            s += "\t@FREE\n"
            s += "\tA=M\n"
            s += "\tSETCDR\n"
        } else {
            s += "\t@FREE\n"
            s += "\tA=M\n"
            s += "\tSETCAR\n"
        }
        s += "\t@FREE\n"
        s += "\tD=M\n"
        s += "\tM=D+1\n"
        if i == 0 {
            break
        }
        s += "\t@FREE\n"
        s += "\tA=M\n"
        s += "\tSETCDR\n"
    }
    return s, nil
}

// a symbol such as #\a represents a character in scheme
// we will compile them to their ascii value and represent as primitive
func compileChar2asm(s string) (string, error) {
    n := 0
    switch s {
    case "newline":
        n = 0x0A
    case "space":
        n = 0x20
    case "tab":
        n = 0x09
    case "eof":
        n = 0x1C
    default:
        if len(s) != 1 {
            return "", fmt.Errorf("unknown char #\\%s", s)
        }
        n = int(s[0])
    }
    n += 0x4000
    return fmt.Sprintf("\t@%d\n\tD=A\n", n), nil
}

var generatedLabels = 0

func genLabel() string {
    s := ""
    for _, c := range fmt.Sprintf("%06d", generatedLabels) {
        s += string(c + 17)
    }
    generatedLabels++
    return "CC" + s
}
