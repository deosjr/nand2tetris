package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

// changes to the nand2tetris vm:
// all functions take exactly one arg
// LCL/THIS/THAT no longer exist
// local/this/that also don't exist
// instead, we can refer to register R0-15 directly
// these are assumed to contain garbage!
// NOTE: adding local back, will be sparingly used
// refers to first n places on working stack
// those need to be explicitly pushed at start of function!
// some are used by SYS and should not be used by others
// SHIFT is gone
// bunch of lisp specific instructions added, including EVAL

type vmTranslator struct {
    generatedLabels int
    b *strings.Builder
    fn string
}

// take a set of .vm files and return a single .asm file
func Translate(filenames []string, compiledMain string) (string, error) {
    t := &vmTranslator{}
    out := preamble(filenames)
    // translating sys.vm first allows us to drop into sys.init at start
    o, err := t.translateFiles("vm/sys.vm")
    if err != nil {
        return "", err
    }
    out += o
    out += builtins
    o, err = t.vm2asm([]string{"vm/main.vm"}, []string{compiledMain})
    if err != nil {
        return "", err
    }
    out += o
    o, err = t.translateFiles(filenames...)
    if err != nil {
        return "", err
    }
    out += o
    return out, nil
}

func (t *vmTranslator) vm2asm(filenames, vmStrings []string) (string, error) {
    if len(filenames) != len(vmStrings) {
        return "", fmt.Errorf("filenames and vmStrings doesnt match")
    }
    var out string
    for i, f := range filenames {
        vm := vmStrings[i]
        asm, err := t.translateString(f, vm)
        if err != nil {
            return "", err
        }
        out += asm
    }
    return out, nil
}

func (t *vmTranslator) translateFiles(filenames ...string) (string, error) {
    out := ""
    for _, f := range filenames {
        o, err := t.translate(f)
        if err != nil {
            return "", err
        }
        out += o
    }
    return out, nil
}

func (t *vmTranslator) translate(filename string) (string, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }
    contents := string(data)
    return t.translateString(filename, contents)
}

func (t *vmTranslator) translateString(filename, contents string) (string, error) {
    t.b = &strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(contents))
    fnsplit := strings.Split(strings.Split(filename, ".")[0], "/")
    t.fn = strings.ToUpper(fnsplit[len(fnsplit)-1])
    for scanner.Scan() {
        line := scanner.Text()
        if err := t.translateLine(line); err != nil {
            return "", err
        }
    }
    return t.b.String(), nil
}

func (t *vmTranslator) translateLine(line string) error {
    line = strings.TrimSpace(line)
    if strings.HasPrefix(line, "//") {
        return nil
    }
    split := strings.Fields(line)
    if len(split) == 0 {
        return nil
    }
    switch split[0] {
    case "label":
        return t.translateLabel(split[1:])
    case "goto":
        return t.translateGoto(split[1:])
    case "if-goto":
        return t.translateIfGoto(split[1:])
    case "function":
        return t.translateFunction(split[1:])
    case "call":
        return t.translateCall(split[1:])
    case "return":
        return t.translateReturn(split[1:])
    case "push":
        return t.translatePush(split[1:])
    case "pop":
        return t.translatePop(split[1:])
    // TODO: neg
    case "eq":
        return t.translateEq(split[1:])
    case "gt":
        return t.translateGt(split[1:])
    case "lt":
        return t.translateLt(split[1:])
    case "and":
        return t.translateAnd(split[1:])
    case "or":
        return t.translateOr(split[1:])
    case "not":
        return t.translateNot(split[1:])
    case "add":
        return t.translateAdd(split[1:])
    case "sub":
        return t.translateSub(split[1:])
    case "read":
        return t.translateRead(split[1:])
    case "write":
        return t.translateWrite(split[1:])
    // compiler optimisations / extentions to the vm language spec as shortcuts
    case "plus1":
        return t.translatePlus1(split[1:])
    case "minus1":
        return t.translatePlus1(split[1:])
    // lisp 
    case "cons":
        return t.translateCons(split[1:])
    case "car":
        return t.translateCar(split[1:])
    case "cdr":
        return t.translateCdr(split[1:])
    case "cadr":
        return t.translateCadr(split[1:])
    case "caddr":
        return t.translateCaddr(split[1:])
    case "is-procedure":
        return t.translateIsProcedure(split[1:])
    case "is-builtin":
        return t.translateIsBuiltin(split[1:])
    case "is-special":
        return t.translateIsSpecial(split[1:])
    case "is-symbol":
        return t.translateIsSymbol(split[1:])
    case "is-primitive":
        return t.translateIsPrimitive(split[1:])
    case "is-emptylist":
        return t.translateIsEmptyList(split[1:])
    case "equal":
        return t.translateEqual(split[1:])
    case "copy-pointer":
        return t.translateCopyPointer(split[1:])
    // syntactic sugar
    case "builtin":
        return t.translateBuiltin(split[1:])
    case "call-builtin":
        return t.translateCallBuiltin(split[1:])
    case "userdefined":
        return t.translateUserdefined(split[1:])
    case "special":
        return t.translateSpecial(split[1:])
    case "symbol":
        return t.translateSymbol(split[1:])
    default:
        return fmt.Errorf("syntax error: %s", line)
    }
    return nil
}

