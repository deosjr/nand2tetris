package main

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "strings"
)

type jackCompiler struct {
    generatedLabels int
    b *strings.Builder
    args map[string]int
    locals map[string]int
    staticVars map[string]string
}

// take a set of .jack files and return machine language code 
func Compile(filenames ...string) ([]uint16, error) {
    var vmStrings []string
    for _, filename := range filenames {
        data, err := os.ReadFile(filename)
        if err != nil {
            return nil, err
        }
        vm, err := jack2vm(filename, string(data))
        if err != nil {
            return nil, err
        }
        vmStrings = append(vmStrings, vm)
    }
    asm, err := vm2asm(filenames, vmStrings)
    if err != nil {
        return nil, err
    }
    return assembleFromString(asm)
}

func jack2vm(filename, in string) (string, error) {
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, filename, "package main\n"+in, 0)
    if err != nil {
        return "", err
    }
    c := &jackCompiler{
        b: &strings.Builder{},
    }
    if err := c.translate(f); err != nil {
        return "", err
    }
    return c.b.String(), nil
}

func (c *jackCompiler) translate(f *ast.File) error {
    c.staticVars = map[string]string{}
    var funcs []*ast.FuncDecl
    for _, decl := range f.Decls {
        switch t := decl.(type) {
        case *ast.FuncDecl:
            funcs = append(funcs, t)
        case *ast.GenDecl:
            vs := t.Specs[0].(*ast.ValueSpec)
            c.staticVars[vs.Names[0].Name] = vs.Type.(*ast.Ident).Name
        default:
            return fmt.Errorf("unexpected %T", t)
        }
    }
    for _, fd := range funcs {
        if err := c.translateFuncDecl(fd); err != nil {
            return err
        }
    }
    return nil
}

func (c *jackCompiler) translateFuncDecl(funcdecl *ast.FuncDecl) error {
    c.args = map[string]int{}
    for _, field := range funcdecl.Type.Params.List {
        c.args[field.Names[0].Name] = len(c.args)
    }
    c.locals = map[string]int{}
    ast.Inspect(funcdecl.Body, func(n ast.Node) bool {
        switch t := n.(type) {
        case *ast.CallExpr:
            return false
        case *ast.Ident:
            name := t.String()
            if _, ok := c.args[name]; ok {
                break
            }
            if _, ok := c.staticVars[name]; ok {
                break
            }
            if _, ok := c.locals[name]; !ok {
                c.locals[name] = len(c.locals)
            }
        }
        return true
    })
    fname := funcdecl.Name.Name
    c.b.WriteString(fmt.Sprintf("function %s %d\n", fname, len(c.locals)))
    return c.translateBlock(funcdecl.Body)
}

func (c *jackCompiler) translateBlock(stmt *ast.BlockStmt) error {
    for _, stmt := range stmt.List {
        switch t := stmt.(type) {
        case *ast.AssignStmt:
            if err := c.translateAssign(t); err != nil {
                return err
            }
        case *ast.ForStmt:
            if err := c.translateFor(t); err != nil {
                return err
            }
        case *ast.ReturnStmt:
            if err := c.translateReturn(t); err != nil {
                return err
            }
        case *ast.ExprStmt:
            call, ok := t.X.(*ast.CallExpr)
            if !ok {
                return fmt.Errorf("unexpected stmt %T", t.X)
            }
            if err := c.translateCall(call); err != nil {
                return err
            }
        default:
            return fmt.Errorf("unexpected stmt %T", t)
        }
    }
    return nil
}

func (c *jackCompiler) translateAssign(stmt *ast.AssignStmt) error {
    if len(stmt.Lhs) > 1 || len(stmt.Rhs) > 1 {
        return fmt.Errorf("unsupported multiple assign")
    }
    if err := c.push(stmt.Rhs[0]); err != nil {
        return err
    }
    return c.pop(stmt.Lhs[0])
}

func (c *jackCompiler) translateCall(stmt *ast.CallExpr) error {
    for _, arg := range stmt.Args {
        if err := c.push(arg); err != nil {
            return err
        }
    }
    if t, ok := stmt.Fun.(*ast.SelectorExpr); ok {
        stmt.Fun = ast.NewIdent(t.X.(*ast.Ident).Name + "." + t.Sel.Name)
    }
    ident := stmt.Fun.(*ast.Ident)
    switch ident.Name {
    case "print":
        // TODO: write only writes one 16-bit word
        c.b.WriteString("\twrite\n")
    case "+":
        c.b.WriteString("\tadd\n")
    case "-":
        c.b.WriteString("\tsub\n")
    case "=":
        c.b.WriteString("\teq\n")
    case ">":
        c.b.WriteString("\tgt\n")
    default:
        c.b.WriteString(fmt.Sprintf("\tcall %s %d\n", stmt.Fun, len(stmt.Args)))
    }
    return nil
}

