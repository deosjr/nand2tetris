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
    "map": 9,
    "write": 10,
}

func compile(filename string) (string, error) {
    sexps, err := lisp.ParseFile(filename)
    if err != nil {
        return "", err
    }
    s := "function main\n"
    for _, sexp := range sexps {
        s += "\tpush environment\n"
        out, err := compileSExp(sexp)
        if err != nil {
            return "", err
        }
        s += out
        s += "\tpush constant 8192\n"   // 0x2000 = emptylist
        s += "\tcons\n"
        s += "\tcons\n"
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
        if n < 0 || n >= 16384 {
            return "", fmt.Errorf("invalid primitive %d", n)
        }
        n += 16384
        return fmt.Sprintf("\tpush constant %d\n", n), nil
    }
    if sexp.IsSymbol() {
        sym := string(sexp.AsSymbol())
        n := len(symbolTable)
        if i, ok := symbolTable[sym]; ok {
            n = i
        } else {
            symbolTable[sym] = n
        }
        n += 24576
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
    s += "\tpush constant 8192\n"   // 0x2000 = emptylist
    for range list {
        s += "\tcons\n"
    }
    return s, nil
}