func (t *vmTranslator) translateLabel(split []string) error {
    if len(split) == 0 {
        return fmt.Errorf("syntax error: not enough arguments to label")
    }
    label := split[0]
    if len(split) > 1 && !strings.HasPrefix(split[1], "//") {
        return fmt.Errorf("syntax error: label %v", split)
    }
    label = strings.ToUpper(label)
    t.b.WriteString(fmt.Sprintf("(%s%s)\n", t.fn, label))
    return nil
}

func (t *vmTranslator) translateGoto(split []string) error {
    if len(split) == 0 {
        return fmt.Errorf("syntax error: not enough arguments to goto")
    }
    label := split[0]
    if len(split) > 1 && !strings.HasPrefix(split[1], "//") {
        return fmt.Errorf("syntax error: goto %v", split)
    }
    label = strings.ToUpper(label)
    if strings.Contains(label, ".") {
        label = strings.Replace(label, ".", "", -1)
    } else {
        label = t.fn + label
    }
    t.b.WriteString(fmt.Sprintf("\t@%s\n\t0;JMP\n", label))
    return nil
}

func (t *vmTranslator) translateIfGoto(split []string) error {
    if len(split) == 0 {
        return fmt.Errorf("syntax error: not enough arguments to if-goto")
    }
    label := split[0]
    if len(split) > 1 && !strings.HasPrefix(split[1], "//") {
        return fmt.Errorf("syntax error: if-goto %v", split)
    }
    label = strings.ToUpper(label)
    if strings.Contains(label, ".") {
        label = strings.Replace(label, ".", "", -1)
    } else {
        label = t.fn + label
    }
    lines := strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "@" + label,
        "!D;JEQ\n",
    }, "\n\t")
    t.b.WriteString(lines)
    return nil
}

func (t *vmTranslator) translateFunction(split []string) error {
    if len(split) < 1 {
        return fmt.Errorf("syntax error: not enough arguments to function")
    }
    label := split[0]
    if len(split) > 1 && !strings.HasPrefix(split[1], "//") {
        return fmt.Errorf("syntax error: function %v", split)
    }
    label = t.fn + strings.ToUpper(label)
    t.b.WriteString(fmt.Sprintf("(%s)\n", label))
    return nil
}

func (t *vmTranslator) translateCall(split []string) error {
    if len(split) < 1 {
        return fmt.Errorf("syntax error: not enough arguments to call")
    }
    funcname := split[0]
    if len(split) > 1 && !strings.HasPrefix(split[1], "//") {
        return fmt.Errorf("syntax error: call %v", split)
    }
    // TODO: currently always need to specify full import i.e. call mult.mult
    funcname = strings.ToUpper(strings.Replace(funcname, ".", "", -1))
    returnlabel := t.genLabel()
    lines := strings.Join([]string{
        "@" + funcname,
        "D=A",
        "@R13",
        "M=D",
        "@" + returnlabel,
        "D=A",
        "@R15",
        "M=D",
        "@SYSCALL",
        "0;JMP",
    }, "\n\t")
    t.b.WriteString(fmt.Sprintf("\t%s\n(%s)\n", lines, returnlabel))
    return nil
}

func (t *vmTranslator) translateReturn(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: return %v", split)
    }
    t.b.WriteString("\t@SYSRETURN\n\t0;JMP\n")
    return nil
}