func (c *jackCompiler) translateReturn(stmt *ast.ReturnStmt) error {
    if len(stmt.Results) == 0 {
        c.b.WriteString("\tpush constant 0\n\treturn\n")
        return nil
    }
    if len(stmt.Results) > 1 {
        return fmt.Errorf("return: unsupported multiple return")
    }
    if err := c.push(stmt.Results[0]); err != nil {
        return err
    }
    c.b.WriteString("\treturn\n")
    return nil
}

// can recursively evaluate
func (c *jackCompiler) push(expr ast.Expr) error {
    switch t := expr.(type) {
    case *ast.BasicLit:
        value := t.Value
        if t.Kind == token.CHAR {
            // todo convert to int representation of char
        }
        c.b.WriteString(fmt.Sprintf("\tpush constant %s\n", value))
    case *ast.Ident:
        ident, ok := expr.(*ast.Ident)
        if !ok {
            return fmt.Errorf("push: unexpected %T", expr)
        }
        c.b.WriteString("\tpush ")
        return c.writeIdent(ident)
    case *ast.CallExpr:
        return c.translateCall(t)
    case *ast.UnaryExpr:
        call := &ast.CallExpr{Fun:toFun(t.Op), Args:[]ast.Expr{t.X}}
        return c.translateCall(call)
    case *ast.BinaryExpr:
        call := &ast.CallExpr{Fun:toFun(t.Op), Args:[]ast.Expr{t.X, t.Y}}
        return c.translateCall(call)
    default:
        return fmt.Errorf("push: unexpected %T", t)
    }
    return nil
}

// cannot recursively evaluate
func (c *jackCompiler) pop(expr ast.Expr) error {
    c.b.WriteString("\tpop ")
    ident, ok := expr.(*ast.Ident)
    if !ok {
        return fmt.Errorf("pop: unexpected %T", expr)
    }
    return c.writeIdent(ident)
}

func (c *jackCompiler) writeIdent(ident *ast.Ident) error {
    name := ident.Name
    if _, ok := c.staticVars[name]; ok {
        c.b.WriteString(fmt.Sprintf("static %s\n", name))
        return nil
    }
    if i, ok := c.args[name]; ok {
        c.b.WriteString(fmt.Sprintf("argument %d\n", i))
        return nil
    }
    if i, ok := c.locals[name]; ok {
        c.b.WriteString(fmt.Sprintf("local %d\n", i))
        return nil
    }
    return fmt.Errorf("pop: not found %s", name)
}

func toFun(t token.Token) *ast.Ident {
    var str string
    switch t {
    case token.MUL:
        str = "mult.mult"
    default:
        str = t.String()
    }
    return ast.NewIdent(str)
}

func (c *jackCompiler) genLabel() string {
    s := ""
    for _, c := range fmt.Sprintf("%06d", c.generatedLabels) {
        s += string(c + 49)
    }
    c.generatedLabels++
    return "yy" + s
}

func inverseComp(op token.Token) *ast.Ident {
    switch op {
    case token.NEQ:
        return ast.NewIdent("=")
    case token.LEQ:
        return ast.NewIdent(">")
    }
    return nil
}

func (c *jackCompiler) translateFor(stmt *ast.ForStmt) error {
    // initialize the loop variable
    if err := c.translateAssign(stmt.Init.(*ast.AssignStmt)); err != nil {
        return err
    }
    looplabel := c.genLabel()
    endlabel := c.genLabel()
    c.b.WriteString(fmt.Sprintf("label %s\n", looplabel))
    // write comparison and jump out of loop
    cond := stmt.Cond.(*ast.BinaryExpr)
    comp := inverseComp(cond.Op)
    call := &ast.CallExpr{Fun:comp, Args:[]ast.Expr{cond.X, cond.Y}}
    if err := c.translateCall(call); err != nil {
        return err
    }
    c.b.WriteString(fmt.Sprintf("\tif-goto %s\n", endlabel))
    // write the actual block within the for loop
    if err := c.translateBlock(stmt.Body); err != nil {
        return err
    }
    // increment or decrement loop counter
    post := stmt.Post.(*ast.IncDecStmt)
    call = &ast.CallExpr{Fun:toFun(token.ADD), Args:[]ast.Expr{post.X, &ast.BasicLit{Value:"1", Kind:token.INT}}}
    if post.Tok == token.DEC {
        call.Fun = toFun(token.SUB)
    }
    assign := &ast.AssignStmt{Lhs:[]ast.Expr{post.X}, Rhs:[]ast.Expr{call}}
    if err := c.translateAssign(assign); err != nil {
        return err
    }
    c.b.WriteString(fmt.Sprintf("\tgoto %s\nlabel %s\n", looplabel, endlabel))
    return nil
}
