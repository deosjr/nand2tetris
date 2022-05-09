package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

type vmTranslator struct {
    generatedLabels int
    b *strings.Builder
    fn string
}

// take a set of .vm files and return a single .asm file
func Translate(filenames []string) (string, error) {
    t := &vmTranslator{}
    out := preamble(filenames)
    // translating sys.vm first allows us to drop into sys.init at start
    o, err := t.translateFiles("vm/sys.vm", "vm/memory.vm")
    if err != nil {
        return "", err
    }
    out += o
    o, err = t.translateFiles(filenames...)
    if err != nil {
        return "", err
    }
    out += o
    out += builtins
    return out, nil
}

func vm2asm(filenames, vmStrings []string) (string, error) {
    if len(filenames) != len(vmStrings) {
        return "", fmt.Errorf("filenames and vmStrings doesnt match")
    }
    t := &vmTranslator{}
    out := preamble(filenames)
    o, err := t.translateFiles("vm/sys.vm")
    if err != nil {
        return "", err
    }
    out += o
    for i, f := range filenames {
        vm := vmStrings[i]
        asm, err := t.translateString(f, vm)
        if err != nil {
            return "", err
        }
        out += asm
    }
    out += builtins
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
    // TODO: shifts have been added as separate unary operators
    // ie 16 different ops for << 0 through << 15
    case "lshift1":
        return t.translateShift(1, split[1:])
    case "lshift3":
        return t.translateShift(3, split[1:])
    case "lshift5":
        return t.translateShift(5, split[1:])
    case "lshift8":
        return t.translateShift(8, split[1:])
    case "lshift15":
        return t.translateShift(15, split[1:])
    // compiler optimisations / extentions to the vm language spec as shortcuts
    case "plus1":
        return t.translatePlus1(split[1:])
    case "minus1":
        return t.translatePlus1(split[1:])
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
    label = t.fn + strings.ToUpper(label)
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
    label = t.fn + strings.ToUpper(label)
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
    if len(split) < 2 {
        return fmt.Errorf("syntax error: not enough arguments to function")
    }
    label := split[0]
    numlcl, err := strconv.Atoi(split[1])
    if err != nil {
        return err
    }
    if len(split) > 2 && !strings.HasPrefix(split[2], "//") {
        return fmt.Errorf("syntax error: function %v", split)
    }
    label = t.fn + strings.ToUpper(label)
    t.b.WriteString(fmt.Sprintf("(%s)\n", label))
    if numlcl == 0 {
        return nil
    }
    lclreturn := t.genLabel()
    lines := strings.Join([]string{
        fmt.Sprintf("@%d", numlcl),
        "D=A",
        "@R14",
        "M=D",
        "@" + lclreturn,
        "D=A",
        "@R15",
        "M=D",
        "@SYSPUSHLCL",
        "0;JMP",
    }, "\n\t")
    t.b.WriteString(fmt.Sprintf("\t%s\n(%s)\n", lines, lclreturn))
    return nil
}

