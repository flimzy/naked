package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	if err := scanFile(os.Args[1]); err != nil {
		panic(err)
	}
}

func scanFile(filename string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return err
	}
	ast.Walk(&visitor{fset: fset}, f)
	return nil
}

func foo() (err error) {
	return
}

func bar() {}

func baz() error {
	return func() (err error) {
		return
	}()
}

type visitor struct {
	fset       *token.FileSet
	hasReturns []*bool
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		if l := len(v.hasReturns); l > 0 {
			v.hasReturns = v.hasReturns[:l-1]
		}
		return nil
	}
	switch t := node.(type) {
	case *ast.FuncDecl:
		hasReturns := t.Type.Results != nil
		v.hasReturns = append(v.hasReturns, &hasReturns)
	case *ast.ReturnStmt:
		if t.Results == nil && hasReturns(v.hasReturns) {
			p := v.fset.Position(t.Return)
			fmt.Printf("%s %d: naked return\n", p.Filename, p.Line)
		}
	default:
		v.hasReturns = append(v.hasReturns, nil)
	}
	return v
}

func hasReturns(r []*bool) bool {
	for i := len(r) - 1; i >= 0; i-- {
		if r[i] != nil {
			return *r[i]
		}
	}
	return false
}