func (t *vmTranslator) translatePush(split []string) error {
    if len(split) < 1 {
        return fmt.Errorf("syntax error: not enough arguments to push")
    }
    if len(split) > 2 && !(strings.HasPrefix(split[1], "//") || strings.HasPrefix(split[2], "//")) {
        return fmt.Errorf("syntax error: push %v", split)
    }
    segment := split[0]
    switch segment {
    case "label":
        label := strings.ToUpper(strings.Replace(split[1], ".", "", -1))
        t.b.WriteString(fmt.Sprintf("\t@%s\n\tD=A\n", label))
    case "constant":
        n, err := strconv.ParseInt(split[1], 0, 0)
        if err != nil {
            return err
        }
        switch n {
        case 0, 1, -1:
            t.b.WriteString(strings.Join([]string{
                "\t@SP",
                "M=M+1",
                "A=M-1",
                fmt.Sprintf("M=%d\n", n),
            }, "\n\t"))
            return nil
        default:
            t.b.WriteString(fmt.Sprintf("\t@%d\n\tD=A\n", n))
        }
    case "local":
        n, err := strconv.ParseInt(split[1], 0, 0)
        if err != nil {
            return err
        }
        t.b.WriteString(strings.Join([]string{
            fmt.Sprintf("\t@%d", n+4),
            "D=A",
            "@ARG",
            "A=D+M",
            "D=M\n",
        }, "\n\t"))
    case "argument":
        t.b.WriteString(strings.Join([]string{
            "\t@ARG",
             "A=M",
             "D=M\n",
        }, "\n\t"))
    case "environment":
        t.b.WriteString(strings.Join([]string{
            "\t@ENV",
             "D=M\n",
        }, "\n\t"))
    case "r":
        n, err := strconv.ParseInt(split[1], 0, 0)
        if err != nil {
            return err
        }
        t.b.WriteString(strings.Join([]string{
            "\t@R"+fmt.Sprintf("%d", n),
             "D=M\n",
        }, "\n\t"))
    case "static":
        varname := strings.ToLower(t.fn) + split[1]
        t.b.WriteString(strings.Join([]string{
            "\t@"+varname,
            "D=M\n",
        }, "\n\t"))
    default:
        return fmt.Errorf("syntax error: push %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "M=M+1",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translatePop(split []string) error {
    if len(split) < 1 {
        return fmt.Errorf("syntax error: not enough arguments to pop")
    }
    segment := split[0]
    if segment == "local" {
        n, err := strconv.Atoi(split[1])
        if err != nil {
            return err
        }
        t.b.WriteString(strings.Join([]string{
         "\t@ARG",
         "D=M",
         "@" + fmt.Sprintf("%d", n+4),
         "D=D+A",
         "@R14",
         "M=D",
         "@SP",
         "AM=M-1",
         "D=M",
         "@R14",
         "A=M",
         "M=D\n",
        },  "\n\t"))
        return nil
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M\n",
    }, "\n\t"))
    switch segment {
    case "argument":
        t.b.WriteString(strings.Join([]string{
            "\t@ARG",
            "A=M",
            "M=D\n",
        }, "\n\t"))
    case "environment":
        t.b.WriteString(strings.Join([]string{
            "\t@ENV",
            "M=D\n",
        }, "\n\t"))
    case "static":
        varname := strings.ToLower(t.fn) + split[1]
        t.b.WriteString(strings.Join([]string{
            "\t@"+varname,
            "M=D\n",
        }, "\n\t"))
    case "r":
        n, err := strconv.Atoi(split[1])
        if err != nil {
            return err
        }
        t.b.WriteString(strings.Join([]string{
            "\t@R"+fmt.Sprintf("%d", n),
             "M=D\n",
        }, "\n\t"))
    default:
        return fmt.Errorf("syntax error: pop %v", split)
    }
    return nil
}

func (t *vmTranslator) translateEq(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: eq %v", split)
    }
    falselabel := t.genLabel()
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "A=A-1",
        "D=M-D",
        "M=0", // if D=0 set M=0xffff otherwise set M=0x0000
        "@"+falselabel,
        "D;JNE",
        "@SP", // TRUE
        "A=M-1",
        "M=!M\n",
    }, "\n\t"))
    t.b.WriteString("("+falselabel+")\n")
    return nil
}

func (t *vmTranslator) translateGt(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: gt %v", split)
    }
    falselabel := t.genLabel()
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "A=A-1",
        "D=M-D",
        "M=0", // if D>0 set M=0xffff otherwise set M=0x0000
        "@"+falselabel,
        "D;JLE",
        "@SP", // TRUE
        "A=M-1",
        "M=!M\n",
    }, "\n\t"))
    t.b.WriteString("("+falselabel+")\n")
    return nil
}

