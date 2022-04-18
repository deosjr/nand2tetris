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
    files map[string]file

    // reused every file
    b *strings.Builder
    staticVars map[string]typeNum
    // reused every func
    args map[string]typeNum
    locals map[string]typeNum
}

type file struct {
    class *class
    funcs map[string]*ast.FuncDecl
    methods map[string]*ast.FuncDecl
    staticVars map[string]typeNum
}

type class struct {
    fields map[string]typeNum
}

type typeNum struct {
    typ string
    num int
}

// take a set of .jack files and return machine language code 
func Compile(filenames ...string) ([]uint16, error) {
    fset := token.NewFileSet()
    var parsed []*ast.File
    for _, filename := range filenames {
        data, err := os.ReadFile(filename)
        if err != nil {
            return nil, err
        }
        f, err := parser.ParseFile(fset, filename, "package main\n"+string(data), 0)
        if err != nil {
            return nil, err
        }
        split := strings.Split(filename, "/")
        filename = strings.Split(split[len(split)-1], ".")[0]
        f.Name = ast.NewIdent(filename)
        parsed = append(parsed, f)
    }
    var vmStrings []string
    c := &jackCompiler{files: map[string]file{}}
    for _, f := range parsed {
        if err := c.analyse(f); err != nil {
            return nil, err
        }
    }
    for _, f := range parsed {
        vm, err := c.translate(f)
        if err != nil {
            return nil, err
        }
        vmStrings = append(vmStrings, vm)
    }
    //fmt.Println(vmStrings)
    asm, err := vm2asm(filenames, vmStrings)
    if err != nil {
        return nil, err
    }
    return assembleFromString(asm)
}

// first pass so we know about types declared in other files
func (c *jackCompiler) analyse(f *ast.File) error {
    file := file{
        funcs: map[string]*ast.FuncDecl{},
        methods: map[string]*ast.FuncDecl{},
        staticVars: map[string]typeNum{},
    }
    for _, decl := range f.Decls {
        switch t := decl.(type) {
        case *ast.FuncDecl:
            if t.Recv == nil {
                file.funcs[t.Name.Name] = t
            } else {
                file.methods[t.Name.Name] = t
            }
        case *ast.GenDecl:
            switch spec := t.Specs[0].(type) {
            case *ast.ValueSpec:
                file.staticVars[spec.Names[0].Name] = typeNum{spec.Type.(*ast.Ident).Name, 0}
            case *ast.TypeSpec:
                if file.class != nil {
                    return fmt.Errorf("multiple type declarations in file")
                }
                file.class = typespec2class(spec)
            }
        default:
            return fmt.Errorf("unexpected %T", t)
        }
    }
    c.files[f.Name.Name] = file
    return nil
}

func typespec2class(t *ast.TypeSpec) *class {
    c := &class{fields: map[string]typeNum{}}
    for i, field := range t.Type.(*ast.StructType).Fields.List {
        c.fields[field.Names[0].Name] = typeNum{field.Type.(*ast.Ident).Name, i}
    }
    return c
}

func (c *jackCompiler) translate(f *ast.File) (string, error) {
    c.b = &strings.Builder{}
    c.staticVars = c.files[f.Name.Name].staticVars
    for _, fd := range c.files[f.Name.Name].funcs {
        if err := c.translateFuncDecl(fd); err != nil {
            return "", err
        }
    }
    for _, fd := range c.files[f.Name.Name].methods {
        if err := c.translateFuncDecl(fd); err != nil {
            return "", err
        }
    }
    return c.b.String(), nil
}

