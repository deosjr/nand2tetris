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
    "+": 6,
    "-": 7,
    ">": 8,
    "null?": 9,
    "apply": 10,
    "write": 11,
    "cons": 12,
    "car": 13,
    "cdr": 14,
    "=": 15,
}

func compile(filenames ...string) (string, error) {
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
    default:
        if len(s) != 1 {
            return "", fmt.Errorf("unknown char #\\%s", s)
        }
        n = int(s[0])
    }
    n += 0x4000
    return fmt.Sprintf("\tpush constant %d\n", n), nil
}