func (t *vmTranslator) translateLt(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: lt %v", split)
    }
    falselabel := t.genLabel()
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "A=A-1",
        "D=M-D",
        "M=0", // if D<0 set M=0xffff otherwise set M=0x0000
        "@"+falselabel,
        "D;JGE",
        "@SP", // TRUE
        "A=M-1",
        "M=!M\n",
    }, "\n\t"))
    t.b.WriteString("("+falselabel+")\n")
    return nil
}

func (t *vmTranslator) translateAnd(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: and %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "A=A-1",
        "M=D&M\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateOr(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: or %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "A=A-1",
        "M=D|M\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateNot(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: not %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "M=!M\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateAdd(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: add %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "A=A-1",
        "M=D+M\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateSub(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: sub %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "A=A-1",
        "M=M-D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateRead(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: read %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@0x6001",
        "DM=M",
        "@SP",
        "M=M+1",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateWrite(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: write %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "@0x6002",
        "M=D\n",
    }, "\n\t"))
    return nil
}

// assumes 2 values pushed two stack: first CAR and then CDR
// consumes both, writes a new cons cell, then returns pointer to stack
// SETCDR goes first, because it overwrites CAR as well!
func (t *vmTranslator) translateCons(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: cons %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "@FREE",
        "A=M",
        "SETCDR",
        "@SP",
        "A=M-1",
        "D=M",
        "@FREE",
        "A=M",
        "SETCAR",
        "@FREE",
        "D=M",
        "M=D+1",
        "@SP",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

// follows a pointer
func (t *vmTranslator) translateCar(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: car %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "A=M",
        "MCAR",
        "@SP",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateCdr(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: cdr %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "A=M",
        "MCDR",
        "@SP",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateCadr(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: cadr %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "A=M",
        "MCDR",
        "A=D",
        "MCAR",
        "@SP",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateCaddr(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: caddr %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "A=M",
        "MCDR",
        "A=D",
        "MCDR",
        "A=D",
        "MCAR",
        "@SP",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateIsSymbol(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: is-symbol %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "ISSYMB",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateIsPrimitive(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: is-primitive %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "ISPRIM",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateIsProcedure(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: is-procedure %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "ISPROC",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateIsBuiltin(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: is-builtin %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "ISBUILTIN",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateIsSpecial(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: is-special %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "ISSPECIAL",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateIsEmptyList(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: is-emptylist %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        "ISEMPTY",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateEqual(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: equal %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "A=A-1",
        "EQLM",
        "M=D\n",
    }, "\n\t"))
    return nil
}

// since procedures go beyond the 15-bit limit of constants we can push
// this function simply adds 0xa000 to the previous value on stack
// NOTE: since we use OR, if we try to create a builtin larger than 0x1fff,
// this wil fail in unexpected ways!
func (t *vmTranslator) translateBuiltin(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: builtin %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@0x7fff",
        "D=A",
        "@0x2001",
        "D=D+A",
        "@SP",
        "A=M-1",
        "M=D|M\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateCallBuiltin(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: call-builtin %v", split)
    }
    returnlabel := t.genLabel()
    lines := strings.Join([]string{
        "\t@" + returnlabel,
        "D=A",
        "@R15",
        "M=D",
        "@SP",
        "AM=M-1",
        "D=M",
        "@0x1fff",
        "A=D&A",    // mask off first three bits (TODO, could check them first)
        "0;JMP",
    }, "\n\t")
    t.b.WriteString(fmt.Sprintf("\t%s\n(%s)\n", lines, returnlabel))
    return nil
}

// same warning as for Builtin
func (t *vmTranslator) translateUserdefined(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: userdefined %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@0x7fff",
        "D=A",
        "@0x0001",
        "D=D+A",
        "@SP",
        "A=M-1",
        "M=D|M\n",
    }, "\n\t"))
    return nil
}

// same warning as for Builtin
func (t *vmTranslator) translateSpecial(split []string) error {
    if len(split) < 1 {
        return fmt.Errorf("syntax error: not enough arguments to special")
    }
    if len(split) > 1 && !strings.HasPrefix(split[1], "//") {
        return fmt.Errorf("syntax error: special %v", split)
    }
    n, err := strconv.ParseInt(split[0], 0, 0)
    if err != nil {
        return err
    }
    t.b.WriteString(strings.Join([]string{
        "\t@"+fmt.Sprintf("%d", n),
        "D=A",
        "@0x7fff",
        "D=D+A",
        "@0x6001",
        "D=D+A",
        "@SP",
        "M=M+1",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translateSymbol(split []string) error {
    if len(split) < 1 {
        return fmt.Errorf("syntax error: not enough arguments to symbol")
    }
    if len(split) > 1 && !strings.HasPrefix(split[1], "//") {
        return fmt.Errorf("syntax error: symbol %v", split)
    }
    n, err := strconv.ParseInt(split[0], 0, 0)
    if err != nil {
        return err
    }
    n += 24576 // symbol prefix 011
    t.b.WriteString(strings.Join([]string{
        "\t@"+fmt.Sprintf("%d", n),
        "D=A",
        "@SP",
        "M=M+1",
        "A=M-1",
        "M=D\n",
    }, "\n\t"))
    return nil
}

// used in eval define to modify ENV; going against general immutability of pointers
// assumes two values on stack: first destination pointer then source pointer
// sole user of R7 and R8 registers
func (t *vmTranslator) translateCopyPointer(split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: copy-pointer %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M",
        "@R7",  // dest
        "M=D",
        "@SP",
        "AM=M-1",
        "D=M",
        "@R8",  // src
        "AM=D",
        "MCDR",
        "@R7",
        "A=M",
        "SETCDR",
        "@R8",
        "A=M",
        "MCAR",
        "@R7",
        "A=M",
        "SETCAR\n",
    }, "\n\t"))
    return nil
}

func (t *vmTranslator) translatePlus1(split []string) error {
    if len(split) < 2 {
        return fmt.Errorf("syntax error: not enough arguments to plus1")
    }
    if len(split) > 2 && !strings.HasPrefix(split[2], "//") {
        return fmt.Errorf("syntax error: plus1 %v", split)
    }
    segment := split[0]
    n, err := strconv.ParseInt(split[1], 0, 0)
    if err != nil && segment != "static" {
        return err
    }
    switch segment {
    case "argument":
        t.b.WriteString(strings.Join([]string{
            "\t@ARG",
            "A=M\n",
        }, "\n\t"))
    case "pointer":
        t.b.WriteString(fmt.Sprintf("\t@%d\n", n+3))
    case "temp":
        t.b.WriteString(fmt.Sprintf("\t@%d\n", n+5))
    case "static":
        varname := strings.ToLower(t.fn) + split[1]
        t.b.WriteString(fmt.Sprintf("\t@%s\n", varname))
    default:
        return fmt.Errorf("syntax error: plus1 %v", split)
    }
    t.b.WriteString("\tM=M+1\n")
    return nil
}

func (t *vmTranslator) translateMin1(split []string) error {
    if len(split) < 2 {
        return fmt.Errorf("syntax error: not enough arguments to min1")
    }
    if len(split) > 2 && !strings.HasPrefix(split[2], "//") {
        return fmt.Errorf("syntax error: min1 %v", split)
    }
    segment := split[0]
    n, err := strconv.ParseInt(split[1], 0, 0)
    if err != nil && segment != "static" {
        return err
    }
    switch segment {
    case "argument":
        t.b.WriteString(strings.Join([]string{
            "\t@ARG",
            "A=M\n",
        }, "\n\t"))
    case "pointer":
        t.b.WriteString(fmt.Sprintf("\t@%d\n", n+3))
    case "temp":
        t.b.WriteString(fmt.Sprintf("\t@%d\n", n+5))
    case "static":
        varname := strings.ToLower(t.fn) + split[1]
        t.b.WriteString(fmt.Sprintf("\t@%s\n", varname))
    default:
        return fmt.Errorf("syntax error: min1 %v", split)
    }
    t.b.WriteString("\tM=M-1\n")
    return nil
}

func (t *vmTranslator) genLabel() string {
    s := ""
    for _, c := range fmt.Sprintf("%06d", t.generatedLabels) {
        s += string(c + 17)
    }
    t.generatedLabels++
    return "XX" + s
}

func preamble(filenames []string) string {
    return fmt.Sprintf(`// VM translation of %s
@256    // TODO: we can probably start stack way earlier
D=A
@SP
M=D
@2048
D=A
@FREE
M=D
@SYSINIT
0;JMP
(SYSEND)
@SYSEND // address 10
0;JMP
`, strings.Join(filenames, ","))
}

const builtins = `// BUILTIN SYS FUNCTIONS
(SYSCALL)
    // push return-address
    @R15       // RET
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // push ENV    // save ENV of calling function
    @ENV
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // push ARG    // save ARG of calling function
    @ARG
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // ARG=SP-4  // reposition ARG
    @SP
    D=M
    @4
    D=D-A
    @ARG
    M=D
    // goto f      // transfer control
    @R13    // FUNC
    A=M
    0;JMP
(SYSRETURN)
    // FRAME = ARG
    @ARG
    D=M
    @R14    // FRAME
    // RET = *(FRAME+1)
    AM=D+1
    D=M
    @R15    // RET
    M=D
    // *ARG = pop()
    @SP
    AM=M-1
    D=M
    @ARG
    A=M
    M=D
    // SP = ARG+1 = FRAME
    @R14
    D=M
    @SP
    M=D
    // ENV = *(FRAME+2)
    @R14
    AM=M+1
    D=M
    @ENV
    M=D
    // ARG = *(FRAME+3)
    @R14
    AM=M+1
    D=M
    @ARG
    M=D
    // goto RET
    @R15
    A=M
    0;JMP
(SYSSTACKOVERFLOW)
    @0x0666
    D=A
    @0x6002
    M=D
    @SYSEND
    0;JMP
(SYSHEAPOVERFLOW)
    @0x0667
    D=A
    @0x6002
    M=D
    @SYSEND
    0;JMP
// BUILTIN FUNCTIONS
// RULES: may use R5-R10 as local vars
// may not call into any other function
// SP-1 contains ARG, always a list
// R15 contains the return address
(BUILTINADD)
    @SP
    A=M-1
    A=M
    MCAR
    @R5     // use R5 as dump var
    M=D
    @SP
    A=M-1
    A=M
    MCDR
    A=D
    MCAR
    @R5
    D=D+M
    @0x1fff
    D=D&A
    @0x4000
    D=D|A
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINSUB)
    @SP
    A=M-1
    A=M
    MCAR
    @R5     // use R5 as dump var
    M=D
    @SP
    A=M-1
    A=M
    MCDR
    A=D
    MCAR
    @R5
    D=M-D
    @0x1fff
    D=D&A
    @0x4000
    D=D|A
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINEQ)
    @SP
    A=M-1
    A=M
    MCAR
    @R5     // use R5 as dump var
    M=D
    @SP
    A=M-1
    A=M
    MCDR
    A=D
    MCAR
    @R5
    D=M-D
    M=0
    @BUILTINEQFALSE
    D;JNE
    @R5
    M=!M
(BUILTINEQFALSE)
    @R5
    D=M
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINGT)
    @SP
    A=M-1
    A=M
    MCAR
    @R5     // use R5 as dump var
    M=D
    @SP
    A=M-1
    A=M
    MCDR
    A=D
    MCAR
    @R5
    D=M-D
    M=0
    @BUILTINGTFALSE
    D;JLE
    @R5
    M=!M
(BUILTINGTFALSE)
    @R5
    D=M
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINISNULL)
    @SP
    A=M-1
    A=M
    MCAR
    ISEMPTY
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINCAR)
    @SP
    A=M-1
    A=M
    MCAR
    A=D
    MCAR
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINCDR)
    @SP
    A=M-1
    A=M
    MCAR
    A=D
    MCDR
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINCONS)
    @SP
    A=M-1
    A=M
    MCDR
    A=D
    MCAR
    @FREE
    A=M
    SETCDR
    @SP
    A=M-1
    A=M
    MCAR
    @FREE
    A=M
    SETCAR
    @FREE
    D=M
    M=D+1
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINWRITE)
    @SP
    A=M-1
    A=M
    MCAR
    @0x6002
    M=D
    @SP
    A=M-1
    M=0
    @R15
    A=M
    0;JMP
(BUILTINASSQ)
    @SP
    A=M-1
    A=M
    MCAR
    @R5
    M=D         // R5 = assoclist
    @SP
    A=M-1
    A=M
    MCDR
    A=D
    MCAR
    @R6
    M=D         // R6 = key
(BUILTINASSQSTART)
    @R5
    A=M
    MCAR
    A=D
    MCAR
    @R6
    EQLM
    @BUILTINASSQFOUND
    D;JNE
    @R5
    A=M
    EMPTYCDR
    @BUILTINASSQFAIL
    D;JNE
    @R5
    A=M
    MCDR
    @R5
    M=D
    @BUILTINASSQSTART
    0;JMP
(BUILTINASSQFOUND)
    @R5
    A=M
    MCAR
    A=D
    MCDR
    @SP
    A=M-1
    M=D
    @R15
    A=M
    0;JMP
(BUILTINASSQFAIL)
    @SP
    A=M-1
    M=0
    @R15
    A=M
    0;JMP
`   // note: this newline is important!
