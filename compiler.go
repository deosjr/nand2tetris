package main

import (
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "strconv"
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
    contents := []string{}
    for _, filename := range filenames {
        data, err := os.ReadFile(filename)
        if err != nil {
            return nil, err
        }
        contents = append(contents, string(data))
    }
    return compile(filenames, contents)
}

func compile(filenames []string, contents []string) ([]uint16, error) {
    fset := token.NewFileSet()
    var parsed []*ast.File
    for i, filename := range filenames {
        data := contents[i]
        f, err := parser.ParseFile(fset, filename, "package main\n"+data, 0)
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
    //fmt.Println(asm)
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
            if name == "true" || name == "false" || name == "nil" {
                break
            }
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
        case *ast.DeclStmt:
            decl := t.Decl.(*ast.GenDecl)
            if decl.Tok != token.VAR {
                return fmt.Errorf("unexpected gendecl %s", decl.Tok)
            }
            for _, spec := range decl.Specs {
                v := spec.(*ast.ValueSpec)
                if tn, ok := c.locals[v.Names[0].Name]; ok && tn.typ == "" {
                    c.locals[v.Names[0].Name] = typeNum{v.Type.(*ast.Ident).Name, tn.num}
                } else {
                    return fmt.Errorf("unexpected var decl %#v", v)
                }
            }
        case *ast.AssignStmt:
            if err := c.translateAssign(t); err != nil {
                return err
            }
        case *ast.IfStmt:
            if err := c.translateIf(t); err != nil {
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
    var voidReturn bool
    if t, ok := stmt.Fun.(*ast.SelectorExpr); ok {
        name := t.X.(*ast.Ident).Name
        if file, isFile := c.files[name]; isFile {
            // calling a function by full path, ie mult.mult
            stmt.Fun = ast.NewIdent(name + "." + t.Sel.Name)
            f, ok := file.funcs[t.Sel.Name]
            if !ok {
                return fmt.Errorf("func not found: %s", stmt.Fun)
            }
            if f.Type.Results == nil {
                voidReturn = true
            }
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
    ident := stmt.Fun.(*ast.Ident)
    // HACK to turn leftshift into unary op...
    if ident.Name == "<<" {
        if len(stmt.Args) != 2 {
            return fmt.Errorf("expected 2 args for leftshift, got %v", len(stmt.Args))
        }
        if err := c.push(stmt.Args[0]); err != nil {
            return err
        }
        n, err := strconv.Atoi(stmt.Args[1].(*ast.BasicLit).Value)
        if err != nil {
            return err
        }
        c.b.WriteString(fmt.Sprintf("\tlshift%d\n", n))
        return nil
    }
    for _, arg := range stmt.Args {
        if err := c.push(arg); err != nil {
            return err
        }
    }
    switch ident.Name {
    case "print":
        // TODO: write only writes one 16-bit word
        c.b.WriteString("\twrite\n")
    case "read":
        // read only reads one 16-bit word
        c.b.WriteString("\tread\n")
    case "+":
        c.b.WriteString("\tadd\n")
    case "-":
        c.b.WriteString("\tsub\n")
    case "&":
        c.b.WriteString("\tand\n")
    case "|":
        c.b.WriteString("\tor\n")
    case "==":
        c.b.WriteString("\teq\n")
    case ">":
        c.b.WriteString("\tgt\n")
    case "<":
        c.b.WriteString("\tlt\n")
    default:
        c.b.WriteString(fmt.Sprintf("\tcall %s %d\n", stmt.Fun, len(stmt.Args)))
    }
    // if we just called a void function, consume the 0 return value
    if voidReturn {
        c.b.WriteString("\tpop temp 1\n") // dump into temp 1 for now
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
        if t.Kind == token.STRING {
            c.b.WriteString(fmt.Sprintf("\tpush constant %d\n\tcall string.new 1\n\tpop temp 0\n", len(t.Value)-2))
            for _, r := range t.Value[1:len(t.Value)-1] {
                // TODO this can at least be twice as fast now, if not much faster
                c.b.WriteString(fmt.Sprintf("\tpush temp 0\n\tpush constant %d\n\tcall string.appendChar 2\n", r))
            }
            c.b.WriteString("\tpush temp 0\n")
            return nil
        }
        value := t.Value
        if t.Kind == token.CHAR {
            value = fmt.Sprintf("%d", value[1])
            if t.Value == "'\\n'" {
                value = "10"
            }
        }
        // if value > 32767, this will result in an A instr
        // that gets interpreted as a C instr instead!
        n, err := strconv.ParseInt(value, 0, 0)
        if err != nil {
            return err
        }
        if n > 32767 {
            return fmt.Errorf("constant overflows A instr: %d", n)
        }
        c.b.WriteString(fmt.Sprintf("\tpush constant %d\n", n))
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
        switch tt := t.Type.(type) {
        case *ast.Ident:
            className := tt.Name
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
        case *ast.ArrayType:
            // TODO: only supporting []int{} declarations
            c.b.WriteString(fmt.Sprintf("\tpush constant %d\n", len(t.Elts)))
            c.b.WriteString("\tcall array.new 1\n")
            c.b.WriteString("\tpop temp 1\n")
            c.b.WriteString("\tpush temp 1\n")
            c.b.WriteString("\tpop pointer 1\n")
            for _, e := range t.Elts {
                bl := e.(*ast.BasicLit)
                if bl.Kind != token.INT {
                    return fmt.Errorf("only int array declarations supported")
                }
                c.b.WriteString(fmt.Sprintf("\tpush constant %s\n", bl.Value))
                c.b.WriteString("\tpop that 0\n")
                c.b.WriteString("\tpush pointer 1\n")
                c.b.WriteString("\tpush constant 1\n")
                c.b.WriteString("\tadd\n")
                c.b.WriteString("\tpop pointer 1\n")
            }
            c.b.WriteString("\tpush temp 1\n")
            return nil
        }
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
    if name == "nil" || name == "false" {
        c.b.WriteString("constant 0\n")
        return nil
    }
    if name == "true" {
        c.b.WriteString("constant 0\n\tnot\n")
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
        return "any", nil
    case *ast.CallExpr:
        return "call", nil
    case *ast.BasicLit:
        if t.Kind == token.INT {
            return "int", nil
        }
        if t.Kind == token.CHAR {
            return "char", nil
        }
        if t.Kind == token.STRING {
            return "string", nil
        }
    case *ast.BinaryExpr:
        return "int", nil
    case *ast.CompositeLit:
        switch tt := t.Type.(type) {
        case *ast.Ident:
            return tt.Name, nil
        case *ast.ArrayType:
            return "array", nil
        }
    case *ast.SelectorExpr:
        typ, err := c.typeOf(t.X)
        if err != nil {
            return "", err
        }
        field := c.files[typ].class.fields[t.Sel.Name]
        return field.typ, nil
    case *ast.IndexExpr:
        return "any", nil
    }
    return "", fmt.Errorf("notfound %T", expr)
}

func toFun(t token.Token) *ast.Ident {
    var str string
    switch t {
    case token.MUL:
        str = "math.mult"
    case token.QUO:
        str = "math.div"
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

// boolean=true means we invert by calling not
func inverseComp(op token.Token) (*ast.Ident, bool) {
    switch op {
    case token.EQL:
        return ast.NewIdent("=="), true
    case token.NEQ:
        return ast.NewIdent("=="), false
    case token.LEQ:
        return ast.NewIdent(">"), false
    case token.LSS:
        return ast.NewIdent("<"), true
    case token.GEQ:
        return ast.NewIdent("<"), false
    case token.GTR:
        return ast.NewIdent(">"), true
    }
    return nil, false
}

func (c *jackCompiler) translateIf(stmt *ast.IfStmt) error {
    endlabel := c.genLabel()
    var comp ast.Expr
    var invert bool
    var args []ast.Expr
    switch cond := stmt.Cond.(type) {
    case *ast.BinaryExpr:
        comp, invert = inverseComp(cond.Op)
        args = []ast.Expr{cond.X, cond.Y}
        call := &ast.CallExpr{Fun:comp, Args:args}
        if err := c.translateCall(call); err != nil {
            return err
        }
    case *ast.UnaryExpr:
        if cond.Op != token.NOT {
            return fmt.Errorf("if: only unary op allowed is 'not', got %s", cond.Op)
        }
        if err := c.push(cond.X); err != nil {
            return err
        }
        invert = false
    case *ast.Ident:
        if err := c.push(cond); err != nil {
            return err
        }
        invert = true
    case *ast.CallExpr:
        invert = true
        if err := c.translateCall(cond); err != nil {
            return err
        }
    default:
        return fmt.Errorf("if: unexpected %T", cond)
    }
    if invert {
        c.b.WriteString("\tnot\n")
    }
    c.b.WriteString(fmt.Sprintf("\tif-goto %s\n", endlabel))
    if err := c.translateBlock(stmt.Body); err != nil {
        return err
    }
    c.b.WriteString(fmt.Sprintf("label %s\n", endlabel))
    return nil
}

func (c *jackCompiler) translateFor(stmt *ast.ForStmt) error {
    // initialize the loop variable
    if stmt.Init != nil {
        if err := c.translateAssign(stmt.Init.(*ast.AssignStmt)); err != nil {
            return err
        }
    }
    looplabel := c.genLabel()
    endlabel := c.genLabel()
    c.b.WriteString(fmt.Sprintf("label %s\n", looplabel))
    // write comparison and jump out of loop
    if stmt.Cond != nil {
        cond := stmt.Cond.(*ast.BinaryExpr)
        comp, invert := inverseComp(cond.Op)
        call := &ast.CallExpr{Fun:comp, Args:[]ast.Expr{cond.X, cond.Y}}
        if err := c.translateCall(call); err != nil {
            return err
        }
        if invert {
            c.b.WriteString("\tnot\n")
        }
        c.b.WriteString(fmt.Sprintf("\tif-goto %s\n", endlabel))
    }
    // write the actual block within the for loop
    if err := c.translateBlock(stmt.Body); err != nil {
        return err
    }
    // increment or decrement loop counter
    if stmt.Post != nil {
        post := stmt.Post.(*ast.IncDecStmt)
        switch post.Tok {
        case token.INC:
            c.b.WriteString("\tplus1 ")
        case token.DEC:
            c.b.WriteString("\tminus1 ")
        default:
            return fmt.Errorf("unexpected %s", post.Tok)
        }
        ident := post.X.(*ast.Ident)
        c.writeIdent(ident)
        /*
        post := stmt.Post.(*ast.IncDecStmt)
        call := &ast.CallExpr{Fun:toFun(token.ADD), Args:[]ast.Expr{post.X, &ast.BasicLit{Value:"1", Kind:token.INT}}}
        if post.Tok == token.DEC {
            call.Fun = toFun(token.SUB)
        }
        assign := &ast.AssignStmt{Lhs:[]ast.Expr{post.X}, Rhs:[]ast.Expr{call}}
        if err := c.translateAssign(assign); err != nil {
            return err
        }
        */
    }
    c.b.WriteString(fmt.Sprintf("\tgoto %s\n", looplabel))
    if stmt.Cond != nil {
        c.b.WriteString(fmt.Sprintf("label %s\n", endlabel))
    }
    return nil
}