func (c *jackCompiler) translateFuncDecl(funcdecl *ast.FuncDecl) error {
    c.args = map[string]typeNum{}
    if funcdecl.Recv != nil {
        for _, field := range funcdecl.Recv.List {
            c.args[field.Names[0].Name] = typeNum{field.Type.(*ast.Ident).Name, 0}
        }
    }
    for _, field := range funcdecl.Type.Params.List {
        c.args[field.Names[0].Name] = typeNum{field.Type.(*ast.Ident).Name, len(c.args)}
    }
    c.locals = map[string]typeNum{}
    ast.Inspect(funcdecl.Body, func(n ast.Node) bool {
        switch t := n.(type) {
        case *ast.CallExpr, *ast.CompositeLit, *ast.SelectorExpr:
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
                c.locals[name] = typeNum{num: len(c.locals)}
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
    // if Lhs is a newly introduced local var, update its type based on Rhs
    typeRhs, err := c.typeOf(stmt.Rhs[0])
    if err != nil {
        return err
    }
    if ident, isIdent := stmt.Lhs[0].(*ast.Ident); isIdent {
        if tn, ok := c.locals[ident.Name]; ok && tn.typ == "" {
            c.locals[ident.Name] = typeNum{typeRhs, tn.num}
        }
    }
    if err := c.push(stmt.Rhs[0]); err != nil {
        return err
    }
    return c.pop(stmt.Lhs[0])
}

func (c *jackCompiler) translateCall(stmt *ast.CallExpr) error {
    if t, ok := stmt.Fun.(*ast.SelectorExpr); ok {
        name := t.X.(*ast.Ident).Name
        if _, isFile := c.files[name]; isFile {
            // calling a function by full path, ie mult.mult
            stmt.Fun = ast.NewIdent(name + "." + t.Sel.Name)
        } else {
            // calling a method on var 'name'
            typ, err := c.typeOf(t.X)
            if err != nil {
                return err
            }
            stmt.Fun = ast.NewIdent(typ + "." + t.Sel.Name)
            stmt.Args = append([]ast.Expr{t.X}, stmt.Args...)
        }
    }
    for _, arg := range stmt.Args {
        if err := c.push(arg); err != nil {
            return err
        }
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
            value = fmt.Sprintf("%d", value[1])
            if t.Value == "'\\n'" {
                value = "10"
            }
        }
        c.b.WriteString(fmt.Sprintf("\tpush constant %s\n", value))
        return nil
    case *ast.Ident:
        c.b.WriteString("\tpush ")
        return c.writeIdent(t)
    case *ast.CallExpr:
        return c.translateCall(t)
    case *ast.UnaryExpr:
        call := &ast.CallExpr{Fun:toFun(t.Op), Args:[]ast.Expr{t.X}}
        return c.translateCall(call)
    case *ast.BinaryExpr:
        call := &ast.CallExpr{Fun:toFun(t.Op), Args:[]ast.Expr{t.X, t.Y}}
        return c.translateCall(call)
    case *ast.IndexExpr:
        if err := c.prepareIndex(t); err != nil {
            return err
        }
        c.b.WriteString("\tpush that 0\n")
        return nil
    case *ast.CompositeLit:
        className := t.Type.(*ast.Ident).Name
        file, ok := c.files[className]
        if !ok || file.class == nil {
            return fmt.Errorf("push: type not found: %s", className)
        }
        if len(t.Elts) != 0 {
            return fmt.Errorf("push: struct init has to be empty")
        }
        c.b.WriteString(fmt.Sprintf("\tpush constant %d\n", len(file.class.fields)))
        c.b.WriteString("\tcall array.new 1\n")
        return nil
    case *ast.SelectorExpr:
        varname, ok := t.X.(*ast.Ident)
        if !ok {
            return fmt.Errorf("push: unexpected %T in selector", t.X)
        }
        typ, err := c.typeOf(t.X)
        if err != nil {
            return err
        }
        i := c.files[typ].class.fields[t.Sel.Name]
        indx := &ast.IndexExpr{X: varname, Index: &ast.BasicLit{Value:fmt.Sprintf("%d", i.num)}}
        return c.push(indx)
    }
    return fmt.Errorf("push: unexpected %T", expr)
}

// cannot recursively evaluate
func (c *jackCompiler) pop(expr ast.Expr) error {
    switch t := expr.(type) {
    case *ast.Ident:
        c.b.WriteString("\tpop ")
        return c.writeIdent(t)
    case *ast.IndexExpr:
        if err := c.prepareIndex(t); err != nil {
            return err
        }
        c.b.WriteString("\tpop that 0\n")
        return nil
    case *ast.SelectorExpr:
        varname, ok := t.X.(*ast.Ident)
        if !ok {
            return fmt.Errorf("pop: unexpected %T in selector", t.X)
        }
        typ, err := c.typeOf(t.X)
        if err != nil {
            return err
        }
        i := c.files[typ].class.fields[t.Sel.Name]
        indx := &ast.IndexExpr{X: varname, Index: &ast.BasicLit{Value:fmt.Sprintf("%d", i.num)}}
        return c.pop(indx)
    }
    return fmt.Errorf("pop: unexpected %T", expr)
}

func (c *jackCompiler) prepareIndex(t *ast.IndexExpr) error {
    if bl, ok := t.Index.(*ast.BasicLit); ok && bl.Value == "0" {
        c.b.WriteString("\tpush ")
        if err := c.writeIdent(t.X.(*ast.Ident)); err != nil {
            return err
        }
    } else {
        call := &ast.CallExpr{Fun:toFun(token.ADD), Args:[]ast.Expr{t.X, t.Index}}
        if err := c.translateCall(call); err != nil {
            return err
        }
    }
    c.b.WriteString("\tpop pointer 1\n")
    return nil
}

func (c *jackCompiler) writeIdent(ident *ast.Ident) error {
    name := ident.Name
    if name == "nil" {
        c.b.WriteString("static sys.nil\n")
        return nil
    }
    if _, ok := c.staticVars[name]; ok {
        c.b.WriteString(fmt.Sprintf("static %s\n", name))
        return nil
    }
    if tn, ok := c.args[name]; ok {
        c.b.WriteString(fmt.Sprintf("argument %d\n", tn.num))
        return nil
    }
    if tn, ok := c.locals[name]; ok {
        c.b.WriteString(fmt.Sprintf("local %d\n", tn.num))
        return nil
    }
    return fmt.Errorf("ident: not found %s", name)
}

func (c *jackCompiler) typeOf(expr ast.Expr) (string, error) {
    switch t := expr.(type) {
    case *ast.Ident:
        name := t.Name
        if tn, ok := c.staticVars[name]; ok {
            return tn.typ, nil
        }
        if tn, ok := c.args[name]; ok {
            return tn.typ, nil
        }
        if tn, ok := c.locals[name]; ok {
            return tn.typ, nil
        }
    case *ast.CallExpr:
        return "call", nil
    case *ast.BasicLit:
        if t.Kind == token.INT {
            return "int", nil
        }
        if t.Kind == token.CHAR {
            return "char", nil
        }
    case *ast.BinaryExpr:
        return "int", nil
    case *ast.CompositeLit:
        return t.Type.(*ast.Ident).Name, nil
    case *ast.SelectorExpr:
        return "any", nil
    }
    return fmt.Sprintf("notfound %T", expr), nil
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