func (t *vmTranslator) translateCall(split []string) error {
    if len(split) < 2 {
        return fmt.Errorf("syntax error: not enough arguments to call")
    }
    funcname := split[0]
    numargs, err := strconv.Atoi(split[1])
    if err != nil {
        return err
    }
    if len(split) > 2 && !strings.HasPrefix(split[2], "//") {
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
        fmt.Sprintf("@%d", numargs),
        "D=A",
        "@R14",
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
    if len(split) < 2 {
        return fmt.Errorf("syntax error: not enough arguments to push")
    }
    if len(split) > 2 && !strings.HasPrefix(split[2], "//") {
        return fmt.Errorf("syntax error: push %v", split)
    }
    segment := split[0]
    n, err := strconv.ParseInt(split[1], 0, 0)
    if err != nil && segment != "static" {
        return err
    }
    switch segment {
    case "constant":
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
    case "local", "argument", "this", "that":
        var varname string
        switch segment {
        case "local":
            varname = "LCL"
        case "argument":
            varname = "ARG"
        case "this":
            varname = "THIS"
        case "that":
            varname = "THAT"
        }
        if (segment == "this" || segment == "that") && n != 0 {
            return fmt.Errorf("syntax error: push %v", split)
        }
        switch n {
        case 0:
            t.b.WriteString(strings.Join([]string{
                "\t@"+varname,
                "A=M",
                "D=M\n",
            }, "\n\t"))
        case 1:
            t.b.WriteString(strings.Join([]string{
                "\t@"+varname,
                "A=M+1",
                "D=M\n",
            }, "\n\t"))
        default:
            t.b.WriteString(strings.Join([]string{
                fmt.Sprintf("\t@%d", n),
                "D=A",
                "@"+varname,
                "A=D+M",
                "D=M\n",
            }, "\n\t"))
        }
    case "temp":
        t.b.WriteString(strings.Join([]string{
            fmt.Sprintf("\t@%d", n+5),
            "D=M\n",
        }, "\n\t"))
    case "static":
        varname := strings.ToLower(t.fn) + split[1]
        t.b.WriteString(strings.Join([]string{
            "\t@"+varname,
            "D=M\n",
        }, "\n\t"))
    case "pointer":
        varname := "THIS"
        if n == 1 {
            varname = "THAT"
        } else if n != 0 {
            return fmt.Errorf("syntax error: push %v", split)
        }
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
    if len(split) < 2 {
        return fmt.Errorf("syntax error: not enough arguments to pop")
    }
    if len(split) > 2 && !strings.HasPrefix(split[2], "//") {
        return fmt.Errorf("syntax error: pop %v", split)
    }
    segment := split[0]
    n, err := strconv.Atoi(split[1])
    if err != nil && segment != "static" {
        return err
    }
    if segment == "local" || segment == "argument" {
        var varname string
        if segment == "local" {
            varname = "LCL"
        } else {
            varname = "ARG"
        }
        switch n {
        case 0:
            t.b.WriteString(strings.Join([]string{
             "@SP",
             "AM=M-1",
             "D=M",
             "\t@" + varname,
             "A=M",
             "M=D\n",
            },  "\n\t"))
        case 1:
            t.b.WriteString(strings.Join([]string{
             "@SP",
             "AM=M-1",
             "D=M",
             "\t@" + varname,
             "A=M+1",
             "M=D\n",
            },  "\n\t"))
        default:
            t.b.WriteString(strings.Join([]string{
             "\t@" + varname,
             "D=M",
             "@" + fmt.Sprintf("%d", n),
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
        }
        return nil
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "AM=M-1",
        "D=M\n",
    }, "\n\t"))
    switch segment {
    case "this", "that":
        var varname string
        switch segment {
        case "this":
            varname = "THIS"
        case "that":
            varname = "THAT"
        }
        if n != 0 {
            return fmt.Errorf("syntax error: pop %v", split)
        }
        t.b.WriteString(strings.Join([]string{
            "\t@"+varname,
            "A=M",
            "M=D\n",
        }, "\n\t"))
    case "static":
        varname := strings.ToLower(t.fn) + split[1]
        t.b.WriteString(strings.Join([]string{
            "\t@"+varname,
            "M=D\n",
        }, "\n\t"))
    case "temp":
        t.b.WriteString(strings.Join([]string{
            fmt.Sprintf("\t@%d", n+5),
            "M=D\n",
        }, "\n\t"))
    case "pointer":
        varname := "THIS"
        if n == 1 {
            varname = "THAT"
        } else if n != 0 {
            return fmt.Errorf("syntax error: pop %v", split)
        }
        t.b.WriteString(strings.Join([]string{
            "\t@"+varname,
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

func (t *vmTranslator) translateShift(n int, split []string) error {
    if len(split) > 0 && !strings.HasPrefix(split[0], "//") {
        return fmt.Errorf("syntax error: lshift %v", split)
    }
    t.b.WriteString(strings.Join([]string{
        "\t@SP",
        "A=M-1",
        fmt.Sprintf("M=M<<%d\n", n),
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
    case "local", "argument", "this", "that":
        var varname string
        switch segment {
        case "local":
            varname = "LCL"
        case "argument":
            varname = "ARG"
        case "this":
            varname = "THIS"
        case "that":
            varname = "THAT"
        }
        if (segment == "this" || segment == "that") && n != 0 {
            return fmt.Errorf("syntax error: plus1 %v", split)
        }
        switch n {
        case 0:
            t.b.WriteString(strings.Join([]string{
                "\t@"+varname,
                "A=M\n",
            }, "\n\t"))
        case 1:
            t.b.WriteString(strings.Join([]string{
                "\t@"+varname,
                "A=M+1\n",
            }, "\n\t"))
        default:
            t.b.WriteString(strings.Join([]string{
                fmt.Sprintf("\t@%d", n),
                "D=A",
                "@"+varname,
                "A=D+M\n",
            }, "\n\t"))
        }
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
    case "local", "argument", "this", "that":
        var varname string
        switch segment {
        case "local":
            varname = "LCL"
        case "argument":
            varname = "ARG"
        case "this":
            varname = "THIS"
        case "that":
            varname = "THAT"
        }
        if (segment == "this" || segment == "that") && n != 0 {
            return fmt.Errorf("syntax error: min1 %v", split)
        }
        switch n {
        case 0:
            t.b.WriteString(strings.Join([]string{
                "\t@"+varname,
                "A=M\n",
            }, "\n\t"))
        case 1:
            t.b.WriteString(strings.Join([]string{
                "\t@"+varname,
                "A=M+1\n",
            }, "\n\t"))
        default:
            t.b.WriteString(strings.Join([]string{
                fmt.Sprintf("\t@%d", n),
                "D=A",
                "@"+varname,
                "A=D+M\n",
            }, "\n\t"))
        }
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
@256
D=A
@SP
M=D
@SYSINIT
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
    // push LCL    // save LCL of calling function
    @LCL
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
    // push THIS   // save THIS of calling function
    @THIS
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // push THAT   // save THAT of calling function
    @THAT
    D=M
    @SP
    M=M+1
    A=M-1
    M=D
    // ARG=SP-n-5  // reposition ARG (n=number of args)
    @SP
    D=M
    @R14    // NUMARGS
    D=D-M
    @5
    D=D-A
    @ARG
    M=D
    // LCL=SP      // reposition LCL
    @SP
    D=M
    @LCL
    M=D
    // goto f      // transfer control
    @R13    // FUNC
    A=M
    0;JMP
(SYSPUSHLCL)
    // func def here i.e. push 0 numlcl times
    @R14    // NUMLCL
    D=M
    @R15    // RET
    A=M
    D;JEQ
    @SP
    M=M+1
    A=M-1
    M=0
    @R14    // NUMLCL
    M=M-1
    @SYSPUSHLCL
    0;JMP
(SYSRETURN)
    // FRAME = LCL
    @LCL
    D=M
    @R14    // FRAME
    DM=D
    // RET = *(FRAME-5)
    @5
    A=D-A
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
    // SP = ARG+1
    @ARG
    D=M+1
    @SP
    M=D
    // THAT = *(FRAME-1)
    @R14
    AM=M-1
    D=M
    @THAT
    M=D
    // THIS = *(FRAME-2)
    @R14
    AM=M-1
    D=M
    @THIS
    M=D
    // ARG = *(FRAME-3)
    @R14
    AM=M-1
    D=M
    @ARG
    M=D
    // LCL = *(FRAME-4)
    @R14
    AM=M-1
    D=M
    @LCL
    M=D
    // goto RET
    @R15
    A=M
    0;JMP`
